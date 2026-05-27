// Package suggestion は GoroPay_Suggestion テーブルへのアクセス Repository を
// 提供する (Unit D 所有、LC-SUGGEST-04)。
//
// DynamoDB スキーマは凍結契約 §5.3 に従う:
//   - PK = suggestionId (ULID)
//   - TTL 属性 = expiresAt (30 分、NFRD-D05/D08)
//   - 主要属性 = userId / plan(JSON) / createdAt
//
// init() で SDK Client を Lambda INIT フェーズに初期化する (P-INIT-01)。
package suggestion

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Plan は保存される提案内容 (repo package 固有型、suggest package とは変換で連携し
// import cycle を避ける)。
type Plan struct {
	StoreName string `json:"storeName"`
	MenuName  string `json:"menuName"`
	Amount    int    `json:"amount"`
	Category  string `json:"category"`
}

// SuggestionRecord は GoroPay_Suggestion テーブルの 1 行 (凍結契約 §5.3)。
type SuggestionRecord struct {
	SuggestionID string
	UserID       string
	Plan         Plan
	CreatedAt    int64 // Unix epoch seconds
	ExpiresAt    int64 // Unix epoch seconds、TTL 属性 (CreatedAt + 1800)
}

// SuggestionStore は GoroPay_Suggestion への保存・取得 interface (LC-SUGGEST-04)。
//
// suggest package はこの interface に依存し、本番は Repository、テストは
// InmemoryStore を注入する (P-MOCK-01)。
type SuggestionStore interface {
	Save(ctx context.Context, rec *SuggestionRecord) error
	// Get は suggestionId で取得する。失効 (TTL 超過) / 不在時は nil, nil を返す
	// (BR-D10、Unit C BR-C10 透過フォールバックの前提)。
	Get(ctx context.Context, suggestionID string) (*SuggestionRecord, error)
}

// DynamoDBAPI は Repository 実装が依存する DynamoDB SDK の最小 interface。
type DynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

// envTableName は GoroPay_Suggestion テーブル名の環境変数名 (凍結契約 §10)。
const envTableName = "DDB_TABLE_SUGGESTION"

// defaultClient は Lambda INIT フェーズで初期化される DynamoDB Client (P-INIT-01)。
var defaultClient *dynamodb.Client

// init は Lambda INIT フェーズで DynamoDB Client を初期化する。失敗時は fail-fast。
func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("suggestion: failed to load AWS config: %v", err))
	}
	defaultClient = dynamodb.NewFromConfig(cfg)
}

// Repository は DynamoDB を使う SuggestionStore implementation。
type Repository struct {
	client    DynamoDBAPI
	tableName string
}

// NewRepository は production 用 Repository を返す (テーブル名は env から取得)。
func NewRepository() *Repository {
	return NewRepositoryWithClient(defaultClient, os.Getenv(envTableName))
}

// NewRepositoryWithClient はテスト用に SDK Client とテーブル名を注入できる constructor。
func NewRepositoryWithClient(client DynamoDBAPI, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

// Save は SuggestionRecord を DynamoDB に書き込む (TTL は expiresAt 属性)。
func (r *Repository) Save(ctx context.Context, rec *SuggestionRecord) error {
	planJSON, err := json.Marshal(rec.Plan)
	if err != nil {
		return fmt.Errorf("suggestion: marshal plan: %w", err)
	}
	item := map[string]types.AttributeValue{
		"suggestionId": &types.AttributeValueMemberS{Value: rec.SuggestionID},
		"userId":       &types.AttributeValueMemberS{Value: rec.UserID},
		"plan":         &types.AttributeValueMemberS{Value: string(planJSON)},
		"createdAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", rec.CreatedAt)},
		"expiresAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", rec.ExpiresAt)},
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("suggestion: putitem: %w", err)
	}
	return nil
}

// Get は suggestionId で 1 件取得する。不在時は nil, nil。
//
// TTL 失効アイテムは DynamoDB が遅延削除するため、削除前に取得されうる。
// expiresAt を読まずとも、呼出側 (SuggestService) は保存値をそのまま使う設計
// だが、念のため期限切れは nil 扱いにはしない (DynamoDB TTL 削除に委ねる、
// BR-D08)。本 MVP では取得できたものは有効とみなす。
func (r *Repository) Get(ctx context.Context, suggestionID string) (*SuggestionRecord, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"suggestionId": &types.AttributeValueMemberS{Value: suggestionID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("suggestion: getitem: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}
	return fromDynamoItem(out.Item)
}

// fromDynamoItem は DynamoDB 属性 map から SuggestionRecord を復元する。
func fromDynamoItem(item map[string]types.AttributeValue) (*SuggestionRecord, error) {
	rec := &SuggestionRecord{}
	if v, ok := item["suggestionId"].(*types.AttributeValueMemberS); ok {
		rec.SuggestionID = v.Value
	}
	if v, ok := item["userId"].(*types.AttributeValueMemberS); ok {
		rec.UserID = v.Value
	}
	if v, ok := item["plan"].(*types.AttributeValueMemberS); ok {
		if err := json.Unmarshal([]byte(v.Value), &rec.Plan); err != nil {
			return nil, fmt.Errorf("suggestion: unmarshal plan: %w", err)
		}
	}
	if v, ok := item["createdAt"].(*types.AttributeValueMemberN); ok {
		fmt.Sscanf(v.Value, "%d", &rec.CreatedAt)
	}
	if v, ok := item["expiresAt"].(*types.AttributeValueMemberN); ok {
		fmt.Sscanf(v.Value, "%d", &rec.ExpiresAt)
	}
	return rec, nil
}
