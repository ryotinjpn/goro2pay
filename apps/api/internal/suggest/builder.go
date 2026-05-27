package suggest

import (
	"context"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/bedrock"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/fallback"
	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
)

// BuiltSuggestion は SuggestionBuilder.Build の戻り値。
type BuiltSuggestion struct {
	Plan             SuggestionPlan
	FallbackUsed     bool
	BedrockLatencyMs int64
	BedrockAttempt   int
}

// SuggestionBuilder は Bedrock 推論 + リトライ + フォールバックを集約する
// interface (P-SG-BUILD-01)。order.PlanBuilder の suggest 版。
type SuggestionBuilder interface {
	Build(ctx context.Context, history []orderhistory.OrderRecord, dayOfWeek string) (*BuiltSuggestion, error)
}

// BedrockSuggestionBuilder は Bedrock + Fallback を統合する SuggestionBuilder 実装。
type BedrockSuggestionBuilder struct {
	bedrock  bedrock.BedrockAdapter
	fallback fallback.FallbackSuggestProvider
}

// NewBedrockSuggestionBuilder は production 用 SuggestionBuilder を返す。
func NewBedrockSuggestionBuilder(b bedrock.BedrockAdapter, f fallback.FallbackSuggestProvider) *BedrockSuggestionBuilder {
	return &BedrockSuggestionBuilder{bedrock: b, fallback: f}
}

func toBedrockHistory(history []orderhistory.OrderRecord) []bedrock.HistoryItem {
	out := make([]bedrock.HistoryItem, len(history))
	for i, r := range history {
		out[i] = bedrock.HistoryItem{
			StoreName: r.StoreName,
			MenuName:  r.MenuName,
			Category:  r.Category,
			Amount:    r.Amount,
			DayOfWeek: r.DayOfWeek,
		}
	}
	return out
}

func toFallbackHistory(history []orderhistory.OrderRecord) []fallback.HistoryItem {
	out := make([]fallback.HistoryItem, len(history))
	for i, r := range history {
		out[i] = fallback.HistoryItem{
			StoreName: r.StoreName,
			MenuName:  r.MenuName,
			Amount:    r.Amount,
			Category:  r.Category,
		}
	}
	return out
}

// Build は Bedrock で先回り提案を生成し、失敗時は履歴最頻フォールバックを返す。
//
// 呼出元 (SuggestService) は履歴十分 (>= 5 件) を保証済みのため、BuildFromHistory
// は最頻パターンを返せる (NFRD-D04)。それも nil の異常時は ok=false 相当として
// FallbackUsed=true・空 Plan を返さず、呼出側が hasSuggestion=false に丸める。
func (b *BedrockSuggestionBuilder) Build(ctx context.Context, history []orderhistory.OrderRecord, dayOfWeek string) (*BuiltSuggestion, error) {
	bp, err := b.bedrock.InferSuggestion(ctx, toBedrockHistory(history), dayOfWeek)
	if err == nil {
		return &BuiltSuggestion{
			Plan:             SuggestionPlan{StoreName: bp.StoreName, MenuName: bp.MenuName, Amount: bp.Amount, Category: bp.Category},
			FallbackUsed:     false,
			BedrockLatencyMs: bp.BedrockLatencyMs,
			BedrockAttempt:   bp.BedrockAttempt,
		}, nil
	}

	// 親 ctx の cancel / deadline は即返す (NFRC-C10 と同方針)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}

	// Bedrock 失敗 → 履歴最頻フォールバック (BR-D05 / NFRD-D04)
	fbPlan := b.fallback.BuildFromHistory(toFallbackHistory(history))
	if fbPlan == nil {
		// BR-D06: 最頻も生成不可 (理論上ほぼ起きない) → 提案なし扱い
		return &BuiltSuggestion{FallbackUsed: true, BedrockAttempt: bedrockMaxAttempts}, nil
	}
	return &BuiltSuggestion{
		Plan:           SuggestionPlan{StoreName: fbPlan.StoreName, MenuName: fbPlan.MenuName, Amount: fbPlan.Amount, Category: fbPlan.Category},
		FallbackUsed:   true,
		BedrockAttempt: bedrockMaxAttempts,
	}, nil
}

// bedrockMaxAttempts は試行回数の記録値 (NFRD-D03、bedrock 側 maxAttempts と整合)。
const bedrockMaxAttempts = 2
