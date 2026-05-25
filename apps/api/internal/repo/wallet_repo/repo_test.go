package wallet_repo

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type stubClient struct {
	getItem    func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	putItem    func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	updateItem func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	scan       func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (s *stubClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return s.getItem(ctx, params, optFns...)
}

func (s *stubClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return s.putItem(ctx, params, optFns...)
}

func (s *stubClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return s.updateItem(ctx, params, optFns...)
}

func (s *stubClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return s.scan(ctx, params, optFns...)
}

func TestDeductConditional_ConditionalCheckFailedMapsToErrInsufficientBalance(t *testing.T) {
	stub := &stubClient{
		updateItem: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{}
		},
	}
	repo := NewRepositoryWithClient(stub, "test-table")
	_, err := repo.DeductConditional(context.Background(), "user_a", 1000)
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestGet_ReturnsNilNilWhenNotFound(t *testing.T) {
	stub := &stubClient{
		getItem: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{Item: nil}, nil
		},
	}
	repo := NewRepositoryWithClient(stub, "test-table")
	got, err := repo.Get(context.Background(), "user_a")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestDeductConditional_Success(t *testing.T) {
	stub := &stubClient{
		updateItem: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			return &dynamodb.UpdateItemOutput{
				Attributes: map[string]types.AttributeValue{
					"userId":  &types.AttributeValueMemberS{Value: "user_a"},
					"balance": &types.AttributeValueMemberN{Value: "29000"},
				},
			}, nil
		},
	}
	repo := NewRepositoryWithClient(stub, "test-table")
	bal, err := repo.DeductConditional(context.Background(), "user_a", 1000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if bal != 29000 {
		t.Fatalf("expected 29000, got %d", bal)
	}
}
