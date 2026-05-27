package suggestion

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
)

// mockDDB は DynamoDBAPI の test mock (in-memory map)。
type mockDDB struct {
	items map[string]map[string]types.AttributeValue
}

func (m *mockDDB) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if m.items == nil {
		m.items = map[string]map[string]types.AttributeValue{}
	}
	id := in.Item["suggestionId"].(*types.AttributeValueMemberS).Value
	m.items[id] = in.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (m *mockDDB) GetItem(_ context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	id := in.Key["suggestionId"].(*types.AttributeValueMemberS).Value
	item, ok := m.items[id]
	if !ok {
		return &dynamodb.GetItemOutput{}, nil
	}
	return &dynamodb.GetItemOutput{Item: item}, nil
}

func TestRepository_SaveGet_RoundTrip(t *testing.T) {
	repo := NewRepositoryWithClient(&mockDDB{}, "GoroPay_Suggestion")
	rec := &SuggestionRecord{
		SuggestionID: "01HXSAVEGET",
		UserID:       "u1",
		Plan:         Plan{StoreName: "CoCo壱", MenuName: "ポークカレー", Amount: 1200, Category: "food"},
		CreatedAt:    100,
		ExpiresAt:    1900,
	}
	require.NoError(t, repo.Save(context.Background(), rec))

	got, err := repo.Get(context.Background(), "01HXSAVEGET")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "CoCo壱", got.Plan.StoreName)
	require.Equal(t, 1200, got.Plan.Amount)
	require.Equal(t, "food", got.Plan.Category)
	require.Equal(t, int64(1900), got.ExpiresAt)
}

func TestRepository_Get_Missing_ReturnsNil(t *testing.T) {
	repo := NewRepositoryWithClient(&mockDDB{items: map[string]map[string]types.AttributeValue{}}, "t")
	got, err := repo.Get(context.Background(), "nonexistent")
	require.NoError(t, err)
	require.Nil(t, got)
}
