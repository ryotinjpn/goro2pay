package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	budgetraise "github.com/ryotinjpn/goro2pay/apps/api/internal/budget_raise"
)

// BudgetRaiseHandler は GET /api/budget/raise/recommendation と POST /api/budget/raise を処理する。
type BudgetRaiseHandler struct {
	svc budgetraise.BudgetRaiseService
}

// NewBudgetRaiseHandler は BudgetRaiseHandler を返す。
func NewBudgetRaiseHandler(svc budgetraise.BudgetRaiseService) *BudgetRaiseHandler {
	return &BudgetRaiseHandler{svc: svc}
}

type recommendationResponseDTO struct {
	CurrentMonthlyBudget     int `json:"currentMonthlyBudget"`
	RecommendedMonthlyBudget int `json:"recommendedMonthlyBudget"`
}

type budgetRaiseRequestDTO struct {
	NewMonthlyBudget int `json:"newMonthlyBudget" binding:"required"`
}

type budgetRaiseResponseDTO struct {
	NewMonthlyBudget int    `json:"newMonthlyBudget"`
	AppliedFrom      string `json:"appliedFrom"` // RFC3339
}

// GetRecommendation は GET /api/budget/raise/recommendation ハンドラ (凍結契約 §6.2)。
func (h *BudgetRaiseHandler) GetRecommendation(c *gin.Context) {
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponseDTO{Error: errorDetail{Code: "UNAUTHORIZED", Message: "認証が必要です"}})
		return
	}

	recommended, err := h.svc.ComputeRecommendedBudget(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, budgetraise.ErrNoBudgetSet) {
			c.JSON(http.StatusBadRequest, errorResponseDTO{Error: errorDetail{Code: "ERR_NO_BUDGET_SET", Message: "予算が設定されていません"}})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponseDTO{Error: errorDetail{Code: "INTERNAL_ERROR", Message: "内部エラー"}})
		return
	}

	// currentMonthlyBudget を別途取得する代わりに recommended から逆算はしない。
	// シンプルに recommended のみ返し、current は frontend が useMetrics から取得。
	c.JSON(http.StatusOK, recommendationResponseDTO{
		CurrentMonthlyBudget:     0, // frontend は useMetrics.monthlyBudget を参照
		RecommendedMonthlyBudget: recommended,
	})
}

// RaiseBudget は POST /api/budget/raise ハンドラ (凍結契約 §6.2)。
func (h *BudgetRaiseHandler) RaiseBudget(c *gin.Context) {
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponseDTO{Error: errorDetail{Code: "UNAUTHORIZED", Message: "認証が必要です"}})
		return
	}

	var req budgetRaiseRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponseDTO{Error: errorDetail{Code: "INVALID_REQUEST", Message: "リクエストが不正です"}})
		return
	}

	result, err := h.svc.Accept(c.Request.Context(), userID, req.NewMonthlyBudget)
	if err != nil {
		if errors.Is(err, budgetraise.ErrInvalidBudget) {
			c.JSON(http.StatusBadRequest, errorResponseDTO{Error: errorDetail{Code: "INVALID_BUDGET", Message: "予算は 1 〜 100,000 円の範囲で指定してください"}})
			return
		}
		if errors.Is(err, budgetraise.ErrNoBudgetSet) {
			c.JSON(http.StatusBadRequest, errorResponseDTO{Error: errorDetail{Code: "ERR_NO_BUDGET_SET", Message: "予算が設定されていません"}})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponseDTO{Error: errorDetail{Code: "INTERNAL_ERROR", Message: "内部エラー"}})
		return
	}

	c.JSON(http.StatusOK, budgetRaiseResponseDTO{
		NewMonthlyBudget: result.NewMonthlyBudget,
		AppliedFrom:      result.AppliedFrom.Format("2006-01-02T15:04:05Z07:00"),
	})
}
