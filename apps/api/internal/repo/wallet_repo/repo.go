package wallet_repo

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

// envTableName は Wallet テーブル名の環境変数名 (凍結 IF §10)。
const envTableName = "DDB_TABLE_WALLET"

// DynamoDBAPI は Repository が依存する DynamoDB SDK の最小 interface (テスト注入用)。
type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// defaultClient は Lambda INIT フェーズで初期化される (P-INIT-01)。
var defaultClient *dynamodb.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("wallet_repo: failed to load AWS config: %v", err))
	}
	defaultClient = dynamodb.NewFromConfig(cfg)
}

// Repository は WalletReader + 内部 Write 操作を実装する。
type Repository struct {
	client    DynamoDBAPI
	tableName string
}

// NewRepository は production 用 Repository を返す (env から table name を取得)。
func NewRepository() *Repository {
	return NewRepositoryWithClient(defaultClient, os.Getenv(envTableName))
}

// NewRepositoryWithClient はテスト用 constructor。
func NewRepositoryWithClient(client DynamoDBAPI, tableName string) *Repository {
	return &Repository{client: client, tableName: tableName}
}

// 確認: WalletReader を実装する。
var _ WalletReader = (*Repository)(nil)

// Get は userID に対応する Wallet を ConsistentRead で取得する (PR-B-05 / P-PERF-01)。
//
// 未作成時は (nil, nil) を返す。
func (r *Repository) Get(ctx context.Context, userID string) (*WalletRecord, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(r.tableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("wallet_repo: getitem: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}
	return decodeWallet(out.Item)
}

// Create は新規 Wallet を作成する (UC-B-01 初回パス)。
//
// 既存ユーザに対する Create は ConditionExpression で弾く (重複作成防止)。
func (r *Repository) Create(ctx context.Context, userID string, balance int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item: map[string]types.AttributeValue{
			"userId":    &types.AttributeValueMemberS{Value: userID},
			"balance":   &types.AttributeValueMemberN{Value: strconv.Itoa(balance)},
			"updatedAt": &types.AttributeValueMemberS{Value: now},
		},
		ConditionExpression: aws.String("attribute_not_exists(userId)"),
	})
	if err != nil {
		var conflict *types.ConditionalCheckFailedException
		if errors.As(err, &conflict) {
			return fmt.Errorf("wallet_repo: wallet already exists for user %s: %w", userID, err)
		}
		return fmt.Errorf("wallet_repo: create: %w", err)
	}
	return nil
}

// DeductConditional は ConditionExpression: balance >= :amount で残高を減算する (P-REL-01)。
//
// 残高不足時は wallet.ErrInsufficientBalance を即返却する (リトライなし)。
func (r *Repository) DeductConditional(ctx context.Context, userID string, amount int) (newBalance int, err error) {
	now := time.Now().UTC().Format(time.RFC3339)
	out, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression:    aws.String("SET balance = balance - :amount, updatedAt = :now"),
		ConditionExpression: aws.String("attribute_exists(userId) AND balance >= :amount"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":amount": &types.AttributeValueMemberN{Value: strconv.Itoa(amount)},
			":now":    &types.AttributeValueMemberS{Value: now},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		var conflict *types.ConditionalCheckFailedException
		if errors.As(err, &conflict) {
			return 0, ErrInsufficientBalance
		}
		return 0, fmt.Errorf("wallet_repo: deduct: %w", err)
	}
	rec, decErr := decodeWallet(out.Attributes)
	if decErr != nil {
		return 0, decErr
	}
	return rec.Balance, nil
}

// UpdateBalance は balance に delta を加算する (UC-B-02 増額パス)。
//
// 減額方向 (delta < 0) でも呼び出し可能だが、unconditional のため
// 残高負化を防ぐ責務は呼び出し側 (wallet.Service) が持つ。
func (r *Repository) UpdateBalance(ctx context.Context, userID string, delta int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression: aws.String("SET balance = balance + :delta, updatedAt = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":delta": &types.AttributeValueMemberN{Value: strconv.Itoa(delta)},
			":now":   &types.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		return fmt.Errorf("wallet_repo: update_balance: %w", err)
	}
	return nil
}

// SetBalance は balance を絶対値で上書きする (UC-B-02 減額打ち切り / UC-B-05 ResetTo 兼用)。
func (r *Repository) SetBalance(ctx context.Context, userID string, balance int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression: aws.String("SET balance = :balance, updatedAt = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":balance": &types.AttributeValueMemberN{Value: strconv.Itoa(balance)},
			":now":     &types.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		return fmt.Errorf("wallet_repo: set_balance: %w", err)
	}
	return nil
}

// ListAllUserIDs は Wallet テーブル全件を Scan して userID を返す (UC-B-05 ResetAll)。
//
// ハッカソン規模 (デモ最大 50 ユーザ) では Scan で十分。
func (r *Repository) ListAllUserIDs(ctx context.Context) ([]string, error) {
	var userIDs []string
	var lastKey map[string]types.AttributeValue
	for {
		out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:            aws.String(r.tableName),
			ProjectionExpression: aws.String("userId"),
			ExclusiveStartKey:    lastKey,
		})
		if err != nil {
			return nil, fmt.Errorf("wallet_repo: scan: %w", err)
		}
		for _, item := range out.Items {
			if v, ok := item["userId"].(*types.AttributeValueMemberS); ok {
				userIDs = append(userIDs, v.Value)
			}
		}
		if out.LastEvaluatedKey == nil {
			break
		}
		lastKey = out.LastEvaluatedKey
	}
	return userIDs, nil
}

func decodeWallet(item map[string]types.AttributeValue) (*WalletRecord, error) {
	rec := &WalletRecord{}
	if v, ok := item["userId"].(*types.AttributeValueMemberS); ok {
		rec.UserID = v.Value
	}
	if v, ok := item["balance"].(*types.AttributeValueMemberN); ok {
		n, err := strconv.Atoi(v.Value)
		if err != nil {
			return nil, fmt.Errorf("wallet_repo: decode balance: %w", err)
		}
		rec.Balance = n
	}
	if v, ok := item["updatedAt"].(*types.AttributeValueMemberS); ok {
		t, err := time.Parse(time.RFC3339, v.Value)
		if err == nil {
			rec.UpdatedAt = t
		}
	}
	return rec, nil
}
