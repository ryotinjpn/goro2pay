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
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// action context attr を仕込み、構造化ログ Handler が自動的に拾う
	ctx := context.WithValue(c.Request.Context(), logging.CtxKeyAction, "logout")
	slog.InfoContext(ctx, "user logout", "userId", userID)

	c.Status(http.StatusNoContent)
}
