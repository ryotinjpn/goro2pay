# Auth Unit — API Layer Generation Summary

**Date**: 2026-05-22
**Stage**: Code Generation Step 5-7 完了

## 生成ファイル

| ファイル | 役割 | 関連 LC |
|---|---|---|
| `apps/api/internal/logging/middleware.go` | RequestContextMiddleware (requestId/traceId/userAgent を context 注入) | LC-AUTH-04 |
| `apps/api/internal/logging/handler.go` | ContextAwareSlogHandler (context から attrs 抽出して JSON 出力) | LC-AUTH-05 |
| `apps/api/internal/auth/middleware.go` | AttachUserID + UserIDFromContext (sub を context に注入) | LC-AUTH-01 |
| `apps/api/internal/handlers/health.go` | GET /health (認証不要) | — |
| `apps/api/internal/handlers/logout.go` | POST /api/auth/logout (監査ログ + 204) | LC-AUTH-06 |
| `apps/api/main.go` | Gin 起動 + middleware 登録 + route 配線 | (横串) |

## 単体テスト

| テストファイル | カバレッジ |
|---|---|
| `internal/logging/middleware_test.go` | requestId 自動生成 / 既存ヘッダ尊重 / traceId 抽出 |
| `internal/logging/handler_test.go` | context attrs → JSON 出力 / 欠落キー省略 / Log Level 制御 |
| `internal/auth/middleware_test.go` | claims 正常 / 欠落 (500 R-JWT-4) / 空 sub / UserIDFromContext |
| `internal/handlers/logout_test.go` | 204 + ログ JSON 出力 / userId 欠落時 401 |

## middleware 登録順序 (main.go)

```
gin.New() →
  Use(RequestContext)          ← requestId/traceId/userAgent
  GET /health (認証不要)
  Group("/api", AttachUserID)  ← userId (sub claim から)
    POST /api/auth/logout
    (Unit B/C/D/E が後続 PR で追加)
```

## 構造化ログ出力サンプル

```json
{
  "time": "2026-05-22T07:30:00Z",
  "level": "INFO",
  "msg": "user logout",
  "userId": "user-sub-uuid",
  "userAgent": "Mozilla/5.0",
  "requestId": "abcdef1234567890",
  "action": "logout"
}
```

NFR Design A-NFR-OBS-01 の 8 項目 (level/timestamp/userId/action/traceId/requestId/email_hash/userAgent) のうち、認証必須 endpoint では emailHash を省略 (P-SEC-02 / R-JWT-2-A)。

## トレーサビリティ

- US-0-02 (JWT セッション維持) → AttachUserID
- FR-AUTH-04 (Logout) → POST /api/auth/logout
- A-NFR-OBS-01 (8 項目構造化ログ) → ContextAwareSlogHandler
- A-NFR-REL-02 (claims 欠落時 500 + ERROR ログ) → middleware の 500 + slog.ErrorContext
