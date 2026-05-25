package wallet

import (
	"context"
	"time"
)

// WalletService は Unit B の公開 Service interface (凍結 IF §3.1)。
type WalletService interface {
	GetBalance(ctx context.Context, userID string) (*WalletSnapshot, error)
	SetBudget(ctx context.Context, userID string, monthlyBudget int) error
	Deduct(ctx context.Context, userID string, amount int, idempotencyKey string) (*DeductResult, error)
	ResetAll(ctx context.Context) (*ResetResult, error)
}

// WalletSnapshot は GetBalance の戻り値 (凍結 IF §3.1)。
type WalletSnapshot struct {
	UserID        string
	Balance       int
	MonthlyBudget int
	UpdatedAt     time.Time
}

// DeductResult は Deduct の戻り値 (凍結 IF §3.1)。
type DeductResult struct {
	NewBalance int
	Idempotent bool
}

// ResetResult は ResetAll の戻り値 (凍結 IF §3.1)。
//
// 個別ユーザの失敗は Errors に集約して継続する (P-OBS-02)。
type ResetResult struct {
	ProcessedUsers int
	Errors         []error
}
