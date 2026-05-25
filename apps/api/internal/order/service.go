package order

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/delivery"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/observability"
	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
)

// OrderService は Unit C の中核 interface (凍結契約 §4.1)。
type OrderService interface {
	PlaceOrder(ctx context.Context, userID string, req PlaceOrderRequest) (*PlaceOrderResult, error)
	GetHistory(ctx context.Context, userID string, limit int) ([]*OrderRecord, error)
}

// Service は OrderService implementation。
//
// 依存は constructor 注入 (P-DI-01)。Wallet / PlanBuilder / DeliveryAdapter /
// OrderHistoryRepository を interface として受け取り、本体実装にもテスト
// (PBT / 統合) にも mock を差し替え可能にする。
type Service struct {
	historyRepo     orderhistory.OrderHistoryRepository
	planBuilder     PlanBuilder
	delivery        delivery.DeliveryAdapter
	wallet          WalletService
	suggestResolver SuggestResolver  // Unit D 配線 (任意、未設定時は Bedrock 推論のみ)
	now             func() time.Time // テスト時刻固定用 (注入式)
}

// NewService は production 用 Service を返す。
func NewService(
	historyRepo orderhistory.OrderHistoryRepository,
	planBuilder PlanBuilder,
	deliveryAdapter delivery.DeliveryAdapter,
	wallet WalletService,
) *Service {
	return &Service{
		historyRepo: historyRepo,
		planBuilder: planBuilder,
		delivery:    deliveryAdapter,
		wallet:      wallet,
		now:         time.Now,
	}
}

// SetClock は now() 関数を差し替える (テスト用、PBT で時刻固定するため)。
func (s *Service) SetClock(now func() time.Time) {
	s.now = now
}

// historyTTLDuration は OrderHistory の TTL (NFRC-C12 / 凍結契約整合)。
const historyTTLDuration = 90 * 24 * time.Hour

// PlaceOrder は Unit C のオーケストレーションを実行する:
//
//  1. 直近履歴を取得 (フォールバック分岐閾値判定 BR-C06 用)
//  2. PlanBuilder で plan を生成 (Bedrock or fallback)
//  3. WalletService.Deduct で残高引き落とし (冪等性チェック)
//  4. 冪等命中時は History から既存 OrderRecord を復元して返す
//  5. 通常時: DeliveryAdapter.Place → OrderHistory.Insert → 結果組み立て
//
// 失敗時も `defer summary.LogComplete()` で 11 項目サマリログを必ず出力する。
func (s *Service) PlaceOrder(ctx context.Context, userID string, req PlaceOrderRequest) (*PlaceOrderResult, error) {
	summary := NewLogSummary(ctx)
	defer summary.LogComplete()

	if err := validateRequest(req); err != nil {
		return nil, err
	}

	summary.SetIdempotencyKey(req.IdempotencyKey)
	summary.SetCategory(req.Category)
	if req.SuggestionID != nil {
		summary.SetSource(string(SourceSuggestion))
	} else {
		summary.SetSource(string(SourceButton))
	}

	now := s.now().UTC()
	dayOfWeek := now.Weekday().String()

	// 1. 直近履歴を取得 (フォールバック分岐閾値判定 BR-C06)
	history, _, err := observability.Measure(ctx, "history_query", func() ([]*orderhistory.OrderRecord, error) {
		return s.historyRepo.Query(ctx, userID, fallbackThreshold*6) // 30 件、Bedrock プロンプト用
	})
	if err != nil {
		// 履歴取得失敗は警告、空履歴で続行 (NFRC-C09 同様にユーザ体験優先)
		history = []*orderhistory.OrderRecord{}
	}
	summary.SetHistoryCount(len(history))

	// 2. plan 生成: suggestionId があれば保存済み提案を優先 (Q-DG1=B / FD §2.3 / BR-C09)、
	//    失効・未配線・解決失敗時は通常の Bedrock 推論にフォールバック (BR-C10、透過)。
	historyForPlan := toServiceHistory(history)
	var planResult *Plan
	if req.SuggestionID != nil && s.suggestResolver != nil {
		if resolved, rerr := s.suggestResolver.ResolveSuggestion(ctx, *req.SuggestionID); rerr == nil && resolved != nil {
			planResult = &Plan{
				StoreName:         resolved.StoreName,
				MenuName:          resolved.MenuName,
				Amount:            resolved.Amount,
				Category:          resolved.Category,
				Source:            "suggestion",
				FallbackTriggered: false,
			}
		}
	}
	if planResult == nil {
		built, _, berr := observability.Measure(ctx, "plan_build", func() (*Plan, error) {
			return s.planBuilder.Build(ctx, historyForPlan, dayOfWeek, req.Category)
		})
		if berr != nil {
			return nil, fmt.Errorf("plan: %w", berr)
		}
		planResult = built
	}
	summary.SetBedrockLatencyMs(planResult.BedrockLatencyMs)
	summary.SetBedrockAttempt(planResult.BedrockAttempt)
	summary.SetFallbackTriggered(planResult.FallbackTriggered)
	summary.SetStoreName(planResult.StoreName)
	summary.SetMenuName(planResult.MenuName)
	summary.SetAmount(planResult.Amount)

	// 3. Wallet.Deduct
	deduct, _, err := observability.Measure(ctx, "wallet_deduct", func() (*WalletDeductResult, error) {
		return s.wallet.Deduct(ctx, userID, req.IdempotencyKey, planResult.Amount)
	})
	if err != nil {
		return nil, err
	}

	// 4. 冪等命中時: Wallet payload から OrderID を取得し既存 record を返す (BR-C15)
	if deduct.Idempotent {
		summary.SetIdempotent(true)
		summary.SetOrderID(deduct.OrderID)
		// 既存 record の orderedAt は不明だが、最新 history を Query して復元
		// (Idempotent な再送はまれ、性能影響軽微)
		recs, qerr := s.historyRepo.Query(ctx, userID, 100)
		if qerr == nil {
			for _, r := range recs {
				if r.OrderID == deduct.OrderID {
					return &PlaceOrderResult{
						OrderID:          r.OrderID,
						StoreName:        r.StoreName,
						MenuName:         r.MenuName,
						Amount:           r.Amount,
						RemainingBalance: deduct.RemainingBalance,
						Idempotent:       true,
					}, nil
				}
			}
		}
		// History から復元できない場合 (極稀) でも plan 情報で応答 (P-3 緩和)
		return &PlaceOrderResult{
			OrderID:          deduct.OrderID,
			StoreName:        planResult.StoreName,
			MenuName:         planResult.MenuName,
			Amount:           planResult.Amount,
			RemainingBalance: deduct.RemainingBalance,
			Idempotent:       true,
		}, nil
	}

	// 5. DeliveryAdapter.Place
	orderID := deduct.OrderID
	if orderID == "" {
		orderID = generateULID(now)
	}
	summary.SetOrderID(orderID)

	_, _, derr := observability.Measure(ctx, "delivery_place", func() (any, error) {
		return nil, s.delivery.Place(ctx, delivery.PlaceOrderRequest{
			OrderID:        orderID,
			UserID:         userID,
			StoreName:      planResult.StoreName,
			MenuName:       planResult.MenuName,
			Amount:         planResult.Amount,
			Category:       planResult.Category,
			IdempotencyKey: req.IdempotencyKey,
		})
	})
	if derr != nil {
		// Adapter 失敗時もユーザ体験優先で 200 返却 (FD Q-9=A)
		// summary に historyInsertFailed フラグは持たせていないが、ログレベル別経路で記録
	}

	// 6. OrderHistory.Insert (失敗時も 200 応答 NFRC-C09)
	rec := &orderhistory.OrderRecord{
		OrderID:        orderID,
		UserID:         userID,
		Category:       planResult.Category,
		StoreName:      planResult.StoreName,
		MenuName:       planResult.MenuName,
		Amount:         planResult.Amount,
		OrderedAt:      now.Format(time.RFC3339),
		IdempotencyKey: req.IdempotencyKey,
		DayOfWeek:      dayOfWeek,
		Source:         planResult.Source,
		ExpiresAt:      now.Add(historyTTLDuration).Unix(),
	}
	_, _, ierr := observability.Measure(ctx, "history_insert", func() (any, error) {
		return nil, s.historyRepo.Insert(ctx, rec)
	})
	_ = ierr // NFRC-C09: 失敗時も 200 応答、ログは observability.Measure 経由で出力済み

	return &PlaceOrderResult{
		OrderID:          orderID,
		StoreName:        planResult.StoreName,
		MenuName:         planResult.MenuName,
		Amount:           planResult.Amount,
		RemainingBalance: deduct.RemainingBalance,
		Idempotent:       false,
	}, nil
}

