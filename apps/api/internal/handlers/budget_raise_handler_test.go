package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	budgetraise "github.com/ryotinjpn/goro2pay/apps/api/internal/budget_raise"
)

type mockBudgetRaiseService struct {
	recommended int
	recErr      error
	result      *budgetraise.BudgetRaiseResult
	acceptErr   error
}

func (m *mockBudgetRaiseService) ComputeRecommendedBudget(_ context.Context, _ string) (int, error) {
	return m.recommended, m.recErr
}
func (m *mockBudgetRaiseService) Accept(_ context.Context, _ string, _ int) (*budgetraise.BudgetRaiseResult, error) {
	return m.result, m.acceptErr
}

func TestBudgetRaiseHandler_GetRecommendation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name       string
		injectUser bool
		recommended int
		recErr     error
		wantStatus int
	}{
		{"正常", true, 45000, nil, http.StatusOK},
		{"ErrNoBudgetSet → 400", true, 0, budgetraise.ErrNoBudgetSet, http.StatusBadRequest},
		{"認証なし → 401", false, 0, nil, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			h := NewBudgetRaiseHandler(&mockBudgetRaiseService{recommended: tc.recommended, recErr: tc.recErr})
			r.GET("/api/budget/raise/recommendation", func(c *gin.Context) {
				if tc.injectUser {
					c.Set("userId", "u1")
				}
				h.GetRecommendation(c)
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/budget/raise/recommendation", nil))
			if w.Code != tc.wantStatus {
				t.Errorf("status want %d got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

func TestBudgetRaiseHandler_RaiseBudget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fixedResult := &budgetraise.BudgetRaiseResult{NewMonthlyBudget: 45000, AppliedFrom: time.Now()}
	cases := []struct {
		name       string
		injectUser bool
		body       string
		result     *budgetraise.BudgetRaiseResult
		acceptErr  error
		wantStatus int
	}{
		{"正常", true, `{"newMonthlyBudget":45000}`, fixedResult, nil, http.StatusOK},
		{"ErrInvalidBudget → 400", true, `{"newMonthlyBudget":0}`, nil, budgetraise.ErrInvalidBudget, http.StatusBadRequest},
		{"不正 JSON → 400", true, `{bad}`, nil, nil, http.StatusBadRequest},
		{"認証なし → 401", false, `{"newMonthlyBudget":45000}`, nil, nil, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			h := NewBudgetRaiseHandler(&mockBudgetRaiseService{result: tc.result, acceptErr: tc.acceptErr})
			r.POST("/api/budget/raise", func(c *gin.Context) {
				if tc.injectUser {
					c.Set("userId", "u1")
				}
				h.RaiseBudget(c)
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/budget/raise", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Errorf("status want %d got %d body=%s", tc.wantStatus, w.Code, strings.TrimSpace(w.Body.String()))
			}
		})
	}
}
