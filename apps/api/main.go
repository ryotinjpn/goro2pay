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

	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/handlers"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
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

	// /health は認証不要 (LWA / load balancer 用)
	r.GET("/health", handlers.Health)

	// /api/* は認証必須グループ
	api := r.Group("/api", auth.AttachUserID())
	{
		api.POST("/auth/logout", handlers.Logout)
		// Unit B/C/D/E が後続 PR で route を追加する
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
