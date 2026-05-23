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

## 環境変数の役割分担

LWA + Lambda 実行のため、env は 2 系統に分かれる:

| 種別 | 例 | 配置 | 理由 |
|---|---|---|---|
| LWA 起動時必須 | `AWS_LWA_PORT=8080` / `PORT=8080` / `READINESS_CHECK_PATH=/health` / `AWS_LWA_INVOKE_MODE=buffered` | Dockerfile (`ENV`) | アプリ全体で固定値、image layer に焼き込む |
| アプリ実行時 (環境別) | `COGNITO_USER_POOL_ID` / `COGNITO_APP_CLIENT_ID` / `LOG_LEVEL` | Lambda env (Terraform で `aws_lambda_function.api.environment.variables` に設定) | 環境ごとに値が異なる / 運用で変えたい |
| デプロイごと | `BUILD_SHA` | Lambda env (CodeBuild が update-function-code 時に追加) | health response に反映 (M-3) |
