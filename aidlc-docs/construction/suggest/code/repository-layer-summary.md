# Unit D — Repository Layer Summary

**Package**: `apps/api/internal/repo/suggestion/`（LC-SUGGEST-04）

## コンポーネント

| ファイル | 内容 |
|---|---|
| `repository.go` | `SuggestionStore` interface + DynamoDB `Repository`（Save=PutItem / Get=GetItem）。`Plan` / `SuggestionRecord` 型。env `DDB_TABLE_SUGGESTION`、`init()` で SDK client 初期化（P-INIT-01） |
| `inmemory.go` | `InmemoryStore`（test-only、TTL 失効を時刻関数で模倣） |
| `repository_test.go` | mock DynamoDB で Save→Get round-trip（plan JSON）+ 不在時 nil |

## スキーマ（凍結契約 §5.3）
- PK `suggestionId`、TTL `expiresAt`(30分)、属性 `userId` / `plan`(JSON 文字列) / `createdAt`
- GSI なし（suggestionId 単独 Get で完結）
- `plan` は `json.Marshal` で 1 属性に格納、Get 時 `json.Unmarshal`

## 設計
- import cycle 回避のため repo は独自 `Plan` 型を持ち、suggest package が `SuggestionPlan` と相互変換（order ⇄ orderhistory と同方式）
