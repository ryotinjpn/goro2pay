package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
)

// Logout は POST /api/auth/logout の handler。
// クライアント側で Cognito GlobalSignOut を呼出済みの想定で、本 handler は
// 監査ログ目的で userId を構造化ログに記録するのみ (Functional Design F-4)。
func Logout(c *gin.Context) {
	// 認証は AttachUserID middleware で済んでいる前提。userId 不在は middleware が
	// 500 で abort しているはずだが、ハンドラ側でも防御的に 401 を返す。
	if _, err := auth.UserIDFromContext(c); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// action context attr を仕込み、構造化ログ Handler が自動的に拾う。
	// userId は AttachUserID が context に載せているため、ContextAwareSlogHandler が
	// JSON 出力時に自動付与する。明示 attrs として渡すと userId キーが二重出力に
	// なる可能性があるためここでは渡さない。
	ctx := context.WithValue(c.Request.Context(), logging.CtxKeyAction, "logout")
	slog.InfoContext(ctx, "user logout")

	c.Status(http.StatusNoContent)
}
