package wallet

import (
	"context"
	"errors"
	"fmt"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/order"
)

// OrderAdapter は Unit B `wallet.Service` を Unit C が期待する
// `order.WalletService` interface に変換するアダプタ。
//
// 引数順序の差異:
//   - order.WalletService.Deduct(userID, idempotencyKey, amount)
//   - wallet.Service.Deduct(userID, amount, idempotencyKey)
//
// 戻り値変換:
//   - wallet.DeductResult{NewBalance, Idempotent}
//     → order.WalletDeductResult{OrderID(=idempotencyKey), RemainingBalance, Idempotent}
//
// sentinel error 変換 (凍結 IF §8 の HTTP マッピングは Handler 層で行うので
// ここでは「Unit C が認識する sentinel」に揃える):
//   - wallet.ErrInsufficientBalance → order.ErrInsufficientFunds
//   - wallet.ErrIdempotencyConflict → order.ErrIdempotencyConflict
type OrderAdapter struct {
	svc WalletService
}

// NewOrderAdapter は OrderAdapter を返す。
func NewOrderAdapter(svc WalletService) *OrderAdapter {
	return &OrderAdapter{svc: svc}
}

// 確認: order.WalletService interface を実装する。
var _ order.WalletService = (*OrderAdapter)(nil)

// Deduct は wallet.Service.Deduct に委譲し、戻り値とエラーを order package の
// 形式に変換する。
//
// OrderID は idempotencyKey を流用 (wallet.DeductResult には OrderID が無いため、
// Unit C 側で別途 OrderID を生成する設計を尊重するなら呼び出し側で上書き可)。
func (a *OrderAdapter) Deduct(ctx context.Context, userID, idempotencyKey string, amount int) (*order.WalletDeductResult, error) {
	res, err := a.svc.Deduct(ctx, userID, amount, idempotencyKey)
	if err != nil {
		switch {
		case errors.Is(err, ErrInsufficientBalance):
			return nil, order.ErrInsufficientFunds
		case errors.Is(err, ErrIdempotencyConflict):
			return nil, order.ErrIdempotencyConflict
		case errors.Is(err, ErrIdempotencyInProgress):
			// Unit C 側に in-progress sentinel が無いため wrap して透過。Handler の
			// default 500 にマップされる (本来は 503 が望ましいが、Unit C 改修不要を
			// 優先)。Wallet 直 endpoint からは 503 として返る。
			return nil, fmt.Errorf("wallet adapter: %w", err)
		default:
			return nil, fmt.Errorf("wallet adapter: %w", err)
		}
	}
	return &order.WalletDeductResult{
		OrderID:          idempotencyKey,
		RemainingBalance: res.NewBalance,
		Idempotent:       res.Idempotent,
	}, nil
}
