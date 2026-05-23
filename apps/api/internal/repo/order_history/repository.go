// Package orderhistory は OrderHistory テーブルへのアクセス Repository を提供する。
//
// DynamoDB スキーマは凍結契約 §3.2 に従う:
//   - PK = USER#{userID}
//   - SK = ORDER#{orderedAt}#{orderID}
//   - TTL 属性 = expiresAt (90 日)
//
// init() で SDK Client を Lambda INIT フェーズに初期化し (P-INIT-01)、
// INVOKE フェーズの応答時間 (NFRC-C18 600ms) を確保する。
package orderhistory

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// OrderRecord は OrderHistory テーブルの 1 行に対応する。
//
// 凍結契約 §4.1 の公開フィールド + Unit C 内部追加属性 (idempotencyKey /
// dayOfWeek / source / expiresAt)。
type OrderRecord struct {
	OrderID        string
	UserID         string
	Category       string
	StoreName      string
	MenuName       string
	Amount         int
	OrderedAt      string // RFC3339 string
	IdempotencyKey string
	DayOfWeek      string
	Source         string // "bedrock" | "fallback_history" | "fallback_default"
	ExpiresAt      int64  // Unix epoch seconds、TTL 属性
}

// OrderHistoryRepository は OrderHistory への CRUD interface。
type OrderHistoryRepository interface {
	Insert(ctx context.Context, record *OrderRecord) error
	GetItem(ctx context.Context, userID, orderID, orderedAt string) (*OrderRecord, error)
	Query(ctx context.Context, userID string, limit int) ([]*OrderRecord, error)
}

// DynamoDBAPI は Repository 実装が依存する DynamoDB SDK の最小 interface。
//
// テスト時は mock を注入する (NFRC-C16 / P-MOCK-01)。
type DynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// defaultClient は Lambda INIT フェーズで初期化される DynamoDB Client (P-INIT-01)。
var defaultClient *dynamodb.Client

const (
	// envTableName は OrderHistory テーブル名の環境変数名。
	envTableName = "ORDER_HISTORY_TABLE_NAME"

	// defaultLimit / maxLimit は Query の件数制御 (FD Q-11=A)。
	defaultLimit = 20
	maxLimit     = 100
)

// init は Lambda INIT フェーズで DynamoDB Client を初期化する。
//
// 失敗時は panic で fail-fast (Lambda 起動失敗 → CloudWatch 即検知)。
func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("orderhistory: failed to load AWS config: %v", err))
	}
	defaultClient = dynamodb.NewFromConfig(cfg)
}

// Repository は DynamoDB を使う OrderHistoryRepository implementation。
type Repository struct {
	client    DynamoDBAPI
	tableName string
}

// NewRepository は production 用 Repository を返す。
//
// テーブル名は ORDER_HISTORY_TABLE_NAME 環境変数から取得する
// (Terraform output 経由で Lambda に注入される)。
func NewRepository() *Repository {
	return NewRepositoryWithClient(defaultClient, os.Getenv(envTableName))
}

