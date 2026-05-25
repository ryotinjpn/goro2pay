// Package metrics は GET /api/metrics のビジネスロジックを提供する。
//
// 凍結契約 §6.1 の MetricsService interface と Metrics DTO を実装する。
package metrics

import (
	"context"
	"errors"
)

// ErrNoBudgetSet は予算未設定ユーザに返す sentinel error (BR-M01)。
// HTTP 400 "ERR_NO_BUDGET_SET" にマッピングされる。
var ErrNoBudgetSet = errors.New("no budget set")

// Metrics は GET /api/metrics レスポンスの DTO (凍結契約 §6.1)。
type Metrics struct {
	DamageCount      int
	ConsumptionRate  float64 // 0.0..1.0
	MonthlyBudget    int
	RemainingBalance int
	ThresholdExceeded bool   // ConsumptionRate > 0.8
	SummaryText      string
}

// MetricsService は GetMetrics を公開する interface (凍結契約 §6.1)。
type MetricsService interface {
	GetMetrics(ctx context.Context, userID string) (*Metrics, error)
}
