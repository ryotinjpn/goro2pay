package order

import (
	"context"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/bedrock"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/adapters/fallback"
)

// Plan は注文計画 (Bedrock 成功 or fallback)。
//
// PlanBuilder の戻り値で、OrderService が Wallet.Deduct → Delivery.Place →
// History.Insert の連鎖を組み立てる際の入力になる。
type Plan struct {
	StoreName         string
	MenuName          string
	Amount            int
	Category          string
	Source            string // "bedrock" | "fallback_history" | "fallback_default"
	BedrockLatencyMs  int64
	BedrockAttempt    int
	FallbackTriggered bool
}

// PlanBuilder は Bedrock 呼出 + リトライ判定 + フォールバック分岐の
// 一連のロジックを集約する interface (P-PLAN-01)。
type PlanBuilder interface {
	Build(ctx context.Context, history []OrderRecord, dayOfWeek string, category string) (*Plan, error)
}

// fallbackThreshold は履歴件数による分岐閾値 (BR-C06、NFRC-C08)。
//
// 5 件以上 → BuildFromHistory、5 件未満 → Default。
const fallbackThreshold = 5

// BedrockPlanBuilder は Bedrock + Fallback を統合する PlanBuilder implementation。
type BedrockPlanBuilder struct {
	bedrock  bedrock.BedrockAdapter
	fallback fallback.FallbackSuggestProvider
}

// NewBedrockPlanBuilder は production 用 PlanBuilder を返す。
func NewBedrockPlanBuilder(b bedrock.BedrockAdapter, f fallback.FallbackSuggestProvider) *BedrockPlanBuilder {
	return &BedrockPlanBuilder{bedrock: b, fallback: f}
}

// toBedrockHistory は OrderRecord を Bedrock の HistoryItem に変換する。
func toBedrockHistory(history []OrderRecord) []bedrock.HistoryItem {
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

// toFallbackHistory は OrderRecord を fallback の HistoryItem に変換する。
func toFallbackHistory(history []OrderRecord) []fallback.HistoryItem {
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

// Build は Bedrock を呼んで成功時はその plan、失敗時はフォールバック plan を返す。
//
// 親 ctx の状態 (cancel / deadline) は Bedrock 呼出側で伝播する。
func (b *BedrockPlanBuilder) Build(ctx context.Context, history []OrderRecord, dayOfWeek string, category string) (*Plan, error) {
	bedrockHistory := toBedrockHistory(history)

	bp, err := b.bedrock.InferOrderPlan(ctx, bedrockHistory, dayOfWeek, category)
	if err == nil {
		return &Plan{
			StoreName:         bp.StoreName,
			MenuName:          bp.MenuName,
			Amount:            bp.Amount,
			Category:          bp.Category,
			Source:            bp.Source,
			BedrockLatencyMs:  bp.BedrockLatencyMs,
			BedrockAttempt:    bp.BedrockAttempt,
			FallbackTriggered: false,
		}, nil
	}

	// 親 ctx が cancel / deadline 超過なら即返す (NFRC-C10)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}

	// Bedrock 失敗 → フォールバック分岐 (NFRC-C08, BR-C06)
	historyCount := len(history)
	fallbackHistory := toFallbackHistory(history)

	var fbPlan *fallback.Plan
	var fallbackType string
	if historyCount >= fallbackThreshold {
		fbPlan = b.fallback.BuildFromHistory(fallbackHistory)
		fallbackType = "build_from_history"
	}
	if fbPlan == nil {
		fbPlan = b.fallback.Default()
		fallbackType = "default"
	}

	LogFallbackTriggered(ctx, "bedrock_failure", historyCount, fallbackType)

	return &Plan{
		StoreName:         fbPlan.StoreName,
		MenuName:          fbPlan.MenuName,
		Amount:            fbPlan.Amount,
		Category:          fbPlan.Category,
		Source:            fbPlan.Source,
		BedrockLatencyMs:  0,
		BedrockAttempt:    2, // 試行は 2 回消費したと記録 (NFRC-C06)
		FallbackTriggered: true,
	}, nil
}
