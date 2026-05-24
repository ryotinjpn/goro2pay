package fallback

// FakeFallbackProvider は FallbackSuggestProvider の関数フィールド注入型 mock (P-MOCK-01)。
type FakeFallbackProvider struct {
	BuildFromHistoryFunc func(history []HistoryItem) *Plan
	DefaultFunc          func() *Plan
	BuildCalls           int
	DefaultCalls         int
}

func (f *FakeFallbackProvider) BuildFromHistory(history []HistoryItem) *Plan {
	f.BuildCalls++
	if f.BuildFromHistoryFunc != nil {
		return f.BuildFromHistoryFunc(history)
	}
	return &Plan{StoreName: "fake-history", MenuName: "fake", Amount: 500, Category: "food", Source: "fallback_history"}
}

func (f *FakeFallbackProvider) Default() *Plan {
	f.DefaultCalls++
	if f.DefaultFunc != nil {
		return f.DefaultFunc()
	}
	return &Plan{StoreName: "fake-default", MenuName: "fake", Amount: 1000, Category: "food", Source: "fallback_default"}
}
