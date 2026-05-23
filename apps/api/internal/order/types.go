// Package order は Unit C 中核の代行手配オーケストレーションを提供する。
package order

import "errors"

// PlaceOrderRequest は OrderService.PlaceOrder の引数 (凍結契約 §4.1)。
type PlaceOrderRequest struct {
	Category       string
	IdempotencyKey string
	SuggestionID   *string // Suggest 経由の 1 タップ注文時のみ
}

// Source は計画の出処を識別する enum (NFRC-C12 ログ項目 source)。
type Source string

const (
	SourceButton     Source = "button"
	SourceSuggestion Source = "suggestion"
)

// PlaceOrderResult は OrderService.PlaceOrder の戻り値 (凍結契約 §4.1)。
type PlaceOrderResult struct {
	OrderID          string
	StoreName        string
	MenuName         string
	Amount           int
	RemainingBalance int
	Idempotent       bool
}

// OrderRecord は履歴テーブルの 1 行 (凍結契約 §4.1 + Unit C 内部追加属性)。
type OrderRecord struct {
	OrderID        string
	UserID         string
	Category       string
	StoreName      string
	MenuName       string
	Amount         int
	OrderedAt      string
	IdempotencyKey string
	DayOfWeek      string
	Source         string
}

// Sentinel errors (凍結契約 §4.3)。
var (
	// ErrInsufficientFunds は WalletService.Deduct から伝播する残高不足エラー。
	// Unit C は HTTP 402 にマッピングする (BR-C39 / NFRC-C22)。
	ErrInsufficientFunds = errors.New("order: insufficient funds")

	// ErrIdempotencyConflict は冪等キー衝突エラー。
	// 同一 idempotencyKey に対し異なる payload で再送された場合に発生する。
	// Unit C は HTTP 409 にマッピングする (BR-C39)。
	ErrIdempotencyConflict = errors.New("order: idempotency conflict")
)
