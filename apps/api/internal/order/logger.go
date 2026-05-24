package order

import (
	"context"
	"log/slog"
)

// LogSummary は PlaceOrder 1 リクエストの 11 項目を蓄積し、完了時に
// `place_order_complete` イベント 1 件として出力する (P-OBS-02 / NFRC-C12)。
//
// PII 構造的防御 (NFRC-C24): Bedrock プロンプト本文・レスポンス本文の
// フィールドを意図的に持たないことで、ログに本文が混入する経路を排除する。
type LogSummary struct {
	ctx               context.Context
	idempotencyKey    string
	idempotent        bool
	orderID           string
	bedrockLatencyMs  int64
	bedrockAttempt    int
	fallbackTriggered bool
	category          string
	amount            int
	storeName         string
	menuName          string
	historyCount      int
	source            string
}

// NewLogSummary は ctx を保持した新しい LogSummary を返す。
//
// 呼出側で `defer s.LogComplete()` パターンを使うことで、成功/エラー両方で
// サマリログ出力が保証される。
func NewLogSummary(ctx context.Context) *LogSummary {
	return &LogSummary{ctx: ctx}
}

func (s *LogSummary) SetIdempotencyKey(v string)    { s.idempotencyKey = v }
func (s *LogSummary) SetIdempotent(v bool)          { s.idempotent = v }
func (s *LogSummary) SetOrderID(v string)           { s.orderID = v }
func (s *LogSummary) SetBedrockLatencyMs(v int64)   { s.bedrockLatencyMs = v }
func (s *LogSummary) SetBedrockAttempt(v int)       { s.bedrockAttempt = v }
func (s *LogSummary) SetFallbackTriggered(v bool)   { s.fallbackTriggered = v }
func (s *LogSummary) SetCategory(v string)          { s.category = v }
func (s *LogSummary) SetAmount(v int)               { s.amount = v }
func (s *LogSummary) SetStoreName(v string)         { s.storeName = v }
func (s *LogSummary) SetMenuName(v string)          { s.menuName = v }
func (s *LogSummary) SetHistoryCount(v int)         { s.historyCount = v }
func (s *LogSummary) SetSource(v string)            { s.source = v }

// LogComplete は蓄積した 11 項目を `place_order_complete` イベントとして出力する。
//
// `defer s.LogComplete()` で呼ぶことを前提とし、エラー時も必ず出力される。
func (s *LogSummary) LogComplete() {
	slog.InfoContext(s.ctx, "place_order_complete",
		"event", "place_order_complete",
		"idempotencyKey", s.idempotencyKey,
		"idempotent", s.idempotent,
		"orderId", s.orderID,
		"bedrockLatencyMs", s.bedrockLatencyMs,
		"bedrockAttempt", s.bedrockAttempt,
		"fallbackTriggered", s.fallbackTriggered,
		"category", s.category,
		"amount", s.amount,
		"storeName", s.storeName,
		"menuName", s.menuName,
		"historyCount", s.historyCount,
		"source", s.source,
	)
}