// GetHistory はユーザの注文履歴を返す (凍結契約 §4.1)。
func (s *Service) GetHistory(ctx context.Context, userID string, limit int) ([]*OrderRecord, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	recs, err := s.historyRepo.Query(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	out := make([]*OrderRecord, len(recs))
	for i, r := range recs {
		out[i] = &OrderRecord{
			OrderID:        r.OrderID,
			UserID:         r.UserID,
			Category:       r.Category,
			StoreName:      r.StoreName,
			MenuName:       r.MenuName,
			Amount:         r.Amount,
			OrderedAt:      r.OrderedAt,
			IdempotencyKey: r.IdempotencyKey,
			DayOfWeek:      r.DayOfWeek,
			Source:         r.Source,
		}
	}
	return out, nil
}

// validateRequest は PlaceOrderRequest の最低限の整合性を検証する。
func validateRequest(req PlaceOrderRequest) error {
	if req.IdempotencyKey == "" {
		return errors.New("idempotencyKey is required")
	}
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Category != "food" {
		// MVP は food のみ受理 (BR-C27)、将来拡張は凍結契約 §4.1 に従う
		return errors.New("category must be 'food'")
	}
	return nil
}

// generateULID は OrderID を採番する。Wallet が OrderID を返さない場合に使用。
//
// crypto/rand エントロピーを使い、now と組み合わせて ULID を生成する。
func generateULID(now time.Time) string {
	t := ulid.Timestamp(now)
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(t, entropy).String()
}

// toServiceHistory は repo の OrderRecord ポインタ slice を service の値 slice に変換する。
func toServiceHistory(in []*orderhistory.OrderRecord) []OrderRecord {
	out := make([]OrderRecord, len(in))
	for i, r := range in {
		out[i] = OrderRecord{
			StoreName: r.StoreName,
			MenuName:  r.MenuName,
			Category:  r.Category,
			Amount:    r.Amount,
			DayOfWeek: r.DayOfWeek,
		}
	}
	return out
}
