# Unit B (`budget`) Code Generation 成果物

**生成完了日**: 2026-05-25
**Plan**: [budget-code-generation-plan.md](../../plans/budget-code-generation-plan.md)
**Status**: 全 23 Step 完了、Build & Sanity 全 PASS

## 成果物サマリ

| 領域 | ファイル数 | 主要成果物 |
|---|---|---|
| Backend (apps/api) | 22 | `internal/wallet/` 全 9 + Repository 4 種 (各 2-3) + cmd/scheduler 4 + main.go 改 |
| Frontend (web) | 11 | hooks 2 + state 1 + components/budget 3 + page 1 + lib/api 1 + tests 4 + useOrder.ts 改 + page.tsx 改 |
| Infrastructure (infra) | 14 | modules/budget 全 8 + tests 4 + api_gateway/routes.tf 改 + envs/dev/main.tf 改 + lambda_api 改 |
| Documentation (aidlc-docs) | 5 | このREADME + 3 サマリ + deployment-runbook |

## 詳細サマリ

- [backend-summary.md](./backend-summary.md) — Service / Repository / Handler / Adapter / Scheduler の実装詳細
- [frontend-summary.md](./frontend-summary.md) — Hooks / Components / State の実装詳細
- [infrastructure-summary.md](./infrastructure-summary.md) — Terraform module / tftest / 既存 module への注入
- [deployment-runbook.md](./deployment-runbook.md) — デプロイ手順 + 動作確認

## 検証結果

| 項目 | 結果 |
|---|---|
| `cd apps/api && go build ./...` | PASS |
| `cd apps/api && go test ./... -count=1` | 17 packages PASS (新規 wallet パッケージ 31 テスト + 既存テスト) |
| `cd apps/api && go vet ./...` | PASS |
| `cd apps/api/cmd/scheduler && make build` | PASS (bootstrap 14MB) |
| `cd web && npx vitest run` | 75 tests PASS (新規 31 + 既存 44) |
| `cd web && npx tsc --noEmit` | PASS |
| `cd infra/modules/budget && terraform validate && terraform test` | 11 PASS |
| `cd infra/modules/lambda_api && terraform test` | 既存 4 PASS (regression なし) |
| `cd infra/modules/api_gateway && terraform test` | 既存 5 PASS (regression なし) |
| `cd infra/envs/dev && terraform init -backend=false && terraform validate` | PASS |

## 主要設計判断

| 判断 | 理由 |
|---|---|
| `apps/scheduler/` ではなく **`apps/api/cmd/scheduler/`** に配置 | Go の `internal/` package アクセス制限のため、`apps/api` と同一 module にする必要があった |
| `wallet_repo` の sentinel `ErrInsufficientBalance` を repo 側に定義し、Service 層で `wallet.ErrInsufficientBalance` に変換 | `wallet` ↔ `wallet_repo` の循環 import 回避 |
| `useWallet` は凍結 IF §9 の 4 フィールドのみ公開、`BalanceDisplay` は内部で `useQuery({queryKey:['balance']})` を直接 subscribe | 凍結 IF を破らず `isFetching` / `isError` 等を扱うため (FD §3.1 注通り) |
| `useOrder.ts` への 402 → `setInsufficientBalance(true)` の最小追記、`mapOrderError` の `/budget-empty` ナビゲートは保持 | 既存 Unit C 挙動を破壊せずに P-DEG-02 を併設 |
| tftest IAM Policy 検証は naming のみに簡略化 | `jsonencode` 結果は plan 時 unknown のため `strcontains` が apply 必須となり mock_provider と非互換 |
| Unit C の env 名 `ORDER_HISTORY_TABLE_NAME` と Unit B の env 名 `DDB_TABLE_*` を併存 | Unit C は既存実装のため触らず、Unit B は凍結 IF §10 通りの命名を採用 |
