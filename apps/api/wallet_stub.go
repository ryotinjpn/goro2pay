// 本ファイルは Unit B WalletService の本実装が main.go に DI 配線されるまでの
// 暫定 noop 実装。Unit B Code Generation 完了時に削除し、Unit B 実装に差し替える。
package main

import (
	"context"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/order"
)

// noopWalletService は Unit B WalletService 未実装時の暫定 stub。
//
// 常に ErrInsufficientFunds を返すため、API は POST /api/orders で 402 を返す。
// これにより Frontend は「残高不足」フローのみ動作確認できる。Unit B 完成後
// に本 stub を削除し main.go から実装を注入する。
type noopWalletService struct{}

// newNoopWalletService は noopWalletService を返す。
func newNoopWalletService() *noopWalletService {
	return &noopWalletService{}
}

// Deduct は常に order.ErrInsufficientFunds を返す。
func (n *noopWalletService) Deduct(ctx context.Context, userID, idempotencyKey string, amount int) (*order.WalletDeductResult, error) {
	return nil, order.ErrInsufficientFunds
}
