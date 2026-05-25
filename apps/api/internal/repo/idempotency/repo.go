package idempotency

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// envTableName は IdempotencyKeys テーブル名の環境変数名 (凍結 IF §10)。
const envTableName = "DDB_TABLE_IDEMPOTENCY"

// DefaultTTL は IdempotencyRecord の保持期間 (PR-B-03)。
const DefaultTTL = 24 * time.Hour

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
		panic(fmt.Sprintf("idempotency: failed to load AWS config: %v", err))
	}
	defaultClient = dynamodb.NewFromConfig(cfg)
}

// Repository は IdempotencyRecord の TryAcquire / SaveResponse を提供する。
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

// TryAcquire は冪等性キーを payload 付きで取得試行する。
//
// - 初回 (acquired=true): existing は nil、後続で SaveResponse を呼ぶ責務がある
// - 既存 (acquired=false): existing に既存レコードを返す。呼び出し側で payload hash 比較 +
//   保存済み response の再生を行う (DR-B-03)
func (r *Repository) TryAcquire(ctx context.Context, key string, payload []byte, ttl time.Duration) (acquired bool, existing *IdempotencyRecord, err error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	_, putErr := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item: map[string]types.AttributeValue{
			"key":       &types.AttributeValueMemberS{Value: key},
			"payload":   &types.AttributeValueMemberB{Value: payload},
			"createdAt": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"expiresAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(expiresAt.Unix(), 10)},
		},
		ConditionExpression: aws.String("attribute_not_exists(#k)"),
		ExpressionAttributeNames: map[string]string{
			"#k": "key",
		},
	})
	if putErr == nil {
		return true, nil, nil
	}
	var conflict *types.ConditionalCheckFailedException
	if errors.As(putErr, &conflict) {
		// 既存レコードを取得して返す
		existingRec, getErr := r.get(ctx, key)
		if getErr != nil {
			return false, nil, fmt.Errorf("idempotency: get existing: %w", getErr)
		}
		return false, existingRec, nil
	}
	return false, nil, fmt.Errorf("idempotency: tryacquire: %w", putErr)
}

// SaveResponse は初回処理結果を response 属性に保存する (PR-B-04)。
//
// 失敗結果も保存する: 同一キーの再送には保存通り返却することで race condition で
// 失敗結果が再現される。
func (r *Repository) SaveResponse(ctx context.Context, key string, response []byte) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{Value: key},
		},
		UpdateExpression: aws.String("SET response = :response"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":response": &types.AttributeValueMemberB{Value: response},
		},
	})
	if err != nil {
		return fmt.Errorf("idempotency: save_response: %w", err)
	}
	return nil
}

func (r *Repository) get(ctx context.Context, key string) (*IdempotencyRecord, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, nil
	}
	rec := &IdempotencyRecord{}
	if v, ok := out.Item["key"].(*types.AttributeValueMemberS); ok {
		rec.Key = v.Value
	}
	if v, ok := out.Item["payload"].(*types.AttributeValueMemberB); ok {
		rec.Payload = v.Value
	}
	if v, ok := out.Item["response"].(*types.AttributeValueMemberB); ok {
		rec.Response = v.Value
	}
	if v, ok := out.Item["createdAt"].(*types.AttributeValueMemberS); ok {
		t, err := time.Parse(time.RFC3339, v.Value)
		if err == nil {
			rec.CreatedAt = t
		}
	}
	if v, ok := out.Item["expiresAt"].(*types.AttributeValueMemberN); ok {
		n, err := strconv.ParseInt(v.Value, 10, 64)
		if err == nil {
			rec.ExpiresAt = time.Unix(n, 0)
		}
	}
	return rec, nil
}
