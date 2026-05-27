package suggest

import (
	"context"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/order"
)

// OrderResolverAdapter は suggest.SuggestService を order.SuggestResolver に
// ブリッジする (Q-DG1=B)。
//
// order package が suggest を import しない (独自 ResolvedSuggestion 型を持つ)
// ため、依存方向を suggest → order の一方向に保つための adapter。
// wallet.OrderAdapter と同じ設計。
type OrderResolverAdapter struct {
	svc SuggestService
}

// NewOrderResolverAdapter は adapter を返す。
func NewOrderResolverAdapter(svc SuggestService) *OrderResolverAdapter {
	return &OrderResolverAdapter{svc: svc}
}

// ResolveSuggestion は order.SuggestResolver の実装。
//
// suggest.SuggestionPlan を order.ResolvedSuggestion に変換する。失効・不在時は
// nil, nil をそのまま伝播し、Unit C が透過フォールバックする (BR-C10)。
func (a *OrderResolverAdapter) ResolveSuggestion(ctx context.Context, suggestionID string) (*order.ResolvedSuggestion, error) {
	plan, err := a.svc.ResolveSuggestion(ctx, suggestionID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, nil
	}
	return &order.ResolvedSuggestion{
		StoreName: plan.StoreName,
		MenuName:  plan.MenuName,
		Amount:    plan.Amount,
		Category:  plan.Category,
	}, nil
}
