package order

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/bedrock"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/fallback"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanBuilder_BedrockSuccess(t *testing.T) {
	bMock := &bedrock.MockBedrockAdapter{
		InferOrderPlanFunc: func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
			return &bedrock.Plan{
				StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000, Category: "food",
				Source: "bedrock", BedrockLatencyMs: 800, BedrockAttempt: 1,
			}, nil
		},
	}
	fMock := &fallback.FakeFallbackProvider{}
	pb := NewBedrockPlanBuilder(bMock, fMock)

	plan, err := pb.Build(context.Background(), nil, "Mon", "food")
	require.NoError(t, err)
	assert.Equal(t, "bedrock", plan.Source)
	assert.False(t, plan.FallbackTriggered)
	assert.Equal(t, 0, fMock.BuildCalls)
	assert.Equal(t, 0, fMock.DefaultCalls)
}

func TestPlanBuilder_BedrockFailureFallbackHistory(t *testing.T) {
	bMock := &bedrock.MockBedrockAdapter{
		InferOrderPlanFunc: func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
			return nil, &types.ThrottlingException{Message: aws.String("Rate exceeded")}
		},
	}
	fMock := &fallback.FakeFallbackProvider{}
	pb := NewBedrockPlanBuilder(bMock, fMock)

	history := make([]OrderRecord, fallbackThreshold) // = 5 件、閾値ぴったり
	for i := range history {
		history[i] = OrderRecord{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Category: "food", Amount: 1000}
	}

	plan, err := pb.Build(context.Background(), history, "Tue", "food")
	require.NoError(t, err)
	assert.Equal(t, "fallback_history", plan.Source)
	assert.True(t, plan.FallbackTriggered)
	assert.Equal(t, 1, fMock.BuildCalls)
	assert.Equal(t, 0, fMock.DefaultCalls)
}

func TestPlanBuilder_BedrockFailureFallbackDefault(t *testing.T) {
	bMock := &bedrock.MockBedrockAdapter{
		InferOrderPlanFunc: func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
			return nil, errors.New("network down")
		},
	}
	fMock := &fallback.FakeFallbackProvider{}
	pb := NewBedrockPlanBuilder(bMock, fMock)

	// 4 件 (閾値未満)
	history := make([]OrderRecord, fallbackThreshold-1)

	plan, err := pb.Build(context.Background(), history, "Wed", "food")
	require.NoError(t, err)
	assert.Equal(t, "fallback_default", plan.Source)
	assert.True(t, plan.FallbackTriggered)
	assert.Equal(t, 0, fMock.BuildCalls)
	assert.Equal(t, 1, fMock.DefaultCalls)
}

func TestPlanBuilder_BedrockFailureBuildFromHistoryReturnsNil(t *testing.T) {
	// 閾値以上だが BuildFromHistory が nil を返した場合は Default にフェイルバック
	bMock := &bedrock.MockBedrockAdapter{
		InferOrderPlanFunc: func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
			return nil, errors.New("fail")
		},
	}
	fMock := &fallback.FakeFallbackProvider{
		BuildFromHistoryFunc: func(history []fallback.HistoryItem) *fallback.Plan {
			return nil
		},
	}
	pb := NewBedrockPlanBuilder(bMock, fMock)

	history := make([]OrderRecord, fallbackThreshold+1)
	plan, err := pb.Build(context.Background(), history, "Thu", "food")
	require.NoError(t, err)
	assert.Equal(t, "fallback_default", plan.Source)
	assert.Equal(t, 1, fMock.BuildCalls)
	assert.Equal(t, 1, fMock.DefaultCalls)
}

func TestPlanBuilder_ContextCanceled(t *testing.T) {
	bMock := &bedrock.MockBedrockAdapter{
		InferOrderPlanFunc: func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
			return nil, context.Canceled
		},
	}
	fMock := &fallback.FakeFallbackProvider{}
	pb := NewBedrockPlanBuilder(bMock, fMock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := pb.Build(ctx, nil, "Fri", "food")
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Equal(t, 0, fMock.BuildCalls)
	assert.Equal(t, 0, fMock.DefaultCalls)
}
