// Package budget_raise は翌月予算増額のビジネスロジックを提供する。
//
// 凍結契約 §6.1 の BudgetRaiseService interface と BudgetRaiseResult DTO を実装する。
package budget_raise

import (
	"context"
	"errors"
	"time"
)

// ErrInvalidBudget は newMonthlyBudget が範囲外のときに返す sentinel error (BR-R02)。
var ErrInvalidBudget = errors.New("invalid budget")

// ErrNoBudgetSet は BudgetSettings 未作成ユーザに返す sentinel error。
var ErrNoBudgetSet = errors.New("no budget set")

// BudgetRaiseResult は POST /api/budget/raise の応答 DTO (凍結契約 §6.1)。
type BudgetRaiseResult struct {
	NewMonthlyBudget int
	AppliedFrom      time.Time // 翌月 1 日 00:00 JST (BR-R03)
}

// BudgetRaiseService は ComputeRecommendedBudget / Accept を公開する interface。
type BudgetRaiseService interface {
	ComputeRecommendedBudget(ctx context.Context, userID string) (int, error)
	Accept(ctx context.Context, userID string, newMonthlyBudget int) (*BudgetRaiseResult, error)
}
