package order

import (
	"context"
	"errors"
)

// WalletDeductResult は WalletService.Deduct の戻り値 (凍結契約 §3.1 整合)。
type WalletDeductResult struct {
	OrderID          string
	RemainingBalance int
	Idempotent       bool
}

// WalletService は Unit B が提供する残高引き落とし interface (凍結契約 §3.1)。
//
// Unit C は本 interface の実装に依存しない (Unit B の go module を直接 import
// しない方針)。Unit B 実装が完了したら main.go の DI 配線で実装を注入する。
type WalletService interface {
	// Deduct は idempotencyKey 単位で残高を amount だけ引き落とす。
	//
	// 同一 key の再送に対しては Idempotent=true で初回の OrderID を返す。
	// 残高不足の場合は order.ErrInsufficientFunds を返す。
	// 冪等キー衝突 (同一 key + 異なる amount 等) の場合は order.ErrIdempotencyConflict を返す。
	Deduct(ctx context.Context, userID, idempotencyKey string, amount int) (*WalletDeductResult, error)
}

// ErrWalletInsufficient / ErrWalletConflict は Unit B 内部の sentinel error 別名。
//
// Unit B が独自の error を返す可能性に備え、Unit C 側で errors.Is でラップ判定可能にする。
// 本実装では order package の sentinel をそのまま再 export する。
var (
	ErrWalletInsufficient = ErrInsufficientFunds
	ErrWalletConflict     = ErrIdempotencyConflict
)

// IsInsufficient は err が残高不足を示すかを判定する (errors.Is ラッパ)。
func IsInsufficient(err error) bool {
	return errors.Is(err, ErrInsufficientFunds)
}

// IsIdempotencyConflict は err が冪等キー衝突を示すかを判定する。
func IsIdempotencyConflict(err error) bool {
	return errors.Is(err, ErrIdempotencyConflict)
}
