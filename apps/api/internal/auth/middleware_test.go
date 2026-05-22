package auth

import (
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
