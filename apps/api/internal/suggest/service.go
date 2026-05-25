package suggest

import (
	"context"
	"log/slog"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/ryotinjpn/goro2pay/apps/api/internal/observability"
	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/suggestion"
)

// SuggestService は Unit D の公開 interface (凍結契約 §5.1)。
type SuggestService interface {
	GetSuggestion(ctx context.Context, userID string) (*Suggestion, error)
	ResolveSuggestion(ctx context.Context, suggestionID string) (*SuggestionPlan, error)
}

const (
	// historyWindow は履歴十分判定・抑制判定の対象期間 (BR-D01)。
	historyWindow = 30 * 24 * time.Hour
	// historyFetchLimit は reader から取得する最大件数。
	historyFetchLimit = 30
	// sufficiencyThreshold は履歴十分の閾値 (BR-D01、Unit C BR-C06 と統一)。
	sufficiencyThreshold = 5
	// suppressionWindow は直近注文抑制の時間窓 (BR-D03、design spec §3.2)。
	suppressionWindow = 3 * time.Hour
	// suggestionTTL は GoroPay_Suggestion の TTL (BR-D08 / NFRD-D05)。
	suggestionTTL = 30 * time.Minute
	// targetCategory は MVP の対象カテゴリ (BR-D17)。
	targetCategory = "food"
	// defaultTitle は API が返す Title (フロント表示は固定文言優先、BR-D13)。
	defaultTitle = "そろそろご飯めんどくさいですよね？"
)

// Service は SuggestService implementation (LC-SUGGEST-01)。
type Service struct {
	reader  orderhistory.OrderHistoryReader
	builder SuggestionBuilder
	store   suggestion.SuggestionStore
	now     func() time.Time
}

// NewService は production 用 Service を返す。
func NewService(reader orderhistory.OrderHistoryReader, builder SuggestionBuilder, store suggestion.SuggestionStore) *Service {
	return &Service{reader: reader, builder: builder, store: store, now: time.Now}
}

// SetClock は now() を差し替える (テスト用)。
func (s *Service) SetClock(now func() time.Time) { s.now = now }

// GetSuggestion は起動時の先回りサジェストを生成する (UC-D-01)。
//
//  1. 履歴取得 → 直近 30 日 5 件未満なら非表示 (BR-D01/D02)
//  2. 直近 3h 同カテゴリ注文済みなら抑制 (BR-D03)
//  3. Bedrock 推論 + リトライ → フォールバック (SuggestionBuilder)
//  4. suggestionId 採番 + TTL 30 分保存
func (s *Service) GetSuggestion(ctx context.Context, userID string) (*Suggestion, error) {
	summary := NewLogSummary(ctx)
	defer summary.LogComplete()

	now := s.now().UTC()

	history, _, err := observability.Measure(ctx, "history_query", func() ([]orderhistory.OrderRecord, error) {
		return s.reader.ListRecent(ctx, userID, historyFetchLimit)
	})
	if err != nil {
		// 履歴取得失敗時はサジェストなしに丸める (NFRD-D07、メイン機能は阻害しない)
		slog.WarnContext(ctx, "suggest_history_query_failed", "event", "suggest_history_query_failed", "error", err.Error())
		summary.SetHasSuggestion(false)
		return &Suggestion{HasSuggestion: false}, nil
	}

	recent := withinWindow(history, now, historyWindow)
	summary.SetHistoryCount(len(recent))

	// BR-D01/D02: 履歴十分判定
	if len(recent) < sufficiencyThreshold {
		slog.InfoContext(ctx, "suggest_insufficient_history", "event", "suggest_insufficient_history", "historyCount", len(recent))
		summary.SetHasSuggestion(false)
		return &Suggestion{HasSuggestion: false}, nil
	}

	// BR-D03: 直近 3h 同カテゴリ注文の抑制
	if hasRecentSameCategoryOrder(recent, now, suppressionWindow, targetCategory) {
		slog.InfoContext(ctx, "suggest_suppressed_recent_order", "event", "suggest_suppressed_recent_order")
		summary.SetSuppressedRecentOrder(true)
		summary.SetHasSuggestion(false)
		return &Suggestion{HasSuggestion: false}, nil
	}

	dayOfWeek := now.In(jst()).Weekday().String()
	built, err := s.builder.Build(ctx, recent, dayOfWeek)
	if err != nil {
		// ctx cancel 等。サジェストなしに丸める (透過、NFRD-D07)
		slog.WarnContext(ctx, "suggest_build_failed", "event", "suggest_build_failed", "error", err.Error())
		summary.SetHasSuggestion(false)
		return &Suggestion{HasSuggestion: false}, nil
	}
	summary.SetFallbackUsed(built.FallbackUsed)
	summary.SetBedrockLatencyMs(built.BedrockLatencyMs)
	summary.SetBedrockAttempt(built.BedrockAttempt)

	// BR-D06: 最頻も生成不可なら非表示
	if built.Plan.StoreName == "" {
		summary.SetHasSuggestion(false)
		return &Suggestion{HasSuggestion: false}, nil
	}

	suggestionID := ulid.Make().String()
	rec := &suggestion.SuggestionRecord{
		SuggestionID: suggestionID,
		UserID:       userID,
		Plan: suggestion.Plan{
			StoreName: built.Plan.StoreName,
			MenuName:  built.Plan.MenuName,
			Amount:    built.Plan.Amount,
			Category:  built.Plan.Category,
		},
		CreatedAt: now.Unix(),
		ExpiresAt: now.Add(suggestionTTL).Unix(),
	}
	if err := s.store.Save(ctx, rec); err != nil {
		// 保存失敗は非致命: カードは表示するが 1 タップ時は Unit C が透過フォールバック (BR-C10)
		slog.WarnContext(ctx, "suggest_store_save_failed", "event", "suggest_store_save_failed", "error", err.Error())
	}

	summary.SetSuggestionID(suggestionID)
	summary.SetHasSuggestion(true)
	plan := built.Plan
	return &Suggestion{
		HasSuggestion: true,
		SuggestionID:  suggestionID,
		Title:         defaultTitle,
		Plan:          &plan,
		FallbackUsed:  built.FallbackUsed,
	}, nil
}

