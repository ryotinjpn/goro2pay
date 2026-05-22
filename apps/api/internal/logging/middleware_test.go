package logging

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestContextMiddleware_GeneratesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContext())

	var capturedRequestID string
	var capturedUserAgent string
	r.GET("/", func(c *gin.Context) {
		v, _ := c.Get(string(CtxKeyRequestID))
		capturedRequestID, _ = v.(string)
		ua, _ := c.Get(string(CtxKeyUserAgent))
		capturedUserAgent, _ = ua.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedRequestID, "requestId should be generated when header missing")
	assert.Equal(t, "test-agent", capturedUserAgent)
}

func TestRequestContextMiddleware_UsesProvidedRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContext())

	var capturedRequestID string
	r.GET("/", func(c *gin.Context) {
		v, _ := c.Get(string(CtxKeyRequestID))
		capturedRequestID, _ = v.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Amzn-RequestId", "explicit-request-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "explicit-request-id", capturedRequestID)
}

func TestRequestContextMiddleware_AttachesTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContext())

	var capturedTraceID string
	r.GET("/", func(c *gin.Context) {
		v, _ := c.Get(string(CtxKeyTraceID))
		capturedTraceID, _ = v.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Amzn-Trace-Id", "Root=1-xxx")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "Root=1-xxx", capturedTraceID)
}

// 優先順 (1): API Gateway HTTP API v2 が直接付与する Apigw-Requestid ヘッダ
func TestRequestContextMiddleware_PrefersApigwRequestid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContext())

	var captured string
	r.GET("/", func(c *gin.Context) {
		v, _ := c.Get(string(CtxKeyRequestID))
		captured, _ = v.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Apigw-Requestid を最優先とし、X-Amzn-RequestId が併存しても無視されること
	req.Header.Set("Apigw-Requestid", "apigw-req-1")
	req.Header.Set("X-Amzn-RequestId", "amzn-fallback")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "apigw-req-1", captured)
}

// 優先順 (2): LWA が x-amzn-request-context ヘッダ JSON で渡す requestContext.requestId
func TestRequestContextMiddleware_FallsBackToLWAHeader_RawJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContext())

	var captured string
	r.GET("/", func(c *gin.Context) {
		v, _ := c.Get(string(CtxKeyRequestID))
		captured, _ = v.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-amzn-request-context",
		`{"requestId":"rc-req-1","authorizer":{"jwt":{"claims":{"sub":"x"}}}}`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "rc-req-1", captured)
}

// 優先順 (2) base64 エンコード形式
func TestRequestContextMiddleware_FallsBackToLWAHeader_Base64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContext())

	var captured string
	r.GET("/", func(c *gin.Context) {
		v, _ := c.Get(string(CtxKeyRequestID))
		captured, _ = v.(string)
		c.Status(http.StatusOK)
	})

	rcJSON := `{"requestId":"rc-base64-1"}`
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-amzn-request-context",
		base64.StdEncoding.EncodeToString([]byte(rcJSON)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "rc-base64-1", captured)
}
