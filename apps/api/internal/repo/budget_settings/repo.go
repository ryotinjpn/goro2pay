package budget_settings

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// envTableName は BudgetSettings テーブル名の環境変数名 (凍結 IF §10)。
const envTableName = "DDB_TABLE_BUDGET_SETTINGS"

// DynamoDBAPI は Repository が依存する DynamoDB SDK の最小 interface。
type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

var defaultClient *dynamodb.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("budget_settings: failed to load AWS config: %v", err))
	}
	defaultClient = dynamodb.NewFromConfig(cfg)
}

// Repository は BudgetSettingsReader + BudgetSettingsWriter を実装する。
type Repository struct {
	client    DynamoDBAPI
	tableName string
}

func NewRepository() *Repository {
	return NewRepositoryWithClient(defaultClient, os.Getenv(envTableName))
}

func NewRepositoryWithClient(client DynamoDBAPI, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

var (
	_ BudgetSettingsReader = (*Repository)(nil)
	_ BudgetSettingsWriter = (*Repository)(nil)
)

// Get は userID に対応する BudgetSettings を取得する。未作成時は (nil, nil)。
func (r *Repository) Get(ctx context.Context, userID string) (*BudgetSettings, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("budget_settings: getitem: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}
	return decode(out.Item)
}

// Set は BudgetSettings を Upsert する (UC-B-01/02 で内部的に呼ばれる + Unit E Writer)。
//
// UpdateItem ベース実装 (Code Review Important 4 修正):
//   - 既存 item の他属性 (`createdAt`、Unit E 実装後の `raiseHistory` 等) を保護する
//   - `monthlyBudget` / `effectiveFrom` / `updatedAt` のみ更新
//   - `createdAt` は item が存在しない場合のみセット (`if_not_exists`)
//
// 凍結 IF §6.3 で Unit E `BudgetRaiseService.Accept` が同 Writer を呼ぶため、
// PutItem 全上書きだと将来 raiseHistory が消失する設計上の問題があった。
func (r *Repository) Set(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression: aws.String("SET monthlyBudget = :mb, effectiveFrom = :ef, updatedAt = :u, createdAt = if_not_exists(createdAt, :u)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":mb": &types.AttributeValueMemberN{Value: strconv.Itoa(monthlyBudget)},
			":ef": &types.AttributeValueMemberS{Value: effectiveFrom.UTC().Format(time.RFC3339)},
			":u":  &types.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		return fmt.Errorf("budget_settings: updateitem: %w", err)
	}
	return nil
}

func decode(item map[string]types.AttributeValue) (*BudgetSettings, error) {
	bs := &BudgetSettings{}
	if v, ok := item["userId"].(*types.AttributeValueMemberS); ok {
		bs.UserID = v.Value
	}
	if v, ok := item["monthlyBudget"].(*types.AttributeValueMemberN); ok {
		n, err := strconv.Atoi(v.Value)
		if err != nil {
			return nil, fmt.Errorf("budget_settings: decode monthlyBudget: %w", err)
		}
		bs.MonthlyBudget = n
	}
	if v, ok := item["effectiveFrom"].(*types.AttributeValueMemberS); ok {
		t, err := time.Parse(time.RFC3339, v.Value)
		if err == nil {
			bs.EffectiveFrom = t
		}
	}
	return bs, nil
}