// ResolveSuggestion は保存済み提案を払い出す (UC-D-02、Unit C が呼ぶ)。
//
// 失効 / 不在時は nil を返し、Unit C が透過的に通常 Bedrock フローへフォールバック
// する (BR-D10/D11、Unit C BR-C09/C10)。Bedrock 再検証はしない。
func (s *Service) ResolveSuggestion(ctx context.Context, suggestionID string) (*SuggestionPlan, error) {
	rec, err := s.store.Get(ctx, suggestionID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		slog.WarnContext(ctx, "suggestion_expired_or_missing", "event", "suggestion_expired_or_missing", "suggestionId", suggestionID)
		return nil, nil
	}
	return &SuggestionPlan{
		StoreName: rec.Plan.StoreName,
		MenuName:  rec.Plan.MenuName,
		Amount:    rec.Plan.Amount,
		Category:  rec.Plan.Category,
	}, nil
}

// jst は Asia/Tokyo タイムゾーン固定値を返す。
func jst() *time.Location { return time.FixedZone("Asia/Tokyo", 9*60*60) }

// withinWindow は orderedAt が now から window 以内の履歴のみ返す。
func withinWindow(history []orderhistory.OrderRecord, now time.Time, window time.Duration) []orderhistory.OrderRecord {
	cutoff := now.Add(-window)
	out := make([]orderhistory.OrderRecord, 0, len(history))
	for _, r := range history {
		t, err := time.Parse(time.RFC3339, r.OrderedAt)
		if err != nil {
			continue
		}
		if t.After(cutoff) {
			out = append(out, r)
		}
	}
	return out
}

// hasRecentSameCategoryOrder は window 以内に同カテゴリ注文があるか判定する (BR-D03)。
func hasRecentSameCategoryOrder(history []orderhistory.OrderRecord, now time.Time, window time.Duration, category string) bool {
	cutoff := now.Add(-window)
	for _, r := range history {
		if r.Category != category {
			continue
		}
		t, err := time.Parse(time.RFC3339, r.OrderedAt)
		if err != nil {
			continue
		}
		if t.After(cutoff) {
			return true
		}
	}
	return false
}
