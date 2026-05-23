package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/bedrock"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/delivery"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/fallback"
	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type harness struct {
	bedrock     *bedrock.MockBedrockAdapter
	fallback    *fallback.FakeFallbackProvider
	delivery    *delivery.FakeDeliveryAdapter
	wallet      *WalletStub
	historyRepo *orderhistory.InmemoryRepository
	svc         *Service
}

func newHarness(initialBalance int) *harness {
	b := &bedrock.MockBedrockAdapter{}
	f := &fallback.FakeFallbackProvider{}
	d := &delivery.FakeDeliveryAdapter{}
	w := NewWalletStub(initialBalance)
	w.SetBalance("user-1", initialBalance)
	r := orderhistory.NewInmemoryRepository()
	pb := NewBedrockPlanBuilder(b, f)
	svc := NewService(r, pb, d, w)
	svc.SetClock(func() time.Time { return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC) })
	return &harness{bedrock: b, fallback: f, delivery: d, wallet: w, historyRepo: r, svc: svc}
}

// Scenario 1: 通常パス (Bedrock 1 回成功)
func TestService_NormalPath(t *testing.T) {
	h := newHarness(10000)
	h.bedrock.InferOrderPlanFunc = func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
		return &bedrock.Plan{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000, Category: "food", Source: "bedrock", BedrockAttempt: 1}, nil
	}

	res, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{
		Category: "food", IdempotencyKey: "01HZIDEM",
	})
	require.NoError(t, err)
	assert.Equal(t, "ゴロゴロ食堂", res.StoreName)
	assert.Equal(t, 9000, res.RemainingBalance)
	assert.False(t, res.Idempotent)
	assert.Equal(t, 1, h.historyRepo.Len())
	assert.Equal(t, 1, h.delivery.Calls)
}

// Scenario 2: リトライ後成功 (Bedrock 1 回目 Throttle → 2 回目成功)
func TestService_RetryThenSuccess(t *testing.T) {
	h := newHarness(10000)
	calls := 0
	// MockBedrockAdapter は 1 リクエストにつき 1 回しか呼ばれない (PlanBuilder の中で
	// ClaudeBedrockAdapter のリトライが起きるが、本テストでは Mock が直接 Plan を
	// 返すためリトライ挙動は ClaudeBedrockAdapter のテスト側に任せる。
	// ここでは「リトライ後成功した」という最終結果を Mock が返す形で再現する)
	h.bedrock.InferOrderPlanFunc = func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
		calls++
		return &bedrock.Plan{StoreName: "ぐうたら亭", MenuName: "手抜き丼", Amount: 800, Category: "food", Source: "bedrock", BedrockAttempt: 2}, nil
	}

	res, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{
		Category: "food", IdempotencyKey: "01HZIDEM2",
	})
	require.NoError(t, err)
	assert.Equal(t, "ぐうたら亭", res.StoreName)
}

// Scenario 3: フォールバック発動 (Bedrock 失敗、履歴 5 件 → BuildFromHistory)
func TestService_FallbackHistory(t *testing.T) {
	h := newHarness(10000)
	for i := 0; i < 5; i++ {
		_ = h.historyRepo.Insert(context.Background(), &orderhistory.OrderRecord{
			OrderID:   "old-" + string(rune('a'+i)),
			UserID:    "user-1",
			StoreName: "ぐうたら亭", MenuName: "手抜き丼", Category: "food", Amount: 800,
			OrderedAt: time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
		})
	}
	h.bedrock.InferOrderPlanFunc = func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
		return nil, errors.New("bedrock down")
	}
	h.fallback.BuildFromHistoryFunc = func(history []fallback.HistoryItem) *fallback.Plan {
		return &fallback.Plan{StoreName: "ぐうたら亭", MenuName: "手抜き丼", Amount: 800, Category: "food", Source: "fallback_history"}
	}

	res, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{
		Category: "food", IdempotencyKey: "01HZIDEM3",
	})
	require.NoError(t, err)
	assert.Equal(t, "ぐうたら亭", res.StoreName)
	assert.Equal(t, 1, h.fallback.BuildCalls)
}

