package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/apperrors"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
)

// requestContextHeader は Lambda Web Adapter (LWA) が API Gateway イベントの
// requestContext オブジェクトを Lambda → Gin に転送する HTTP ヘッダ名。
// LWA は body / path / queryStringParameters は通常の HTTP メッセージに展開するが、
// requestContext は JSON 文字列として `x-amzn-request-context` ヘッダに格納する
// (base64 エンコードされる場合あり)。
//
// 参考: https://github.com/awslabs/aws-lambda-web-adapter
const requestContextHeader = "x-amzn-request-context"

// AttachUserID は API Gateway Cognito JWT Authorizer が付与した
// claims から sub を抽出し、gin.Context および request.Context に
// "userId" として注入する Gin middleware。
//
// unit-interfaces.md §2.1 で公開されている契約。
// AccessToken の claims には email がないため、本 middleware では
// emailHash の生成は行わない (NFR Design P-SEC-02 サブパターン B-2、
// 認証前 endpoint で別途 handler 内で生成)。
func AttachUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := extractClaims(c)
		sub, _ := claims["sub"].(string)
		if sub == "" {
			// claims 欠落は設定ミス・バグの可能性 (R-JWT-4)
			slog.ErrorContext(c.Request.Context(), "missing sub claim, possible misconfiguration",
				"event", "auth_claims_missing")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// gin.Context と Go context.Context の両方にセット
		c.Set(string(logging.CtxKeyUserID), sub)
		ctx := context.WithValue(c.Request.Context(), logging.CtxKeyUserID, sub)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// UserIDFromContext は middleware で載せた userId を gin.Context から取り出す。
// 認証必須 handler はこれを呼んで userId を得る。
//
// 取得経路 (どちらに値があっても拾える):
//  1. gin.Context.Get(CtxKeyUserID) — handler から呼ぶ通常経路
//  2. fallback: c.Request.Context().Value(CtxKeyUserID) —
//     gin から外れた goroutine 内で c.Request.Context() を切り出した
//     非同期処理から呼ぶケース (後続 Unit C/D 想定)
//
// どちらにも値が無ければ ErrUnauthorized を返す (防御的)。
// claims 欠落で middleware が abort していれば、本関数まで到達しない。
func UserIDFromContext(c *gin.Context) (string, error) {
	// (1) gin.Context lookup
	if v, exists := c.Get(string(logging.CtxKeyUserID)); exists {
		if sub, ok := v.(string); ok && sub != "" {
			return sub, nil
		}
	}
	// (2) request.Context() fallback (middleware も両方に書いている)
	if c.Request != nil {
		if v := c.Request.Context().Value(logging.CtxKeyUserID); v != nil {
			if sub, ok := v.(string); ok && sub != "" {
				return sub, nil
			}
		}
	}
	return "", apperrors.ErrUnauthorized
}

// extractClaims は API Gateway Cognito JWT Authorizer が
// `event.requestContext.authorizer.jwt.claims` として渡す map を取り出す。
//
// 取得経路は優先順:
//  1. テスト用に gin.Context に "authorizer.claims" として直接 set されていれば
//     それを返す (本番では使わない)
//  2. LWA が転送する `x-amzn-request-context` HTTP ヘッダから JSON を parse し、
//     `authorizer.jwt.claims` を取り出す (本番経路、API Gateway HTTP API v2)
//
// LWA + API Gateway HTTP API v2 + Cognito JWT Authorizer の組合せでは、
// claims は必ず `requestContext.authorizer.jwt.claims` (map[string]string) に
// 入る。この map のキー名は Cognito の標準 (sub, email, token_use, ...) に従う。
func extractClaims(c *gin.Context) map[string]any {
	// (1) テスト経路: gin.Context に直接注入されていればそれを使う
	if v, ok := c.Get("authorizer.claims"); ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}

	// (2) 本番経路: LWA が転送する x-amzn-request-context ヘッダを parse
	raw := c.GetHeader(requestContextHeader)
	if raw == "" {
		return map[string]any{}
	}

	payload := decodeRequestContextPayload(raw)
	if payload == nil {
		return map[string]any{}
	}

	return claimsFromRequestContext(payload)
}

// decodeRequestContextPayload は LWA が x-amzn-request-context ヘッダに載せた
// requestContext JSON を取り出す。LWA の version によって base64 エンコードを
// 行う場合と行わない場合があるため、両方をハンドリングする。
func decodeRequestContextPayload(raw string) map[string]any {
	// まず生 JSON として parse を試す (LWA v0.8 系以降の挙動)
	var direct map[string]any
	if err := json.Unmarshal([]byte(raw), &direct); err == nil {
		return direct
	}

	// 失敗したら base64 デコードしてから parse (古い LWA / 互換)
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

// claimsFromRequestContext は requestContext JSON から
// authorizer.jwt.claims を取り出す。HTTP API v2 (JWT Authorizer) の形式に
// 準拠。HTTP API v1 (REST API カスタム Authorizer) の `authorizer.claims`
// 直下にも対応する (互換)。
func claimsFromRequestContext(rc map[string]any) map[string]any {
	authorizer, ok := rc["authorizer"].(map[string]any)
	if !ok {
		return map[string]any{}
	}

	// HTTP API v2 + JWT Authorizer: authorizer.jwt.claims
	if jwt, ok := authorizer["jwt"].(map[string]any); ok {
		if claims, ok := jwt["claims"].(map[string]any); ok {
			return claims
		}
	}

	// 互換: REST API + カスタム Authorizer: authorizer.claims
	if claims, ok := authorizer["claims"].(map[string]any); ok {
		return claims
	}

	return map[string]any{}
}
