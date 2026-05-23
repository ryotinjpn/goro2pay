package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/order"
)

// OrderHandler は POST /api/orders / GET /api/orders を処理する Gin handler。
//
// エラーマッピングは凍結契約 §4.3 に従う:
//   - order.ErrInsufficientFunds -> 402 INSUFFICIENT_FUNDS
//   - order.ErrIdempotencyConflict -> 409 IDEMPOTENCY_CONFLICT (BR-C39)
//   - context.DeadlineExceeded -> 504 GATEWAY_TIMEOUT
//   - その他 -> 500 INTERNAL_ERROR
type OrderHandler struct {
	svc order.OrderService
}

// NewOrderHandler は OrderHandler を返す。
func NewOrderHandler(svc order.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// placeOrderRequestDTO は POST /api/orders の JSON ペイロード。
type placeOrderRequestDTO struct {
	Category       string  `json:"category" binding:"required"`
	IdempotencyKey string  `json:"idempotencyKey" binding:"required"`
	SuggestionID   *string `json:"suggestionId,omitempty"`
}

// placeOrderResponseDTO は 200 応答ボディ。
type placeOrderResponseDTO struct {
	OrderID          string `json:"orderId"`
	StoreName        string `json:"storeName"`
	MenuName         string `json:"menuName"`
	Amount           int    `json:"amount"`
	RemainingBalance int    `json:"remainingBalance"`
	Idempotent       bool   `json:"idempotent"`
}

// errorResponseDTO は標準エラー応答 (凍結契約 §4.3)。
type errorResponseDTO struct {
	Error errorDetail `json:"error"`
}
type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PlaceOrder は POST /api/orders ハンドラ。
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponseDTO{Error: errorDetail{Code: "UNAUTHORIZED", Message: "認証が必要です"}})
		return
	}

	var dto placeOrderRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, errorResponseDTO{Error: errorDetail{Code: "BAD_REQUEST", Message: err.Error()}})
		return
	}

	res, err := h.svc.PlaceOrder(c.Request.Context(), userID, order.PlaceOrderRequest{
		Category:       dto.Category,
		IdempotencyKey: dto.IdempotencyKey,
		SuggestionID:   dto.SuggestionID,
	})
	if err != nil {
		mapOrderError(c, err)
		return
	}

	c.JSON(http.StatusOK, placeOrderResponseDTO{
		OrderID:          res.OrderID,
		StoreName:        res.StoreName,
		MenuName:         res.MenuName,
		Amount:           res.Amount,
		RemainingBalance: res.RemainingBalance,
		Idempotent:       res.Idempotent,
	})
}

// orderRecordDTO は GET /api/orders の 1 件分。
type orderRecordDTO struct {
	OrderID   string `json:"orderId"`
	Category  string `json:"category"`
	StoreName string `json:"storeName"`
	MenuName  string `json:"menuName"`
	Amount    int    `json:"amount"`
	OrderedAt string `json:"orderedAt"`
}

// getHistoryResponseDTO は GET /api/orders の 200 応答ボディ。
type getHistoryResponseDTO struct {
	Items []orderRecordDTO `json:"items"`
}

// GetHistory は GET /api/orders ハンドラ。
func (h *OrderHandler) GetHistory(c *gin.Context) {
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponseDTO{Error: errorDetail{Code: "UNAUTHORIZED", Message: "認証が必要です"}})
		return
	}

	limit := 20
	// クエリパラメータ ?limit=N に対応 (FD Q-11=A)
	if v := c.Query("limit"); v != "" {
		if n, perr := parsePositiveInt(v); perr == nil {
			limit = n
		}
	}

	recs, err := h.svc.GetHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponseDTO{Error: errorDetail{Code: "INTERNAL_ERROR", Message: "履歴取得に失敗しました"}})
		return
	}

	items := make([]orderRecordDTO, len(recs))
	for i, r := range recs {
		items[i] = orderRecordDTO{
			OrderID: r.OrderID, Category: r.Category, StoreName: r.StoreName,
			MenuName: r.MenuName, Amount: r.Amount, OrderedAt: r.OrderedAt,
		}
	}
	c.JSON(http.StatusOK, getHistoryResponseDTO{Items: items})
}

// mapOrderError は OrderService エラーを HTTP ステータスにマッピングする。
func mapOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, order.ErrInsufficientFunds):
		c.JSON(http.StatusPaymentRequired, errorResponseDTO{
			Error: errorDetail{Code: "INSUFFICIENT_FUNDS", Message: "今月のダメ予算が足りません"},
		})
	case errors.Is(err, order.ErrIdempotencyConflict):
		c.JSON(http.StatusConflict, errorResponseDTO{
			Error: errorDetail{Code: "IDEMPOTENCY_CONFLICT", Message: "重複リクエストです"},
		})
	case errors.Is(err, order.ErrWalletUnconfigured):
		// Unit B WalletService 未配線 (B-C2 暫定): 503 SERVICE_UNAVAILABLE。
		// 402 INSUFFICIENT_FUNDS と区別することで NFRC-C22 の 402 メトリクス
		// 汚染を避け、Frontend も `/budget-empty` 遷移を起動しない。
		c.JSON(http.StatusServiceUnavailable, errorResponseDTO{
			Error: errorDetail{Code: "SERVICE_UNAVAILABLE", Message: "ダメ化サービスは準備中です"},
		})
	default:
		c.JSON(http.StatusInternalServerError, errorResponseDTO{
			Error: errorDetail{Code: "INTERNAL_ERROR", Message: "ダメ化に失敗しました"},
		})
	}
}

// parsePositiveInt は "20" 形式を int に変換する。負・0 はエラー。
func parsePositiveInt(s string) (int, error) {
	n := 0
	if len(s) == 0 {
		return 0, errors.New("empty")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("not a number")
		}
		n = n*10 + int(c-'0')
		if n > 1000 {
			return 0, errors.New("too large")
		}
	}
	if n <= 0 {
		return 0, errors.New("must be positive")
	}
	return n, nil
}
