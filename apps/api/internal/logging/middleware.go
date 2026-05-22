// Package logging は構造化ログ用の middleware と slog Handler を提供する。
// NFR Design P-OBS-01 (context-based slog Handler) と LC-AUTH-04 (RequestContextMiddleware) に従う。
package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// Context key 型。任意の文字列キーは衝突するので独自型で隔離する。
type ctxKey string

const (
	CtxKeyRequestID ctxKey = "requestId"
	CtxKeyTraceID   ctxKey = "traceId"
	CtxKeyUserAgent ctxKey = "userAgent"
	CtxKeyUserID    ctxKey = "userId"
	CtxKeyEmailHash ctxKey = "emailHash" // 認証前 endpoint 限定 (P-SEC-02)
	CtxKeyAction    ctxKey = "action"
)

// RequestContext は requestId / traceId / userAgent を context に注入する Gin middleware。
// 全ハンドラより先に登録すること。
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// requestId: API Gateway の Request ID があればそれ、なければ新規生成
		requestID := c.GetHeader("X-Amzn-RequestId")
		if requestID == "" {
			requestID = randomHexID(16)
		}

		// traceId: X-Amzn-Trace-Id ヘッダから (なければ空文字)
		traceID := c.GetHeader("X-Amzn-Trace-Id")

		// userAgent
		userAgent := c.GetHeader("User-Agent")

		// gin.Context と Go context.Context の両方に書く
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, CtxKeyRequestID, requestID)
		if traceID != "" {
			ctx = context.WithValue(ctx, CtxKeyTraceID, traceID)
		}
		ctx = context.WithValue(ctx, CtxKeyUserAgent, userAgent)
		c.Request = c.Request.WithContext(ctx)

		// gin.Context にも入れる (gin handler が cleanly 取り出せるように)
		c.Set(string(CtxKeyRequestID), requestID)
		if traceID != "" {
			c.Set(string(CtxKeyTraceID), traceID)
		}
		c.Set(string(CtxKeyUserAgent), userAgent)

		c.Next()
	}
}

func randomHexID(byteLen int) string {
	b := make([]byte, byteLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
