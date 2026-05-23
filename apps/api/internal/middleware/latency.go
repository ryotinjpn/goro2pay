// Package middleware は API Lambda 共通の Gin middleware を提供する。
//
// LatencyMiddleware は P-OBS-01 (Latency Measurement) の E2E 計測担当で、
// 全 API リクエストに `request_complete` ログを書き出し、CloudWatch メトリクス
// フィルタ (NFRC-C13-1) で p95 集計可能にする。
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Latency は Gin handler chain の最初に登録する middleware を返す。
//
// すべての route について処理開始から終了までの所要時間を ms 単位で計測し、
// `request_complete` イベントとして slog.InfoContext で出力する。
//
// CloudWatch Logs metric filter (`{ $.event = "request_complete" && $.latencyMs = * }`)
// で抽出 → p95 集計 → NFRC-C13-1 アラーム発動の検知導線を成立させる。
func Latency() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		elapsed := time.Since(started).Milliseconds()
		slog.InfoContext(c.Request.Context(), "request_complete",
			"event", "request_complete",
			"latencyMs", elapsed,
			"statusCode", c.Writer.Status(),
			"method", c.Request.Method,
			"path", c.FullPath(),
		)
	}
}
