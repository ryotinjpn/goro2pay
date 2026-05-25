// Package suggest は Unit D の学習・先回りサジェスト機能を提供する。
//
// GetSuggestion (起動時の先回り提案) と ResolveSuggestion (1 タップ注文時の
// 保存値払い出し、Unit C が呼ぶ) の 2 ユースケースを担う。
package suggest

// SuggestionPlan は提案 1 件の内容 (凍結契約 §5.1)。
type SuggestionPlan struct {
	StoreName string
	MenuName  string
	Amount    int
	Category  string
}

// Suggestion は GetSuggestion の戻り値 (凍結契約 §5.1)。
//
// HasSuggestion=false のとき SuggestionID / Plan は空 (US-2-02、カード非表示)。
// FallbackUsed は内部ログ用で API レスポンスには出さない (BR-D12)。
type Suggestion struct {
	HasSuggestion bool
	SuggestionID  string
	Title         string
	Plan          *SuggestionPlan
	FallbackUsed  bool
}
