# scheduler — 月初リセット Lambda

Unit B (`budget` / ダメ予算) の月初リセット用 Lambda。EventBridge Scheduler から呼ばれ、全アクティブユーザの `Wallet.balance` を `BudgetSettings.monthlyBudget` で完全リセットする (PR-B-01)。

## 配置

`apps/api/cmd/scheduler/` 配下。**`apps/api` と同一 Go module を共有** することで Unit B の `internal/` パッケージ (`internal/wallet`, `internal/repo/...`, `internal/logging`) にアクセスできる。

## 役割

- 入力: EventBridge Scheduler (cron `0 15 L * ? *` UTC = 月末最終日 0:00 JST)
- 動作: `wallet.Service.ResetAll(ctx)` を呼び出し、全ユーザの残高を予算額に再設定
- 出力: CloudWatch Logs に処理サマリ (`processedUsers`, `failedUsers`)

## ビルド手順

terraform apply 前にローカルで `bootstrap` バイナリを生成する必要がある (Q-I2=A: zip 方式、Q-I3=A: CI/CD なし):

```bash
make -C apps/api/cmd/scheduler build
```

`bootstrap` バイナリ (Linux arm64) が `apps/api/cmd/scheduler/bootstrap` に生成される。

## デプロイ前提

1. `make build` を実行して `bootstrap` を生成
2. `cd infra/envs/dev && terraform apply` を実行
3. Terraform の `archive_file` が `bootstrap` を zip 化して Lambda にデプロイ

`bootstrap` が無いと `terraform apply` は失敗する。

## 凍結 IF / 設計成果物

- 凍結 IF: [unit-interfaces.md §3.5](../../../../aidlc-docs/construction/interfaces/unit-interfaces.md)
- Functional Design: [UC-B-05](../../../../aidlc-docs/construction/budget/functional-design/business-logic-model.md)
- NFR Design: [P-OBS-02 ResetAll エラーログ](../../../../aidlc-docs/construction/budget/nfr-design/nfr-design-patterns.md)
- Infrastructure Design: [§3.2 Scheduler Lambda](../../../../aidlc-docs/construction/budget/infrastructure-design/infrastructure-design.md)
