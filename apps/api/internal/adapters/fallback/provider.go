package fallback

import (
	"math/rand"
)

// HistoryItem は fallback が参照する履歴エントリ。
//
// BedrockAdapter の HistoryItem と同じ形だがパッケージ独立性のため再宣言する
// (Adapter 同士の循環依存を避ける)。
type HistoryItem struct {
	StoreName string
	MenuName  string
	Amount    int
	Category  string
}

// Plan は fallback が返す注文計画。BedrockAdapter.Plan と互換のフィールドを持つ。
type Plan struct {
	StoreName string
	MenuName  string
	Amount    int
	Category  string
	Source    string // "fallback_history" | "fallback_default"
}

// FallbackSuggestProvider はフォールバック計画を生成する interface。
//
// Unit C / D で共有可能 (Q-D2=B / NFRC-C08)。
type FallbackSuggestProvider interface {
	BuildFromHistory(history []HistoryItem) *Plan
	Default() *Plan
}

// SimpleFallbackProvider は最頻パターン抽出 + ランダム Default の標準実装。
//
// rand は math/rand を採用 (暗号学的乱数は不要、ダメ化エンタメ性で十分)。
// テスト時は NewSimpleFallbackProviderWithRand で seed 固定の rand を注入できる。
type SimpleFallbackProvider struct {
	rng *rand.Rand
}

// NewSimpleFallbackProvider は production 用、デフォルトの rng を使う実装を返す。
func NewSimpleFallbackProvider() *SimpleFallbackProvider {
	return &SimpleFallbackProvider{rng: rand.New(rand.NewSource(rand.Int63()))}
}

// NewSimpleFallbackProviderWithRand はテスト用に rng を注入できる constructor。
func NewSimpleFallbackProviderWithRand(rng *rand.Rand) *SimpleFallbackProvider {
	return &SimpleFallbackProvider{rng: rng}
}

// BuildFromHistory は履歴から最頻パターンを抽出して Plan を返す。
//
// 同点首位が複数ある場合は最初に出現した item を採用する (決定論性のため)。
// 履歴が空の場合は nil を返す (呼出元は閾値分岐 BR-C06 で空時には Default を呼ぶ)。
func (p *SimpleFallbackProvider) BuildFromHistory(history []HistoryItem) *Plan {
	if len(history) == 0 {
		return nil
	}
	type key struct {
		Store string
		Menu  string
	}
	counts := make(map[key]int)
	first := make(map[key]HistoryItem)
	for _, item := range history {
		k := key{item.StoreName, item.MenuName}
		counts[k]++
		if _, ok := first[k]; !ok {
			first[k] = item
		}
	}
	var maxCount int
	var winner key
	for _, item := range history {
		k := key{item.StoreName, item.MenuName}
		if counts[k] > maxCount {
			maxCount = counts[k]
			winner = k
		}
	}
	chosen := first[winner]
	return &Plan{
		StoreName: chosen.StoreName,
		MenuName:  chosen.MenuName,
		Amount:    chosen.Amount,
		Category:  chosen.Category,
		Source:    "fallback_history",
	}
}

// Default は DefaultStores から rand.Intn でランダム選択する。
func (p *SimpleFallbackProvider) Default() *Plan {
	idx := p.rng.Intn(len(DefaultStores))
	s := DefaultStores[idx]
	return &Plan{
		StoreName: s.StoreName,
		MenuName:  s.MenuName,
		Amount:    s.Amount,
		Category:  s.Category,
		Source:    "fallback_default",
	}
}
