package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/apperrors"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
)

func TestAttachUserID_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// claims をセットする pre-middleware
	r.Use(func(c *gin.Context) {
		c.Set("authorizer.claims", map[string]any{
			"sub":   "user-sub-uuid",
			"email": "irrelevant@example.com",
		})
		c.Next()
	})
	r.Use(AttachUserID())

	r.GET("/", func(c *gin.Context) {
		userID, err := UserIDFromContext(c)
		assert.NoError(t, err)
		assert.Equal(t, "user-sub-uuid", userID)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAttachUserID_MissingClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// claims をセットしない (つまり Authorizer 失敗 / 設定ミスのシミュレーション)
	r.Use(AttachUserID())

	r.GET("/", func(c *gin.Context) {
		// abort されているのでここには来ないはず
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "claims 欠落は 500 (R-JWT-4)")
}

func TestAttachUserID_EmptySub(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authorizer.claims", map[string]any{
			"sub": "",
		})
		c.Next()
	})
	r.Use(AttachUserID())

	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserIDFromContext_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, err := UserIDFromContext(c)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
}

func TestUserIDFromContext_SetByMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(string(logging.CtxKeyUserID), "sub-from-middleware")

	got, err := UserIDFromContext(c)
	assert.NoError(t, err)
	assert.Equal(t, "sub-from-middleware", got)
}

// 本番経路: LWA が API Gateway HTTP API v2 + JWT Authorizer の
// requestContext を x-amzn-request-context ヘッダ (生 JSON) に載せて転送するパターン。
func TestAttachUserID_FromLWAHeader_RawJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AttachUserID())
	r.GET("/", func(c *gin.Context) {
		userID, err := UserIDFromContext(c)
		assert.NoError(t, err)
		assert.Equal(t, "sub-from-jwt-authorizer", userID)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-amzn-request-context",
		`{"authorizer":{"jwt":{"claims":{"sub":"sub-from-jwt-authorizer","token_use":"access"}}}}`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// LWA の古いバージョンは requestContext を base64 エンコードしてヘッダに載せる。
func TestAttachUserID_FromLWAHeader_Base64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AttachUserID())
	r.GET("/", func(c *gin.Context) {
		userID, err := UserIDFromContext(c)
		assert.NoError(t, err)
		assert.Equal(t, "sub-base64", userID)
		c.Status(http.StatusOK)
	})

	rcJSON := `{"authorizer":{"jwt":{"claims":{"sub":"sub-base64","token_use":"access"}}}}`
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-amzn-request-context",
		base64.StdEncoding.EncodeToString([]byte(rcJSON)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 互換: REST API + カスタム Authorizer の場合は authorizer.claims 直下に claims が入る。
func TestAttachUserID_FromLWAHeader_RESTAPICompat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AttachUserID())
	r.GET("/", func(c *gin.Context) {
		userID, err := UserIDFromContext(c)
		assert.NoError(t, err)
		assert.Equal(t, "sub-rest", userID)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-amzn-request-context",
		`{"authorizer":{"claims":{"sub":"sub-rest"}}}`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ヘッダもテスト経路も無い場合は claims 欠落で 500 (R-JWT-4)
func TestAttachUserID_NoHeader_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AttachUserID())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ヘッダはあるが破損している (JSON でも base64 でもない) 場合も 500
func TestAttachUserID_MalformedHeader_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AttachUserID())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-amzn-request-context", "this-is-not-json-and-not-base64-padded-properly!@#$")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// テスト経路 (gin.Context 直接 set) が LWA ヘッダより優先されることを確認
// (既存テストが LWA ヘッダなしでも動作する後方互換のため)。
func TestAttachUserID_GinContextOverridesHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authorizer.claims", map[string]any{"sub": "from-gin-context"})
		c.Next()
	})
	r.Use(AttachUserID())
	r.GET("/", func(c *gin.Context) {
		userID, err := UserIDFromContext(c)
		assert.NoError(t, err)
		assert.Equal(t, "from-gin-context", userID)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-amzn-request-context",
		`{"authorizer":{"jwt":{"claims":{"sub":"from-header"}}}}`)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
