package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/apperrors"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
)

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
// claims 欠落で middleware が abort していれば、本関数まで到達しない。
// もし呼ばれて値がなければ ErrUnauthorized を返す (防御的)。
func UserIDFromContext(c *gin.Context) (string, error) {
	v, exists := c.Get(string(logging.CtxKeyUserID))
	if !exists {
		return "", apperrors.ErrUnauthorized
	}
	sub, ok := v.(string)
	if !ok || sub == "" {
		return "", apperrors.ErrUnauthorized
	}
	return sub, nil
}

// extractClaims は API Gateway Cognito Authorizer が event.requestContext.authorizer.claims
// として渡す map を gin.Context から取り出す。LWA + API Gateway HTTP API (payload v2) の
// 場合、Lambda Adapter は claims を request header に flatten する仕様があるため、
// 環境ごとに実装が異なる可能性がある。本実装は LWA が gin.Context に直接 set する
// パスを想定し、なければ空の map を返す (テスト時に手動で c.Set できる)。
func extractClaims(c *gin.Context) map[string]any {
	if v, ok := c.Get("authorizer.claims"); ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return map[string]any{}
}
