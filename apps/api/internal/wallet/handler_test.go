package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/apperrors"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/logging"
)

// mockSvc は WalletService の関数型 mock。
type mockSvc struct {
	getBalanceFn func(ctx context.Context, userID string) (*WalletSnapshot, error)
	setBudgetFn  func(ctx context.Context, userID string, monthlyBudget int) error
	deductFn     func(ctx context.Context, userID string, amount int, idempotencyKey string) (*DeductResult, error)
	resetAllFn   func(ctx context.Context) (*ResetResult, error)
}

func (m *mockSvc) GetBalance(ctx context.Context, userID string) (*WalletSnapshot, error) {
	return m.getBalanceFn(ctx, userID)
}
func (m *mockSvc) SetBudget(ctx context.Context, userID string, monthlyBudget int) error {
	return m.setBudgetFn(ctx, userID, monthlyBudget)
}
func (m *mockSvc) Deduct(ctx context.Context, userID string, amount int, key string) (*DeductResult, error) {
	return m.deductFn(ctx, userID, amount, key)
}
func (m *mockSvc) ResetAll(ctx context.Context) (*ResetResult, error) {
	return m.resetAllFn(ctx)
}

func setupRouter(svc WalletService, withUserID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc)
	// 簡易 middleware: テストで userID を context に注入
	r.Use(func(c *gin.Context) {
		if withUserID != "" {
			c.Set(string(logging.CtxKeyUserID), withUserID)
		}
		c.Next()
	})
	r.GET("/api/wallet", h.GetBalance)
	r.POST("/api/wallet/budget", h.SetBudget)
	return r
}

func TestHandler_GetBalance_OK(t *testing.T) {
	svc := &mockSvc{
		getBalanceFn: func(ctx context.Context, userID string) (*WalletSnapshot, error) {
			return &WalletSnapshot{
				UserID: userID, Balance: 28800, MonthlyBudget: 30000,
				UpdatedAt: time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	r := setupRouter(svc, "user_a")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/wallet", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	var got walletResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Balance != 28800 || got.MonthlyBudget != 30000 {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestHandler_GetBalance_NotFound(t *testing.T) {
	svc := &mockSvc{
		getBalanceFn: func(ctx context.Context, userID string) (*WalletSnapshot, error) {
			return nil, nil
		},
	}
	r := setupRouter(svc, "user_a")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/wallet", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_GetBalance_Unauthorized(t *testing.T) {
	svc := &mockSvc{
		getBalanceFn: func(ctx context.Context, userID string) (*WalletSnapshot, error) {
			return nil, nil
		},
	}
	r := setupRouter(svc, "") // userID を context にセットしない
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/wallet", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandler_SetBudget_OK(t *testing.T) {
	svc := &mockSvc{
		setBudgetFn: func(ctx context.Context, userID string, monthlyBudget int) error {
			if monthlyBudget != 30000 {
				t.Fatalf("expected 30000, got %d", monthlyBudget)
			}
			return nil
		},
	}
	r := setupRouter(svc, "user_a")
	body, _ := json.Marshal(setBudgetRequestDTO{MonthlyBudget: 30000})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/wallet/budget", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	var got setBudgetResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.MonthlyBudget != 30000 {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestHandler_SetBudget_BadRequest(t *testing.T) {
	svc := &mockSvc{}
	r := setupRouter(svc, "user_a")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/wallet/budget", bytes.NewReader([]byte(`{"invalid":"json`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_SetBudget_OutOfRange(t *testing.T) {
	svc := &mockSvc{
		setBudgetFn: func(ctx context.Context, userID string, monthlyBudget int) error {
			return ErrBudgetOutOfRange
		},
	}
	r := setupRouter(svc, "user_a")
	body, _ := json.Marshal(setBudgetRequestDTO{MonthlyBudget: 999999})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/wallet/budget", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_SetBudget_Unauthorized(t *testing.T) {
	svc := &mockSvc{}
	r := setupRouter(svc, "")
	body, _ := json.Marshal(setBudgetRequestDTO{MonthlyBudget: 30000})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/wallet/budget", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// apperrors パッケージの ErrUnauthorized が import される確認
var _ = apperrors.ErrUnauthorized
