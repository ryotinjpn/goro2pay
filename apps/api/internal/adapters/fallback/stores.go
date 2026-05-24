// Package fallback は Bedrock 呼出失敗時のフォールバック計画生成を提供する。
//
// MVP の Default 5 店舗は BR-C07 で確定したラインナップ。本実装は
// internal 定数として持ち、テストから直接参照可能にしている。
package fallback

// Store は fallback の店舗定義。
//
// Plan 互換のフィールドを持ち、FallbackSuggestProvider が Plan に変換する。
type Store struct {
	StoreName string
	MenuName  string
	Amount    int
	Category  string
}

// DefaultStores は Bedrock 失敗時に Default() でランダム選択する 5 店舗 (BR-C07)。
var DefaultStores = []Store{
	{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000, Category: "food"},
	{StoreName: "ぐうたら亭", MenuName: "手抜き丼", Amount: 800, Category: "food"},
	{StoreName: "ダメ屋", MenuName: "やる気なしカレー", Amount: 1200, Category: "food"},
	{StoreName: "怠惰キッチン", MenuName: "何でもよし弁当", Amount: 1500, Category: "food"},
	{StoreName: "ふぬけ食堂", MenuName: "しょうがない定食", Amount: 900, Category: "food"},
}
