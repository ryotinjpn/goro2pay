# apps/api — ゴロゴロPay API Lambda

API Lambda 本体（Go + Gin + Lambda Web Adapter）。

## ローカル開発

```bash
cd apps/api
go mod download
go run ./...                # Gin が :8080 で待ち受け
curl http://localhost:8080/health
```

## ビルド (arm64)

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/bootstrap .
```

## Docker build

```bash
docker buildx build --platform linux/arm64 -t gp-dev-api-image:local -f Dockerfile .
```

## ディレクトリ

- `internal/auth/`     — AuthMiddleware (AccessToken claims), EmailNormalizer, EmailHasher
- `internal/logging/`  — RequestContextMiddleware, ContextAwareSlogHandler
- `internal/handlers/` — HTTP handlers (health, logout)
- `internal/apperrors/` — sentinel errors

## テスト

```bash
go test ./...
```

PBT は `internal/auth/email_test.go` の `TestProperty*` 配下。

## 配置

Inception unit-of-work.md §4.1 の Go ソース配置 `apps/api/`。
