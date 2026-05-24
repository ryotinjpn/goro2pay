package orderhistory

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDynamoClient は DynamoDBAPI を satisfy する手書き mock。
type fakeDynamoClient struct {
	putItemFn func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	getItemFn func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	queryFn   func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

func (f *fakeDynamoClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return f.putItemFn(ctx, params, optFns...)
}

func (f *fakeDynamoClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return f.getItemFn(ctx, params, optFns...)
}

func (f *fakeDynamoClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return f.queryFn(ctx, params, optFns...)
}

func sampleRecord() *OrderRecord {
	return &OrderRecord{
		OrderID:        "01HZ123",
		UserID:         "user-1",
		Category:       "food",
		StoreName:      "ゴロゴロ食堂",
		MenuName:       "おまかせ定食",
		Amount:         1000,
		OrderedAt:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: "01HZIDEM",
		DayOfWeek:      "Mon",
		Source:         "bedrock",
		ExpiresAt:      time.Now().Add(90 * 24 * time.Hour).Unix(),
	}
}

func TestRepository_Insert_Success(t *testing.T) {
	called := false
	fake := &fakeDynamoClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			called = true
			assert.Equal(t, "test-table", *params.TableName)
			assert.Contains(t, *params.ConditionExpression, "attribute_not_exists")
			return &dynamodb.PutItemOutput{}, nil
		},
	}
	repo := NewRepositoryWithClient(fake, "test-table")
	err := repo.Insert(context.Background(), sampleRecord())
	require.NoError(t, err)
	assert.True(t, called)
}

func TestRepository_Insert_Duplicate(t *testing.T) {
	fake := &fakeDynamoClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{Message: aws.String("duplicate")}
		},
	}
	repo := NewRepositoryWithClient(fake, "test-table")
	err := repo.Insert(context.Background(), sampleRecord())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestRepository_GetItem_Found(t *testing.T) {
	fake := &fakeDynamoClient{
		getItemFn: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"orderId":   &types.AttributeValueMemberS{Value: "01HZ"},
					"userId":    &types.AttributeValueMemberS{Value: "user-1"},
					"category":  &types.AttributeValueMemberS{Value: "food"},
					"storeName": &types.AttributeValueMemberS{Value: "ゴロゴロ食堂"},
					"menuName":  &types.AttributeValueMemberS{Value: "おまかせ定食"},
					"amount":    &types.AttributeValueMemberN{Value: "1000"},
					"orderedAt": &types.AttributeValueMemberS{Value: "2026-05-24T00:00:00Z"},
				},
			}, nil
		},
	}
	repo := NewRepositoryWithClient(fake, "t")
	rec, err := repo.GetItem(context.Background(), "user-1", "01HZ", "2026-05-24T00:00:00Z")
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.Equal(t, "01HZ", rec.OrderID)
	assert.Equal(t, 1000, rec.Amount)
}

func TestRepository_GetItem_NotFound(t *testing.T) {
	fake := &fakeDynamoClient{
		getItemFn: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{}, nil
		},
	}
	repo := NewRepositoryWithClient(fake, "t")
	rec, err := repo.GetItem(context.Background(), "u", "o", "t")
	require.NoError(t, err)
	assert.Nil(t, rec)
}

func TestRepository_Query_DescOrderAndLimit(t *testing.T) {
	fake := &fakeDynamoClient{
		queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			require.NotNil(t, params.ScanIndexForward)
			assert.False(t, *params.ScanIndexForward, "must be descending")
			require.NotNil(t, params.Limit)
			assert.Equal(t, int32(10), *params.Limit)
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	repo := NewRepositoryWithClient(fake, "t")
	_, err := repo.Query(context.Background(), "user-1", 10)
	require.NoError(t, err)
}

func TestRepository_Query_LimitClamping(t *testing.T) {
	tests := []struct {
		input, expected int
	}{
		{0, defaultLimit},
		{-5, defaultLimit},
		{50, 50},
		{1000, maxLimit},
	}
	for _, tt := range tests {
		called := int32(0)
		fake := &fakeDynamoClient{
			queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				called = *params.Limit
				return &dynamodb.QueryOutput{}, nil
			},
		}
		repo := NewRepositoryWithClient(fake, "t")
		_, err := repo.Query(context.Background(), "u", tt.input)
		require.NoError(t, err)
		assert.Equal(t, int32(tt.expected), called)
	}
}

func TestInmemoryRepository_InsertAndQuery(t *testing.T) {
	repo := NewInmemoryRepository()
	for i, t1 := range []string{"2026-05-22T00:00:00Z", "2026-05-23T00:00:00Z", "2026-05-24T00:00:00Z"} {
		_ = i
		err := repo.Insert(context.Background(), &OrderRecord{
			OrderID:   "id-" + t1,
			UserID:    "user-1",
			OrderedAt: t1,
		})
		require.NoError(t, err)
	}
	got, err := repo.Query(context.Background(), "user-1", 10)
	require.NoError(t, err)
	require.Len(t, got, 3)
	// 降順
	assert.Equal(t, "2026-05-24T00:00:00Z", got[0].OrderedAt)
	assert.Equal(t, "2026-05-22T00:00:00Z", got[2].OrderedAt)
}

func TestInmemoryRepository_DuplicateInsert(t *testing.T) {
	repo := NewInmemoryRepository()
	rec := sampleRecord()
	require.NoError(t, repo.Insert(context.Background(), rec))
	err := repo.Insert(context.Background(), rec)
	assert.ErrorIs(t, err, ErrInmemoryDuplicate)
}
