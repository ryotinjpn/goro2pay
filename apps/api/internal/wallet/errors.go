package wallet

import "errors"

// Sentinel errors (凍結 IF unit-interfaces.md §3.1)。
var (
	// ErrInsufficientBalance は Deduct で残高不足の場合に返される。
	// Handler 層は HTTP 402 INSUFFICIENT_BALANCE にマップする。
	ErrInsufficientBalance = errors.New("wallet: insufficient balance")

	// ErrIdempotencyConflict は同一 idempotencyKey に対して異なる payload が
	// 渡された場合に返される。Handler 層は HTTP 409 IDEMPOTENCY_CONFLICT にマップ。
	ErrIdempotencyConflict = errors.New("wallet: idempotency key conflict with different payload")

	// ErrBudgetOutOfRange は SetBudget の monthlyBudget が範囲外 (1..100,000)
	// または 1,000 円刻みでない場合に返される。Handler 層は HTTP 400 にマップ。
	ErrBudgetOutOfRange = errors.New("wallet: monthly budget out of range")

	// ErrIdempotencyInProgress は IdempotencyRecord が存在するが Response が
	// 空 (= 初回 PutItem 後の SaveResponse 失敗、または初回処理が in-flight)
	// の状態で同一キーが再送された場合に返される。Handler 層は HTTP 503
	// SERVICE_UNAVAILABLE にマップし、クライアントは少し待って再試行する。
	//
	// 本エラーが空 Response 時に NewBalance:0 を返す corner case (Code Review
	// Critical 2) を防ぐ。
	ErrIdempotencyInProgress = errors.New("wallet: idempotent request still in progress")
)
