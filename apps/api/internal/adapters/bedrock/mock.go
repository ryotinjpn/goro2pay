package bedrock

import "context"

// MockBedrockAdapter は BedrockAdapter の関数フィールド注入型 mock (P-MOCK-01)。
//
// テストごとに InferOrderPlanFunc を再代入してシナリオを切替える。
// 連続呼出時は Calls カウンタを参照することで attempt 回数の検証も可能。
type MockBedrockAdapter struct {
	InferOrderPlanFunc  func(ctx context.Context, history []HistoryItem, dayOfWeek string, category string) (*Plan, error)
	InferSuggestionFunc func(ctx context.Context, history []HistoryItem, dayOfWeek string) (*Plan, error)
	Calls               int
	SuggestCalls        int
}

// InferOrderPlan は MockBedrockAdapter の interface 実装。
//
// InferOrderPlanFunc 未設定時はデフォルト成功応答を返す (テストで毎回 setup
// しなくて済むようにするため)。
func (m *MockBedrockAdapter) InferOrderPlan(ctx context.Context, history []HistoryItem, dayOfWeek string, category string) (*Plan, error) {
	m.Calls++
	if m.InferOrderPlanFunc != nil {
		return m.InferOrderPlanFunc(ctx, history, dayOfWeek, category)
	}
	return &Plan{
		StoreName:        "ゴロゴロ食堂",
		MenuName:         "おまかせ定食",
		Amount:           1000,
		Category:         "food",
		Source:           "bedrock",
		BedrockLatencyMs: 100,
		BedrockAttempt:   1,
	}, nil
}

// InferSuggestion は MockBedrockAdapter の interface 実装 (Unit D)。
//
// InferSuggestionFunc 未設定時はデフォルト成功応答を返す。
func (m *MockBedrockAdapter) InferSuggestion(ctx context.Context, history []HistoryItem, dayOfWeek string) (*Plan, error) {
	m.SuggestCalls++
	if m.InferSuggestionFunc != nil {
		return m.InferSuggestionFunc(ctx, history, dayOfWeek)
	}
	return &Plan{
		StoreName:        "ゴロゴロ食堂",
		MenuName:         "おまかせ定食",
		Amount:           1000,
		Category:         "food",
		Source:           "bedrock",
		BedrockLatencyMs: 100,
		BedrockAttempt:   1,
	}, nil
}
