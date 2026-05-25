// Package wallet_repo は Wallet テーブルへの DynamoDB アクセスを提供する。
//
// 公開 IF (凍結 IF §3.2):
//   - WalletReader: Unit E が読取参照する subset
//   - WalletRecord: 公開 DTO
//
// Write 系 (DeductConditional / SetBalance / Create / UpdateBalance / ListAllUserIDs)
// は Unit B 内部 (wallet.Service) のみが利用する。
package wallet_repo

import (
	"context"
	"errors"
	"time"
)

// WalletRecord は Wallet テーブルの 1 行に対応する公開 DTO (凍結 IF §3.2)。
type WalletRecord struct {
	UserID    string
	Balance   int
	UpdatedAt time.Time
}

// WalletReader は Unit E (MetricsService) が読取参照する subset。
type WalletReader interface {
	Get(ctx context.Context, userID string) (*WalletRecord, error)
}

// ErrInsufficientBalance は DeductConditional で残高不足の場合に返される
// repo 内部 sentinel。Service 層 (wallet.Service) で wallet.ErrInsufficientBalance
// に変換される (循環 import 回避のため repo 側の sentinel として定義)。
var ErrInsufficientBalance = errors.New("wallet_repo: insufficient balance")
