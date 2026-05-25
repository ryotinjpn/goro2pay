// Package main は EventBridge Scheduler から月初に呼ばれる Lambda エントリポイント。
//
// runtime: provided.al2023 + arch=arm64、handler 名は `bootstrap`
// schedule: cron(0 15 L * ? *) UTC = 月末最終日 0:00 JST (Q-B8=A)
//
// 役割: 全アクティブユーザの Wallet.balance を BudgetSettings.monthlyBudget に
// リセットする (UC-B-05、PR-B-01 完全リセット)。
//
// 配置: `apps/api/cmd/scheduler/` 配下。`apps/api` と同一 Go module を共有することで
// `internal/` パッケージにアクセス可能になる (Go の internal package 制約)。
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_reset_log"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/wallet_repo"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/wallet"
)

// init は構造化ログ Handler を default logger に設定する (P-OBS-01、Unit A 統一)。
func init() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(logging.NewContextAwareSlogHandler(os.Stdout, level)))
}

// runReset は実処理を行うコア関数。テスト容易性のため main の handler 本体を
// 関数として切り出し、wallet.WalletService interface を引数に取れるようにする。
func runReset(ctx context.Context, svc wallet.WalletService) error {
	result, err := svc.ResetAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "monthly_reset failed",
			"action", "monthly_reset",
			"error", err.Error(),
		)
		return err
	}
	slog.InfoContext(ctx, "monthly_reset summary",
		"action", "monthly_reset",
		"processedUsers", result.ProcessedUsers,
		"failedUsers", len(result.Errors),
	)
	return nil
}

// handler は Lambda エントリポイント。EventBridge Scheduler は固定 input "{}" を渡す。
//
// Scheduler は ResetAll のみを呼ぶため idempotency repo は構築しない
// (Code Review Important 5)。Scheduler Lambda の IAM Role も idempotency_keys
// テーブルへの権限を付与しない最小権限設計と整合する。
func handler(ctx context.Context, _ json.RawMessage) error {
	walletRepo := wallet_repo.NewRepository()
	settingsRepo := budget_settings.NewRepository()
	resetLogRepo := budget_reset_log.NewRepository()

	svc := wallet.NewSchedulerService(walletRepo, settingsRepo, resetLogRepo)
	return runReset(ctx, svc)
}

func main() {
	lambda.Start(handler)
}
