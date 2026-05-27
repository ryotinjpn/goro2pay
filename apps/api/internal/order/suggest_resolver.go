package order

import "context"

// ResolvedSuggestion は SuggestResolver が返す解決済み提案 (order package 固有型)。
//
// order が suggest package を直接 import しないために独自型を定義し、
// 配線側 (main.go) で adapter 経由で suggest.SuggestService をブリッジする
// (WalletService / OrderAdapter と同じ方式、Q-DG1=B)。
type ResolvedSuggestion struct {
	StoreName string
	MenuName  string
	Amount    int
	Category  string
}

// SuggestResolver は 1 タップ注文時に suggestionId から保存済み提案を解決する
// interface (凍結契約 §4.5、Unit D 提供)。
//
// 失効・不在時は nil, nil を返す契約 (BR-C10、Unit C は透過的に Bedrock
// 推論へフォールバックする)。Bedrock 再検証はしない (BR-C09)。
type SuggestResolver interface {
	ResolveSuggestion(ctx context.Context, suggestionID string) (*ResolvedSuggestion, error)
}

// SetSuggestResolver は SuggestResolver を注入する (main.go の DI 配線)。
//
// 未設定 (nil) の場合、PlaceOrder は suggestionId を無視して常に Bedrock 推論を
// 行う (Unit D 未配線環境でも動作する後方互換)。
func (s *Service) SetSuggestResolver(r SuggestResolver) {
	s.suggestResolver = r
}