// NewRepositoryWithClient はテスト用に SDK Client とテーブル名を注入できる constructor。
func NewRepositoryWithClient(client DynamoDBAPI, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

// pk は PK 値 USER#{userID} を返す。
func pk(userID string) string {
	return "USER#" + userID
}

// sk は SK 値 ORDER#{orderedAt}#{orderID} を返す。
func sk(orderedAt, orderID string) string {
	return "ORDER#" + orderedAt + "#" + orderID
}

// dynamoItem は OrderRecord を DynamoDB 属性 map に変換する。
//
// DynamoDB 属性は camelCase を使う (Unit C 内部規約、Q-N6/凍結契約整合)。
func dynamoItem(r *OrderRecord) (map[string]types.AttributeValue, error) {
	item := map[string]types.AttributeValue{
		"PK":             &types.AttributeValueMemberS{Value: pk(r.UserID)},
		"SK":             &types.AttributeValueMemberS{Value: sk(r.OrderedAt, r.OrderID)},
		"orderId":        &types.AttributeValueMemberS{Value: r.OrderID},
		"userId":         &types.AttributeValueMemberS{Value: r.UserID},
		"category":       &types.AttributeValueMemberS{Value: r.Category},
		"storeName":      &types.AttributeValueMemberS{Value: r.StoreName},
		"menuName":       &types.AttributeValueMemberS{Value: r.MenuName},
		"amount":         &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", r.Amount)},
		"orderedAt":      &types.AttributeValueMemberS{Value: r.OrderedAt},
		"idempotencyKey": &types.AttributeValueMemberS{Value: r.IdempotencyKey},
		"dayOfWeek":      &types.AttributeValueMemberS{Value: r.DayOfWeek},
		"source":         &types.AttributeValueMemberS{Value: r.Source},
		"expiresAt":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", r.ExpiresAt)},
	}
	return item, nil
}

// fromDynamoItem は DynamoDB 属性 map から OrderRecord を復元する。
func fromDynamoItem(item map[string]types.AttributeValue) (*OrderRecord, error) {
	type row struct {
		OrderID        string `dynamodbav:"orderId"`
		UserID         string `dynamodbav:"userId"`
		Category       string `dynamodbav:"category"`
		StoreName      string `dynamodbav:"storeName"`
		MenuName       string `dynamodbav:"menuName"`
		Amount         int    `dynamodbav:"amount"`
		OrderedAt      string `dynamodbav:"orderedAt"`
		IdempotencyKey string `dynamodbav:"idempotencyKey"`
		DayOfWeek      string `dynamodbav:"dayOfWeek"`
		Source         string `dynamodbav:"source"`
		ExpiresAt      int64  `dynamodbav:"expiresAt"`
	}
	var r row
	if err := attributevalue.UnmarshalMap(item, &r); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &OrderRecord{
		OrderID:        r.OrderID,
		UserID:         r.UserID,
		Category:       r.Category,
		StoreName:      r.StoreName,
		MenuName:       r.MenuName,
		Amount:         r.Amount,
		OrderedAt:      r.OrderedAt,
		IdempotencyKey: r.IdempotencyKey,
		DayOfWeek:      r.DayOfWeek,
		Source:         r.Source,
		ExpiresAt:      r.ExpiresAt,
	}, nil
}

// Insert は OrderRecord を DynamoDB に書き込む。
//
// PK + SK の組み合わせは ULID + RFC3339 タイムスタンプで実質的に衝突しないが、
// DynamoDB の attribute_not_exists ConditionExpression で重複検知も行う。
func (r *Repository) Insert(ctx context.Context, record *OrderRecord) error {
	item, err := dynamoItem(record)
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	})
	if err != nil {
		var conflict *types.ConditionalCheckFailedException
		if errors.As(err, &conflict) {
			return fmt.Errorf("orderhistory: duplicate insert: %w", err)
		}
		return fmt.Errorf("orderhistory: putitem: %w", err)
	}
	return nil
}

// GetItem は (userID, orderID, orderedAt) で 1 件取得する。
//
// 凍結契約 §3.2 の SK は ORDER#{orderedAt}#{orderID} なので、両方を引数に取る。
// 見つからない場合は nil, nil を返す (sentinel error は使わない)。
func (r *Repository) GetItem(ctx context.Context, userID, orderID, orderedAt string) (*OrderRecord, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk(userID)},
			"SK": &types.AttributeValueMemberS{Value: sk(orderedAt, orderID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("orderhistory: getitem: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}
	return fromDynamoItem(out.Item)
}

// Query は userID の履歴を降順 (新しい順) で limit 件返す。
//
// FD Q-11=A: デフォルト 20、最大 100、TTL 90 日内 (TTL 期限切れアイテムは
// DynamoDB が自動削除、Query 結果に含まれない)。
func (r *Repository) Query(ctx context.Context, userID string, limit int) ([]*OrderRecord, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk(userID)},
		},
		ScanIndexForward: aws.Bool(false), // 降順 (orderedAt 新しい順)
		Limit:            aws.Int32(int32(limit)),
	})
	if err != nil {
		return nil, fmt.Errorf("orderhistory: query: %w", err)
	}
	records := make([]*OrderRecord, 0, len(out.Items))
	for _, item := range out.Items {
		r, err := fromDynamoItem(item)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}