// Scenario 4: 永続エラー → 即フォールバック Default
func TestService_PermanentErrorFallbackDefault(t *testing.T) {
	h := newHarness(10000)
	h.bedrock.InferOrderPlanFunc = func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
		return nil, &types.ValidationException{Message: aws.String("bad params")}
	}

	res, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{
		Category: "food", IdempotencyKey: "01HZIDEM4",
	})
	require.NoError(t, err)
	assert.Equal(t, "fake-default", res.StoreName)
	assert.Equal(t, 1, h.fallback.DefaultCalls)
}

// Scenario 5: 冪等命中
func TestService_IdempotencyHit(t *testing.T) {
	h := newHarness(10000)
	h.bedrock.InferOrderPlanFunc = func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
		return &bedrock.Plan{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000, Category: "food", Source: "bedrock"}, nil
	}

	first, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{
		Category: "food", IdempotencyKey: "01HZSAME",
	})
	require.NoError(t, err)

	second, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{
		Category: "food", IdempotencyKey: "01HZSAME",
	})
	require.NoError(t, err)

	assert.Equal(t, first.OrderID, second.OrderID)
	assert.True(t, second.Idempotent)
	assert.Equal(t, 9000, h.wallet.Balance("user-1"), "balance must change only once")
}

// Scenario 6: 連打 (異なる idempotencyKey で 2 回連続成功)
func TestService_BurstWithDifferentKeys(t *testing.T) {
	h := newHarness(10000)
	h.bedrock.InferOrderPlanFunc = func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
		return &bedrock.Plan{StoreName: "X", MenuName: "Y", Amount: 1000, Category: "food", Source: "bedrock"}, nil
	}

	_, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{Category: "food", IdempotencyKey: "k1"})
	require.NoError(t, err)
	_, err = h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{Category: "food", IdempotencyKey: "k2"})
	require.NoError(t, err)

	assert.Equal(t, 8000, h.wallet.Balance("user-1"))
}

// Scenario 7: 残高不足
func TestService_InsufficientFunds(t *testing.T) {
	h := newHarness(500) // amount 1000 < balance 500
	h.bedrock.InferOrderPlanFunc = func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
		return &bedrock.Plan{StoreName: "X", MenuName: "Y", Amount: 1000, Category: "food", Source: "bedrock"}, nil
	}

	_, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{Category: "food", IdempotencyKey: "k-poor"})
	assert.True(t, IsInsufficient(err))
}

// Scenario 8: Context Canceled
func TestService_ContextCanceled(t *testing.T) {
	h := newHarness(10000)
	h.bedrock.InferOrderPlanFunc = func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
		return nil, context.Canceled
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := h.svc.PlaceOrder(ctx, "user-1", PlaceOrderRequest{Category: "food", IdempotencyKey: "k-cancel"})
	require.Error(t, err)
}

// Scenario 9: invalid request (空の category) → 400 系エラー
func TestService_InvalidRequest(t *testing.T) {
	h := newHarness(10000)
	_, err := h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{Category: "", IdempotencyKey: "k"})
	assert.Error(t, err)
	_, err = h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{Category: "errand", IdempotencyKey: "k"})
	assert.Error(t, err)
	_, err = h.svc.PlaceOrder(context.Background(), "user-1", PlaceOrderRequest{Category: "food", IdempotencyKey: ""})
	assert.Error(t, err)
}

func TestService_GetHistory(t *testing.T) {
	h := newHarness(10000)
	for i := 0; i < 3; i++ {
		_ = h.historyRepo.Insert(context.Background(), &orderhistory.OrderRecord{
			OrderID:   "id-" + string(rune('a'+i)),
			UserID:    "user-1",
			OrderedAt: time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
		})
	}
	got, err := h.svc.GetHistory(context.Background(), "user-1", 10)
	require.NoError(t, err)
	assert.Len(t, got, 3)
}
