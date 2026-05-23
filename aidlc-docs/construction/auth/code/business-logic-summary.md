# Auth Unit — Business Logic Generation Summary

**Date**: 2026-05-22
**Stage**: Code Generation Step 2-4 完了

## 生成ファイル

| ファイル | 行数 | 役割 |
|---|---|---|
| `apps/api/internal/auth/email.go` | ~30 | `Normalize(email string) string` / `Hash(email string) string` (SHA256 hex) |
| `apps/api/internal/auth/email_test.go` | ~110 | gopter PBT 5 プロパティ + Example-based テスト 7 件 |
| `apps/api/internal/apperrors/errors.go` | ~10 | `ErrUnauthorized` sentinel error |

## 関数シグネチャ

```go
package auth

func Normalize(email string) string  // 前後空白除去 + ToLower
func Hash(email string) string       // SHA256(Normalize(email)) を hex 64 chars

package apperrors

var ErrUnauthorized = errors.New("unauthorized")
```

## PBT プロパティ (NFR Design A-NFR-TEST-02 / P-TEST-01)

| Property | 内容 |
|---|---|
| TestPropertyNormalizeIdempotent | `Normalize(Normalize(x)) == Normalize(x)` |
| TestPropertyNormalizeCaseInsensitive | `Normalize(ToUpper(x)) == Normalize(x)` |
| TestPropertyHashLength | `len(Hash(x)) == 64`, all hex |
| TestPropertyHashNormalizeConsistency | `Hash(x) == Hash(Normalize(x))` |
| TestPropertyHashCollisionResistance | `Normalize(x) != Normalize(y) => Hash(x) != Hash(y)` |

## Example-based テスト

- Normalize: 7 ケース（lowercase/uppercase/mixed/whitespace/empty/unicode）
- Hash: 確定値整合 + 大文字小文字整合 + 異なる入力で異なる hash

## トレーサビリティ

- US-0-01 / US-0-02 の email 入力処理基盤
- A-NFR-SEC-08 (email 平文を log に出さない、hash で代替)
- A-NFR-TEST-02 (PBT を normalize / emailHash に限定適用)

## 注意事項

- AccessToken には email claim がないため、middleware では email_hash を生成しない
  (NFR Design P-SEC-02 サブパターン B-2 のみ運用)
- 認証前 endpoint (Signup / Login) で request body から email を取得 → Normalize → Hash の順で処理
