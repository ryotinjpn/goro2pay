package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/wallet_repo"
)

// Service は MetricsService の DynamoDB 実装 (P-ME-PARALLEL-01)。
type Service struct {
	BudgetSettings budget_settings.BudgetSettingsReader
	Wallet         wallet_repo.WalletReader
	OrderHistory   orderhistory.OrderHistoryReader
}

// NewService は Service を返す。
func NewService(
	bs budget_settings.BudgetSettingsReader,
	w wallet_repo.WalletReader,
	oh orderhistory.OrderHistoryReader,
) *Service {
	return &Service{BudgetSettings: bs, Wallet: w, OrderHistory: oh}
}

// GetMetrics は P-ME-PARALLEL-01 に従い BudgetSettings + Wallet を並列取得し、
// CountThisMonth を直列呼出してメトリクスを返す。
func (s *Service) GetMetrics(ctx context.Context, userID string) (*Metrics, error) {
	start := time.Now()

	// Phase 1: 並列読取
	g, gctx := errgroup.WithContext(ctx)
	var bs *budget_settings.BudgetSettings
	var w *wallet_repo.WalletRecord

	g.Go(func() error {
		v, err := s.BudgetSettings.Get(gctx, userID)
		if err != nil {
			return fmt.Errorf("budget_settings: %w", err)
		}
		bs = v
		return nil
	})
	g.Go(func() error {
		v, err := s.Wallet.Get(gctx, userID)
		if err != nil {
			return fmt.Errorf("wallet: %w", err)
		}
		w = v
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("metrics: parallel read: %w", err)
	}

	// Phase 2: 未設定チェック (BR-M01)
	if bs == nil || bs.MonthlyBudget == 0 {
		return nil, ErrNoBudgetSet
	}
	if w == nil {
		return nil, ErrNoBudgetSet
	}

	// Phase 3: CountThisMonth 直列呼出
	count, err := s.OrderHistory.CountThisMonth(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("metrics: count_this_month: %w", err)
	}

	latencyMs := time.Since(start).Milliseconds()
	slog.Info("metrics.GetMetrics", "userId", userID, "latencyMs", latencyMs, "damageCount", count)

	return computeMetrics(bs, w, count), nil
}
