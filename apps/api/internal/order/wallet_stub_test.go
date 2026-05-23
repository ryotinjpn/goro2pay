package order

import (
	"context"
	"sync"

	"github.com/oklog/ulid/v2"
	cryptoRand "crypto/rand"
	"time"
)

// WalletStub は Unit B WalletService interface の in-memory 実装 (LC-19)。
//
// PBT (P-PBT-01) で Unit C 視点の冪等性 (P-1, P-3) を検証するために使う。
// 同一 idempotencyKey + userID に対する 2 回目以降の Deduct は初回の結果を
// そのまま返し、残高は変動させない。
type WalletStub struct {
	mu             sync.Mutex
	balances       map[string]int                  // key: userID
	idempotencyMap map[string]*WalletDeductResult  // key: userID + "|" + idempotencyKey
	nextOrderTime  time.Time                       // テスト時刻固定用
}

// NewWalletStub は初期残高を持つ WalletStub を返す。
func NewWalletStub(initialBalance int) *WalletStub {
	return &WalletStub{
		balances:       make(map[string]int),
		idempotencyMap: make(map[string]*WalletDeductResult),
		nextOrderTime:  time.Now().UTC(),
	}
}

// SetBalance はテスト用に userID の残高を直接設定する。
func (w *WalletStub) SetBalance(userID string, amount int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.balances[userID] = amount
}

// Balance はテスト assertion 用に現在残高を返す。
func (w *WalletStub) Balance(userID string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.balances[userID]
}

// Deduct は WalletService.Deduct を実装する。
//
// 同一 (userID, idempotencyKey) の 2 回目以降は初回結果を返し残高変動なし
// (Unit C P-1 / P-3 の前提となる冪等挙動を再現)。
func (w *WalletStub) Deduct(ctx context.Context, userID, idempotencyKey string, amount int) (*WalletDeductResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	key := userID + "|" + idempotencyKey
	if existing, ok := w.idempotencyMap[key]; ok {
		// 冪等命中: 初回結果のコピーを Idempotent=true で返す
		return &WalletDeductResult{
			OrderID:          existing.OrderID,
			RemainingBalance: existing.RemainingBalance,
			Idempotent:       true,
		}, nil
	}

	if w.balances[userID] < amount {
		return nil, ErrInsufficientFunds
	}
	w.balances[userID] -= amount

	t := ulid.Timestamp(w.nextOrderTime)
	entropy := ulid.Monotonic(cryptoRand.Reader, 0)
	orderID := ulid.MustNew(t, entropy).String()

	result := &WalletDeductResult{
		OrderID:          orderID,
		RemainingBalance: w.balances[userID],
		Idempotent:       false,
	}
	// 初回結果を冪等マップに保存 (2 回目以降の Idempotent=true で参照)
	stored := *result
	w.idempotencyMap[key] = &stored
	return result, nil
}
