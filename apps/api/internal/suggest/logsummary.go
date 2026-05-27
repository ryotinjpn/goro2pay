package suggest

import (
	"context"
	"log/slog"
)

// LogSummary は GetSuggestion 1 リクエストの項目を蓄積し、完了時に
// `get_suggestion_complete` イベント 1 件として出力する (P-SG-OBS-01 / NFRD-D11)。
//
// PII 構造的防御 (NFRD-D16): Bedrock プロンプト本文・レスポンス本文の
// フィールドを意図的に持たない。
type LogSummary struct {
	ctx                   context.Context
	suggestionID          string
	hasSuggestion         bool
	fallbackUsed          bool
	historyCount          int
	suppressedRecentOrder bool
	bedrockLatencyMs      int64
	bedrockAttempt        int
}

// NewLogSummary は ctx を保持した新しい LogSummary を返す。
// `defer s.LogComplete()` パターンで成功/エラー両方の出力を保証する。
func NewLogSummary(ctx context.Context) *LogSummary {
	return &LogSummary{ctx: ctx}
}

func (s *LogSummary) SetSuggestionID(v string)        { s.suggestionID = v }
func (s *LogSummary) SetHasSuggestion(v bool)         { s.hasSuggestion = v }
func (s *LogSummary) SetFallbackUsed(v bool)          { s.fallbackUsed = v }
func (s *LogSummary) SetHistoryCount(v int)           { s.historyCount = v }
func (s *LogSummary) SetSuppressedRecentOrder(v bool) { s.suppressedRecentOrder = v }
func (s *LogSummary) SetBedrockLatencyMs(v int64)     { s.bedrockLatencyMs = v }
func (s *LogSummary) SetBedrockAttempt(v int)         { s.bedrockAttempt = v }

// LogComplete は蓄積した項目を `get_suggestion_complete` として出力する。
func (s *LogSummary) LogComplete() {
	slog.InfoContext(s.ctx, "get_suggestion_complete",
		"event", "get_suggestion_complete",
		"suggestionId", s.suggestionID,
		"hasSuggestion", s.hasSuggestion,
		"fallbackUsed", s.fallbackUsed,
		"historyCount", s.historyCount,
		"suppressedRecentOrder", s.suppressedRecentOrder,
		"bedrockLatencyMs", s.bedrockLatencyMs,
		"bedrockAttempt", s.bedrockAttempt,
	)
}
