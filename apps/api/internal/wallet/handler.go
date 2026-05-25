package wallet

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/apperrors"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
)

// Handler は Unit B の HTTP Handler (LC-BUDGET-01)。
//
// エンドポイント:
//   - GET /api/wallet — 残高 + 月間予算取得
//   - POST /api/wallet/budget — 月間予算設定 (初回 / 変更)
type Handler struct {
	svc WalletService
	now func() time.Time // テスト容易性のため注入可能 (Code Review Minor 9)
}

// NewHandler は Handler を返す。
func NewHandler(svc WalletService) *Handler {
	return &Handler{svc: svc, now: time.Now}
}

// SetClock はテスト用に時刻関数を上書きする。
func (h *Handler) SetClock(now func() time.Time) {
	h.now = now
}

// errorResponseDTO は標準エラー応答 (凍結 IF §8 と整合)。
type errorResponseDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// walletResponseDTO は GET /api/wallet の 200 応答ボディ (凍結 IF §3.3)。
type walletResponseDTO struct {
	Balance       int    `json:"balance"`
	MonthlyBudget int    `json:"monthlyBudget"`
	UpdatedAt     string `json:"updatedAt"`
}

// setBudgetRequestDTO は POST /api/wallet/budget の JSON ペイロード。
type setBudgetRequestDTO struct {
	MonthlyBudget int `json:"monthlyBudget" binding:"required"`
}

// setBudgetResponseDTO は POST /api/wallet/budget の 200 応答ボディ (凍結 IF §3.3)。
type setBudgetResponseDTO struct {
	MonthlyBudget int    `json:"monthlyBudget"`
	AppliedFrom   string `json:"appliedFrom"`
}

// GetBalance は GET /api/wallet ハンドラ。
//
// 残高 + 月間予算を JSON で返す。Wallet 未作成時は 404 (Frontend 側で /budget へ
// リダイレクトする判定材料になる)。
func (h *Handler) GetBalance(c *gin.Context) {
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です")
		return
	}

	snap, err := h.svc.GetBalance(c.Request.Context(), userID)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	if snap == nil {
		writeError(c, http.StatusNotFound, "WALLET_NOT_FOUND", "予算が未設定です")
		return
	}

	slog.InfoContext(c.Request.Context(), "wallet_get_balance",
		"action", "get_balance",
		"newBalance", snap.Balance,
	)

	c.JSON(http.StatusOK, walletResponseDTO{
		Balance:       snap.Balance,
		MonthlyBudget: snap.MonthlyBudget,
		UpdatedAt:     snap.UpdatedAt.Format(time.RFC3339),
	})
}

// SetBudget は POST /api/wallet/budget ハンドラ。
//
// 初回 / 変更を統合的に処理する。即時反映 (PR-B-02)。
func (h *Handler) SetBudget(c *gin.Context) {
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です")
		return
	}

	var dto setBudgetRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	if err := h.svc.SetBudget(c.Request.Context(), userID, dto.MonthlyBudget); err != nil {
		mapServiceError(c, err)
		return
	}

	now := h.now().UTC()
	slog.InfoContext(c.Request.Context(), "wallet_set_budget",
		"action", "set_budget",
		"amount", dto.MonthlyBudget,
	)

	c.JSON(http.StatusOK, setBudgetResponseDTO{
		MonthlyBudget: dto.MonthlyBudget,
		AppliedFrom:   now.Format(time.RFC3339),
	})
}

// mapServiceError は WalletService エラーを HTTP ステータスにマップする (凍結 IF §8)。
func mapServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrBudgetOutOfRange):
		writeError(c, http.StatusBadRequest, "VALIDATION_FAILED", "予算は 1,000〜100,000 円 (1,000 円刻み) で設定してください")
	case errors.Is(err, ErrInsufficientBalance):
		writeError(c, http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "今月のダメ予算を使い切りました")
	case errors.Is(err, ErrIdempotencyConflict):
		writeError(c, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "重複リクエストです")
	case errors.Is(err, ErrIdempotencyInProgress):
		writeError(c, http.StatusServiceUnavailable, "IDEMPOTENCY_IN_PROGRESS", "処理中です。少し待って再試行してください")
	case errors.Is(err, ErrInvalidInput):
		writeError(c, http.StatusBadRequest, "VALIDATION_FAILED", "入力が不正です")
	case errors.Is(err, apperrors.ErrUnauthorized):
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です")
	default:
		slog.ErrorContext(c.Request.Context(), "wallet_internal_error", "error", err.Error())
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバエラーが発生しました")
	}
}

func writeError(c *gin.Context, status int, code, msg string) {
	c.JSON(status, errorResponseDTO{Code: code, Message: msg})
}
