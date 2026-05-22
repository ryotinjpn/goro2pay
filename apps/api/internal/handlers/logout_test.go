package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
)

func TestLogout_HappyPath(t *testing.T) {
	// 構造化ログを buffer に切替
	var buf bytes.Buffer
	prevDefault := slog.Default()
	slog.SetDefault(slog.New(logging.NewContextAwareSlogHandler(&buf, slog.LevelInfo)))
	t.Cleanup(func() { slog.SetDefault(prevDefault) })

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(logging.RequestContext())
	// claims をセット
	r.Use(func(c *gin.Context) {
		c.Set("authorizer.claims", map[string]any{
			"sub": "user-sub-1",
		})
		c.Next()
	})
	r.Use(auth.AttachUserID())
	r.POST("/api/auth/logout", Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// ログ JSON を検証
	var entry map[string]any
	err := json.Unmarshal(buf.Bytes(), &entry)
	assert.NoError(t, err, "ログが JSON 形式で出力されること")
	assert.Equal(t, "user logout", entry["msg"])
	assert.Equal(t, "logout", entry["action"])
	assert.Equal(t, "user-sub-1", entry["userId"])
}

func TestLogout_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// auth middleware を通さずに logout handler を呼ぶ (防御的テスト)
	r.POST("/api/auth/logout", Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
