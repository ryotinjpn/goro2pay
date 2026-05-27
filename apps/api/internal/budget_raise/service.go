package budget_raise

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
)

// Service は BudgetRaiseService の実装。
type Service struct {
	BudgetSettings       budget_settings.BudgetSettingsReader
	BudgetSettingsWriter budget_settings.BudgetSettingsWriter
}

// NewService は Service を返す。
func NewService(r budget_settings.BudgetSettingsReader, w budget_settings.BudgetSettingsWriter) *Service {
	return &Service{BudgetSettings: r, BudgetSettingsWriter: w}
}

// ComputeRecommendedBudget は現在の予算から推奨増額後予算を返す。
func (s *Service) ComputeRecommendedBudget(ctx context.Context, userID string) (int, error) {
	bs, err := s.BudgetSettings.Get(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("budget_raise: get budget settings: %w", err)
	}
	if bs == nil || bs.MonthlyBudget == 0 {
		return 0, ErrNoBudgetSet
	}
	return computeRecommendedBudget(bs.MonthlyBudget), nil
}

// Accept は newMonthlyBudget を検証し翌月 1 日 JST から適用する (BR-R02 / BR-R03 / BR-R04)。
func (s *Service) Accept(ctx context.Context, userID string, newMonthlyBudget int) (*BudgetRaiseResult, error) {
	if newMonthlyBudget < minBudget || newMonthlyBudget > maxBudget {
		return nil, ErrInvalidBudget
	}

	bs, err := s.BudgetSettings.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("budget_raise: get budget settings: %w", err)
	}
	if bs == nil {
		return nil, ErrNoBudgetSet
	}

	effectiveFrom := computeNextMonthStart()
	if err := s.BudgetSettingsWriter.Set(ctx, userID, newMonthlyBudget, effectiveFrom); err != nil {
		return nil, fmt.Errorf("budget_raise: set budget settings: %w", err)
	}

	slog.Info("budget_raise.Accept", "userId", userID, "newBudget", newMonthlyBudget, "effectiveFrom", effectiveFrom)
	return &BudgetRaiseResult{
		NewMonthlyBudget: newMonthlyBudget,
		AppliedFrom:      effectiveFrom,
	}, nil
}
