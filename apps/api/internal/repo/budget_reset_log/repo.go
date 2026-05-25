package budget_reset_log

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

// envTableName は BudgetResetLog テーブル名の環境変数名 (凍結 IF §10)。
const envTableName = "DDB_TABLE_BUDGET_RESET_LOG"

// ErrAlreadyLogged は (resetDate, userID) で既にログが存在する場合に返される。
var ErrAlreadyLogged = errors.New("budget_reset_log: already logged for (resetDate, userID)")

// DynamoDBAPI は Repository が依存する DynamoDB SDK の最小 interface。
type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

var defaultClient *dynamodb.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("budget_reset_log: failed to load AWS config: %v", err))
	}
	defaultClient = dynamodb.NewFromConfig(cfg)
}

// Repository は Get / Insert を提供する。
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

// Get は (resetDate, userID) で既存ログを取得する。未作成時は (nil, nil)。
func (r *Repository) Get(ctx context.Context, resetDate, userID string) (*Log, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"resetDate": &types.AttributeValueMemberS{Value: resetDate},
			"userId":    &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("budget_reset_log: getitem: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}
	return decode(out.Item)
}

// Insert は新規ログを書き込む。
//
// `(resetDate, userID)` で既存ログがある場合は ErrAlreadyLogged を返す (P-REL-03)。
func (r *Repository) Insert(ctx context.Context, log *Log) error {
	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item: map[string]types.AttributeValue{
			"resetDate":   &types.AttributeValueMemberS{Value: log.ResetDate},
			"userId":      &types.AttributeValueMemberS{Value: log.UserID},
			"prevBalance": &types.AttributeValueMemberN{Value: strconv.Itoa(log.PrevBalance)},
			"newBalance":  &types.AttributeValueMemberN{Value: strconv.Itoa(log.NewBalance)},
			"at":          &types.AttributeValueMemberS{Value: log.At.UTC().Format(time.RFC3339)},
		},
		ConditionExpression: aws.String("attribute_not_exists(resetDate) AND attribute_not_exists(userId)"),
	})
	if err != nil {
		var conflict *types.ConditionalCheckFailedException
		if errors.As(err, &conflict) {
			return ErrAlreadyLogged
		}
		return fmt.Errorf("budget_reset_log: insert: %w", err)
	}
	return nil
}

func decode(item map[string]types.AttributeValue) (*Log, error) {
	log := &Log{}
	if v, ok := item["resetDate"].(*types.AttributeValueMemberS); ok {
		log.ResetDate = v.Value
	}
	if v, ok := item["userId"].(*types.AttributeValueMemberS); ok {
		log.UserID = v.Value
	}
	if v, ok := item["prevBalance"].(*types.AttributeValueMemberN); ok {
		n, err := strconv.Atoi(v.Value)
		if err != nil {
			return nil, fmt.Errorf("decode prevBalance: %w", err)
		}
		log.PrevBalance = n
	}
	if v, ok := item["newBalance"].(*types.AttributeValueMemberN); ok {
		n, err := strconv.Atoi(v.Value)
		if err != nil {
			return nil, fmt.Errorf("decode newBalance: %w", err)
		}
		log.NewBalance = n
	}
	if v, ok := item["at"].(*types.AttributeValueMemberS); ok {
		t, err := time.Parse(time.RFC3339, v.Value)
		if err == nil {
			log.At = t
		}
	}
	return log, nil
}
