package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/metrics"
)

type mockMetricsService struct {
	result *metrics.Metrics
	err    error
}

func (m *mockMetricsService) GetMetrics(_ context.Context, _ string) (*metrics.Metrics, error) {
	return m.result, m.err
}

func TestMetricsHandler_GetMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		injectUser bool
		svcResult  *metrics.Metrics
		svcErr     error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "正常",
			injectUser: true,
			svcResult: &metrics.Metrics{
				DamageCount: 5, ConsumptionRate: 0.5,
				MonthlyBudget: 30000, RemainingBalance: 15000,
				ThresholdExceeded: false, SummaryText: "今月のダメ化回数: 5 回、消化額 ¥15,000",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "ErrNoBudgetSet → 400",
			injectUser: true,
			svcErr:     metrics.ErrNoBudgetSet,
			wantStatus: http.StatusBadRequest,
			wantCode:   "ERR_NO_BUDGET_SET",
		},
		{
			name:       "内部エラー → 500",
			injectUser: true,
			svcErr:     context.DeadlineExceeded,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
		{
			name:       "認証なし → 401",
			injectUser: false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			svc := &mockMetricsService{result: tc.svcResult, err: tc.svcErr}
			h := NewMetricsHandler(svc)

			r.GET("/api/metrics", func(c *gin.Context) {
				if tc.injectUser {
					c.Set("userId", "u1")
				}
				h.GetMetrics(c)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status want %d got %d", tc.wantStatus, w.Code)
			}
			if tc.wantCode != "" {
				var body map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &body)
				errObj, _ := body["error"].(map[string]interface{})
				code, _ := errObj["code"].(string)
				if code != tc.wantCode {
					t.Errorf("code want %q got %q", tc.wantCode, code)
				}
			}
		})
	}
}
