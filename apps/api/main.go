// goro2pay API Lambda のエントリポイント。
// Gin で起動し、Lambda Web Adapter 経由で API Gateway HTTP API のリクエストを受ける。
// LWA は localhost:8080 で listen している HTTP server を Lambda invoke にブリッジする。
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/handlers"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
)

func main() {
	// 構造化ログ Handler を default logger に設定 (NFR Design P-OBS-01)
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(logging.NewContextAwareSlogHandler(os.Stdout, level)))

	// Gin Engine
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

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
	slog.Info("api server starting", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
