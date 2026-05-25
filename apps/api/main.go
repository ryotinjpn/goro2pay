// goro2pay API Lambda のエントリポイント。
// Gin で起動し、Lambda Web Adapter 経由で API Gateway HTTP API のリクエストを受ける。
// LWA は localhost:8080 で listen している HTTP server を Lambda invoke にブリッジする。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/bedrock"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/delivery"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/fallback"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/handlers"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/middleware"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/order"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_reset_log"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/budget_settings"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/idempotency"
	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/wallet_repo"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/wallet"
)

const shutdownTimeout = 5 * time.Second

func main() {
	// 構造化ログ Handler を default logger に設定 (NFR Design P-OBS-01)
	level := slog.LevelInfo
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(logging.NewContextAwareSlogHandler(os.Stdout, level)))

	// Gin Engine: LOG_LEVEL=debug 時は Gin の Debug 出力を活かす
	if logLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// panic を 500 として返す Recovery middleware (gin.New() にはデフォルトで含まれない)。
	// 後続 Unit B/C/D/E のハンドラ内 panic でコンテナがクラッシュするのを防ぐ。
	r.Use(gin.Recovery())

	// 全 route で構造化ログ用 context を注入
	r.Use(logging.RequestContext())

	// E2E レイテンシ計測 (NFRC-C13-1 アラーム検知用、Unit C P-OBS-01)
	// LatencyMiddleware は logging.RequestContext の後に登録することで、
	// `request_complete` イベントに traceId / requestId / userAgent が付与される。
	r.Use(middleware.Latency())

	// Unit C 依存コンポーネント (P-DI-01 手動 DI、Unit A 統一)。
	// Bedrock / DynamoDB SDK は package init() で初期化済み (P-INIT-01)。
	bedrockAdapter := bedrock.NewClaudeBedrockAdapter()
	// import cycle (bedrock → order) 回避のため bedrock package で
	// RetryReporter 型を定義し、main.go 側で order.LogBedrockRetry を bridge
	// して注入する (P-OBS-03 / NFRC-C13-2)。
	bedrockAdapter.SetRetryReporter(order.LogBedrockRetry)
	deliveryAdapter := delivery.NewMockDeliveryAdapter()
	fallbackProvider := fallback.NewSimpleFallbackProvider()
	planBuilder := order.NewBedrockPlanBuilder(bedrockAdapter, fallbackProvider)
	orderHistoryRepo := orderhistory.NewRepository()

	// Unit B WalletService 本実装 (Code Generation 完了で配線)。
	// Repository は env から DDB_TABLE_* を読み込んで SDK Client を共有する
	// (各 Repo の package-level init() で初期化済み、P-INIT-01)。
	walletRepo := wallet_repo.NewRepository()
	settingsRepo := budget_settings.NewRepository()
	idemRepo := idempotency.NewRepository()
	resetLogRepo := budget_reset_log.NewRepository()
	walletSvc := wallet.NewService(walletRepo, settingsRepo, idemRepo, resetLogRepo)

	// Unit C `order.WalletService` interface には引数順序 / 戻り値型の差異があるため
	// Adapter を経由する (wallet.OrderAdapter)。
	walletAdapter := wallet.NewOrderAdapter(walletSvc)
	orderSvc := order.NewService(orderHistoryRepo, planBuilder, deliveryAdapter, walletAdapter)
	orderHandler := handlers.NewOrderHandler(orderSvc)
	walletHandler := wallet.NewHandler(walletSvc)

	// /health は認証不要 (LWA / load balancer 用)
	r.GET("/health", handlers.Health)

	// /api/* は認証必須グループ
	api := r.Group("/api", auth.AttachUserID())
	{
		api.POST("/auth/logout", handlers.Logout)
		api.POST("/orders", orderHandler.PlaceOrder)
		api.GET("/orders", orderHandler.GetHistory)
		// Unit B (budget) ルート
		api.GET("/wallet", walletHandler.GetBalance)
		api.POST("/wallet/budget", walletHandler.SetBudget)
		// Unit D/E が後続 PR で route を追加する
	}

	// LWA は localhost:8080 を期待する
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// SIGTERM / SIGINT を受け取って graceful shutdown する。
	// LWA は AWS_LWA_GRACEFUL_SHUTDOWN を有効にすると Lambda リサイクル時に
	// SIGTERM を送るため、ここで in-flight リクエストを完了させる。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("api server starting", "addr", addr, "log_level", level.String())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		slog.Error("server failed", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining connections", "timeout", shutdownTimeout)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		slog.Info("server stopped cleanly")
	}
}
