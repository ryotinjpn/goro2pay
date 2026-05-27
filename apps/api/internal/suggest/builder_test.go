package suggest

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/bedrock"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/fallback"
)

func TestBuilder_BedrockSuccess(t *testing.T) {
	mock := &bedrock.MockBedrockAdapter{
		InferSuggestionFunc: func(_ context.Context, _ []bedrock.HistoryItem, _ string) (*bedrock.Plan, error) {
			return &bedrock.Plan{StoreName: "CoCo壱", MenuName: "カレー", Amount: 1200, Category: "food", BedrockAttempt: 1}, nil
		},
	}
	b := NewBedrockSuggestionBuilder(mock, fallback.NewSimpleFallbackProvider())

	got, err := b.Build(context.Background(), recordsDaysAgo(1, 2, 3, 4, 5), "Monday")
	require.NoError(t, err)
	require.False(t, got.FallbackUsed)
	require.Equal(t, "CoCo壱", got.Plan.StoreName)
	require.Equal(t, 1200, got.Plan.Amount)
}

func TestBuilder_BedrockFailure_FallsBackToHistory(t *testing.T) {
	mock := &bedrock.MockBedrockAdapter{
		InferSuggestionFunc: func(_ context.Context, _ []bedrock.HistoryItem, _ string) (*bedrock.Plan, error) {
			return nil, errors.New("bedrock down")
		},
	}
	b := NewBedrockSuggestionBuilder(mock, fallback.NewSimpleFallbackProvider())

	// 5 件すべて同一 (CoCo壱/ポークカレー) → 最頻フォールバックが成立
	got, err := b.Build(context.Background(), recordsDaysAgo(1, 2, 3, 4, 5), "Monday")
	require.NoError(t, err)
	require.True(t, got.FallbackUsed)
	require.NotEmpty(t, got.Plan.StoreName)
}
