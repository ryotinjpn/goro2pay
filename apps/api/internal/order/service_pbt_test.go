package order

import (
	"context"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/bedrock"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/delivery"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/fallback"
	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
)

// makeSvcForPBT は PBT 用のサービスを構築する。
//
// Bedrock は「常に決定論的な Plan を返す」mock。Wallet と Repository は
// in-memory 実装を使い、N 回送信時の状態遷移を実装と同じパターンで再現する。
func makeSvcForPBT(initialBalance int) (*Service, *WalletStub, *orderhistory.InmemoryRepository) {
	b := &bedrock.MockBedrockAdapter{
		InferOrderPlanFunc: func(ctx context.Context, history []bedrock.HistoryItem, dayOfWeek string, category string) (*bedrock.Plan, error) {
			return &bedrock.Plan{
				StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食",
				Amount: 1000, Category: "food", Source: "bedrock", BedrockAttempt: 1,
			}, nil
		},
	}
	f := &fallback.FakeFallbackProvider{}
	d := &delivery.FakeDeliveryAdapter{}
	w := NewWalletStub(initialBalance)
	w.SetBalance("user-pbt", initialBalance)
	r := orderhistory.NewInmemoryRepository()
	pb := NewBedrockPlanBuilder(b, f)
	svc := NewService(r, pb, d, w)
	svc.SetClock(func() time.Time { return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC) })
	return svc, w, r
}

// P-1: 同一 idempotencyKey の N 回送信に対し、残高変動は最大 1 回。
func TestPBT_P1_BalanceInvariantUnderRepeatedRequests(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)
	properties.Property("P-1: same idempotencyKey deducts at most once",
		prop.ForAll(
			func(idempotencyKey string, n int) bool {
				svc, w, _ := makeSvcForPBT(100000)
				before := w.Balance("user-pbt")
				for i := 0; i < n; i++ {
					_, err := svc.PlaceOrder(context.Background(), "user-pbt", PlaceOrderRequest{
						Category: "food", IdempotencyKey: idempotencyKey,
					})
					if err != nil {
						return false
					}
				}
				after := w.Balance("user-pbt")
				diff := before - after
				// 残高変動は plan.Amount (1000) ぴったりか、または変動なし (n=0 のとき)
				return diff == 0 || diff == 1000
			},
			gen.RegexMatch(`^[0-9A-HJKMNP-TV-Z]{26}$`),
			gen.IntRange(1, 5),
		))

	properties.TestingRun(t)
}

// P-3: 同一 idempotencyKey の N 回送信に対し、レスポンス payload は 2 回目以降完全一致。
func TestPBT_P3_IdempotentResponseConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)
	properties.Property("P-3: idempotent responses are identical",
		prop.ForAll(
			func(idempotencyKey string, n int) bool {
				svc, _, _ := makeSvcForPBT(100000)

				first, err := svc.PlaceOrder(context.Background(), "user-pbt", PlaceOrderRequest{
					Category: "food", IdempotencyKey: idempotencyKey,
				})
				if err != nil {
					return false
				}

				for i := 1; i < n; i++ {
					res, err := svc.PlaceOrder(context.Background(), "user-pbt", PlaceOrderRequest{
						Category: "food", IdempotencyKey: idempotencyKey,
					})
					if err != nil {
						return false
					}
					if res.OrderID != first.OrderID {
						return false
					}
					if res.Amount != first.Amount {
						return false
					}
					if res.StoreName != first.StoreName {
						return false
					}
					if res.MenuName != first.MenuName {
						return false
					}
					if !res.Idempotent {
						return false
					}
				}
				return true
			},
			gen.RegexMatch(`^[0-9A-HJKMNP-TV-Z]{26}$`),
			gen.IntRange(2, 5),
		))

	properties.TestingRun(t)
}
