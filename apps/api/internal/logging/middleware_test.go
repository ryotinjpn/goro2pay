package logging

import (
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
