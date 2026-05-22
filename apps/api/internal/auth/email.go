// Package auth はメールアドレスの正規化とハッシュ化、認証 middleware を提供する。
// Functional Design business-rules.md §1 R-Email-3 (大文字小文字不問) と
// NFR Design P-SEC-02 (email_hash 認証前 endpoint 限定) に従う。
package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Normalize はメールアドレスを正規化する。
// - 前後の空白を除去
// - すべて小文字に変換
//
// プロパティ:
//   - べき等性: Normalize(Normalize(x)) == Normalize(x)
//   - 大文字小文字不問: Normalize(strings.ToUpper(x)) == Normalize(x)
func Normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Hash は正規化したメールアドレスの SHA256 hex 文字列を返す。
// ログのトレーサビリティ用 (A-NFR-SEC-08)。Salt は使わない (本 MVP 範囲)。
//
// プロパティ:
//   - 長さ不変: len(Hash(x)) == 64
//   - 正規化整合: Hash(x) == Hash(Normalize(x))
//   - 衝突なし: x != y のとき Hash(x) != Hash(y) (高確率、SHA256 の性質)
func Hash(email string) string {
	normalized := Normalize(email)
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
