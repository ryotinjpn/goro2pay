// 本ファイルは Unit B WalletService の本実装が main.go に DI 配線されるまでの
// 暫定 stub。Unit B Code Generation 完了時に削除し、Unit B 実装に差し替える。
package main

import (
	"context"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/order"
)

// unconfiguredWalletService は Unit B WalletService 未実装時の暫定 stub。
//
// レビュー指摘 (B-C2) 対応: 当初 ErrInsufficientFunds を返していたが、
// production にデプロイされると「全注文が 402 INSUFFICIENT_FUNDS」となり、
// 以下の問題があった:
//   1. NFRC-C22 の 402 メトリクス (Unit E 等が観測予定) を恒常的に汚染する
//   2. Unit B 実装到達後の "本物の残高不足" との区別がつかない
//   3. Frontend の `/budget-empty` 遷移が常に発火し、UX が壊れて見える
//
// 修正後は order.ErrWalletUnconfigured を返し、Handler 側で 503
// SERVICE_UNAVAILABLE にマッピングする。これにより:
//   - 観測 (CloudWatch): 402 と 503 が明確に区別される
//   - ユーザ向け: 「メンテナンス中」相当のメッセージ
//   - Unit B 完成時: main.go の DI 配線で実装を差替えるだけ
type unconfiguredWalletService struct{}

// newUnconfiguredWalletService は unconfiguredWalletService を返す。
func newUnconfiguredWalletService() *unconfiguredWalletService {
	return &unconfiguredWalletService{}
}

// Deduct は常に order.ErrWalletUnconfigured を返す (Unit B 未配線)。
func (n *unconfiguredWalletService) Deduct(ctx context.Context, userID, idempotencyKey string, amount int) (*order.WalletDeductResult, error) {
	return nil, order.ErrWalletUnconfigured
}
