// Package idempotency は IdempotencyKeys テーブルへの DynamoDB アクセスを提供する。
//
// Deduct の冪等性保証 (CR-B-04 / PR-B-03 / PR-B-04) を担う。
// TTL 24h は DynamoDB TTL の自動削除に委ねる (NFR-REL-02)。
package idempotency

import "time"

// IdempotencyRecord は IdempotencyKeys テーブルの 1 行に対応する。
//
// Payload は冪等性ヒット判定 (DR-B-03) で hash 比較される。
// Response は初回処理結果のシリアライズ (成功 = newBalance / 失敗 = error 種別)。
type IdempotencyRecord struct {
	Key       string
	Payload   []byte
	Response  []byte
	CreatedAt time.Time
	ExpiresAt time.Time
}
