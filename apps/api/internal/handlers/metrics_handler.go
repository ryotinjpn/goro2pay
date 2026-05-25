package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/metrics"
)

// MetricsHandler は GET /api/metrics を処理する Gin handler。
type MetricsHandler struct {
	svc metrics.MetricsService
}

// NewMetricsHandler は MetricsHandler を返す。
func NewMetricsHandler(svc metrics.MetricsService) *MetricsHandler {
	return &MetricsHandler{svc: svc}
}

type metricsResponseDTO struct {
	DamageCount       int     `json:"damageCount"`
	ConsumptionRate   float64 `json:"consumptionRate"`
	MonthlyBudget     int     `json:"monthlyBudget"`
	RemainingBalance  int     `json:"remainingBalance"`
	ThresholdExceeded bool    `json:"thresholdExceeded"`
	SummaryText       string  `json:"summaryText"`
}

// GetMetrics は GET /api/metrics ハンドラ (凍結契約 §6.2)。
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponseDTO{Error: errorDetail{Code: "UNAUTHORIZED", Message: "認証が必要です"}})
		return
	}

	m, err := h.svc.GetMetrics(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, metrics.ErrNoBudgetSet) {
			c.JSON(http.StatusBadRequest, errorResponseDTO{Error: errorDetail{Code: "ERR_NO_BUDGET_SET", Message: "予算が設定されていません"}})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponseDTO{Error: errorDetail{Code: "INTERNAL_ERROR", Message: "内部エラー"}})
		return
	}

	c.JSON(http.StatusOK, metricsResponseDTO{
		DamageCount:       m.DamageCount,
		ConsumptionRate:   m.ConsumptionRate,
		MonthlyBudget:     m.MonthlyBudget,
		RemainingBalance:  m.RemainingBalance,
		ThresholdExceeded: m.ThresholdExceeded,
		SummaryText:       m.SummaryText,
	})
}
