// Package logging は構造化ログ用の middleware と slog Handler を提供する。
// NFR Design P-OBS-01 (context-based slog Handler) と LC-AUTH-04 (RequestContextMiddleware) に従う。
package logging

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Context key 型。任意の文字列キーは衝突するので独自型で隔離する。
type ctxKey string

const (
	CtxKeyRequestID ctxKey = "requestId"
	CtxKeyTraceID   ctxKey = "traceId"
	CtxKeyUserAgent ctxKey = "userAgent"
	CtxKeyUserID    ctxKey = "userId"
	CtxKeyEmailHash ctxKey = "email_hash" // 認証前 endpoint 限定 (P-SEC-02 / A-NFR-OBS-01 snake_case)
	CtxKeyAction    ctxKey = "action"
)

// requestContextHeader は LWA が API Gateway イベントの requestContext を
// HTTP ヘッダで Lambda → Gin に転送するためのヘッダ名。
const requestContextHeader = "x-amzn-request-context"

// RequestContext は requestId / traceId / userAgent を context に注入する Gin middleware。
// 全ハンドラより先に登録すること。
//
// requestId 取得経路の優先順:
//  1. `Apigw-Requestid` ヘッダ (API Gateway HTTP API v2 が直接付与)
//  2. `x-amzn-request-context` ヘッダ JSON の `requestContext.requestId`
//     (LWA 経由、API Gateway HTTP API v2 + REST 互換)
//  3. `X-Amzn-RequestId` ヘッダ (ALB 系の互換、HTTP API v2 では通常付与されない)
//  4. ランダム ID (フォールバック、開発・テスト時)
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := resolveRequestID(c)
		traceID := c.GetHeader("X-Amzn-Trace-Id")
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

// resolveRequestID は API Gateway / LWA / ALB 系から requestId を取り出し、
// どこにも無ければランダム ID を返す。優先順は RequestContext の doc 参照。
func resolveRequestID(c *gin.Context) string {
	if id := c.GetHeader("Apigw-Requestid"); id != "" {
		return id
	}
	if rc := c.GetHeader(requestContextHeader); rc != "" {
		if payload := decodeRequestContextPayload(rc); payload != nil {
			if id, ok := payload["requestId"].(string); ok && id != "" {
				return id
			}
		}
	}
	if id := c.GetHeader("X-Amzn-RequestId"); id != "" {
		return id
	}
	return randomHexID(16)
}

// decodeRequestContextPayload は LWA が x-amzn-request-context ヘッダに載せた
// requestContext JSON を取り出す。生 JSON / base64 エンコードの両方に対応。
//
// auth.middleware.go にも同名関数があるが、循環 import を避けるため logging 側
// にも独自実装を持つ。挙動は同一。
func decodeRequestContextPayload(raw string) map[string]any {
	var direct map[string]any
	if err := json.Unmarshal([]byte(raw), &direct); err == nil {
		return direct
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil
	}
	var fromB64 map[string]any
	if err := json.Unmarshal(decoded, &fromB64); err != nil {
		return nil
	}
	return fromB64
}

// randomHexID は crypto/rand から hex 文字列 ID を生成する。
// crypto/rand.Read が失敗した場合は時刻ベースの fallback ID を返し、
// 全リクエストで同じ ID が出続ける状況を防ぐ (固定 ID で観測ログが
// 集約不能になるのを避ける)。
func randomHexID(byteLen int) string {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand 失敗時は ERROR ログ + UnixNano fallback で fail-safe
		slog.Error("logging.randomHexID: crypto/rand.Read failed; using time fallback",
			"event", "rand_failure", "error", err)
		now := time.Now().UnixNano()
		// 8 bytes の UnixNano を hex 化したものを基準値にして長さを揃える
		base := []byte{
			byte(now >> 56), byte(now >> 48), byte(now >> 40), byte(now >> 32),
			byte(now >> 24), byte(now >> 16), byte(now >> 8), byte(now),
		}
		// byteLen <= 8 ならカット、>8 なら base を繰り返して埋める
		out := make([]byte, byteLen)
		for i := 0; i < byteLen; i++ {
			out[i] = base[i%len(base)]
		}
		return hex.EncodeToString(out)
	}
	return hex.EncodeToString(b)
}
