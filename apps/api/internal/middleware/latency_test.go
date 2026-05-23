package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatency_EmitsRequestCompleteLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	r := gin.New()
	r.Use(Latency())
	r.GET("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, buf.String())

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "request_complete", entry["event"])
	assert.Equal(t, "GET", entry["method"])
	assert.Equal(t, "/api/orders", entry["path"])
	assert.EqualValues(t, http.StatusOK, entry["statusCode"])
	_, ok := entry["latencyMs"]
	assert.True(t, ok)
}
