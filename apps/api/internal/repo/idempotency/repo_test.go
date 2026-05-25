package idempotency

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type stubClient struct {
	getItem    func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	putItem    func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	updateItem func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
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

func TestTryAcquire_SuccessOnFirstCall(t *testing.T) {
	stub := &stubClient{
		putItem: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
	}
	repo := NewRepositoryWithClient(stub, "test-table")
	acquired, existing, err := repo.TryAcquire(context.Background(), "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY", []byte(`{"amount":100}`), DefaultTTL)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !acquired {
		t.Fatal("expected acquired=true")
	}
	if existing != nil {
		t.Fatalf("expected nil existing, got %+v", existing)
	}
}

func TestTryAcquire_ConditionalFailedReturnsExisting(t *testing.T) {
	stub := &stubClient{
		putItem: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{}
		},
		getItem: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"key":      &types.AttributeValueMemberS{Value: "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY"},
					"payload":  &types.AttributeValueMemberB{Value: []byte(`{"amount":100}`)},
					"response": &types.AttributeValueMemberB{Value: []byte(`{"newBalance":29900}`)},
				},
			}, nil
		},
	}
	repo := NewRepositoryWithClient(stub, "test-table")
	acquired, existing, err := repo.TryAcquire(context.Background(), "user_a:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY", []byte(`{"amount":100}`), DefaultTTL)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if acquired {
		t.Fatal("expected acquired=false")
	}
	if existing == nil {
		t.Fatal("expected existing record")
	}
	if string(existing.Response) != `{"newBalance":29900}` {
		t.Fatalf("unexpected response: %s", string(existing.Response))
	}
}

func TestSaveResponse(t *testing.T) {
	called := false
	stub := &stubClient{
		updateItem: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			called = true
			return &dynamodb.UpdateItemOutput{}, nil
		},
	}
	repo := NewRepositoryWithClient(stub, "test-table")
	if err := repo.SaveResponse(context.Background(), "key", []byte("response")); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !called {
		t.Fatal("expected UpdateItem to be called")
	}
}

// TTL の具体的な経過確認は DynamoDB に委譲するため、ここでは単に SaveResponse が
// 呼ばれることのみ確認する。

var _ = time.Now // keep import
