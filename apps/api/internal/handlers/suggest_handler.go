package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/auth"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/suggest"
)

// SuggestHandler は Unit D の GET /api/suggest ハンドラ (LC-SUGGEST-02)。
type SuggestHandler struct {
	svc suggest.SuggestService
}

// NewSuggestHandler は SuggestHandler を返す。
func NewSuggestHandler(svc suggest.SuggestService) *SuggestHandler {
	return &SuggestHandler{svc: svc}
}

// suggestPlanDTO は提案内容の JSON 表現 (凍結契約 §5.2)。
type suggestPlanDTO struct {
	StoreName string `json:"storeName"`
	MenuName  string `json:"menuName"`
	Amount    int    `json:"amount"`
	Category  string `json:"category"`
}

// suggestResponseDTO は GET /api/suggest のレスポンス (凍結契約 §5.2)。
//
// FallbackUsed は内部ログ専用で API には出さない (BR-D12)。
type suggestResponseDTO struct {
	HasSuggestion bool            `json:"hasSuggestion"`
	SuggestionID  string          `json:"suggestionId,omitempty"`
	Title         string          `json:"title,omitempty"`
	Plan          *suggestPlanDTO `json:"plan,omitempty"`
}

// GetSuggestion は GET /api/suggest ハンドラ。
//
// マウント時に 1 回呼ばれる (NFRD-D17)。履歴不足・抑制時は hasSuggestion:false。
func (h *SuggestHandler) GetSuggestion(c *gin.Context) {
	userID, err := auth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponseDTO{Error: errorDetail{Code: "UNAUTHORIZED", Message: "認証が必要です"}})
		return
	}

	res, err := h.svc.GetSuggestion(c.Request.Context(), userID)
	if err != nil {
		// サジェスト取得失敗はメイン機能を阻害しない: 200 + hasSuggestion:false に丸める
		// (NFRD-D07)。Service 側でも丸めているが二重防御。
		c.JSON(http.StatusOK, suggestResponseDTO{HasSuggestion: false})
		return
	}

	if !res.HasSuggestion {
		c.JSON(http.StatusOK, suggestResponseDTO{HasSuggestion: false})
		return
	}

	resp := suggestResponseDTO{
		HasSuggestion: true,
		SuggestionID:  res.SuggestionID,
		Title:         res.Title,
	}
	if res.Plan != nil {
		resp.Plan = &suggestPlanDTO{
			StoreName: res.Plan.StoreName,
			MenuName:  res.Plan.MenuName,
			Amount:    res.Plan.Amount,
			Category:  res.Plan.Category,
		}
	}
	c.JSON(http.StatusOK, resp)
}
