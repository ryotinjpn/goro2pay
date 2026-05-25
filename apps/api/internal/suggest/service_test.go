package suggest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	orderhistory "github.com/ryotinjpn/goro2pay/apps/api/internal/repo/order_history"
	"github.com/ryotinjpn/goro2pay/apps/api/internal/repo/suggestion"
)

var fixedNow = time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)

// mockReader は OrderHistoryReader の test mock。
type mockReader struct {
	records []orderhistory.OrderRecord
	err     error
}

func (m *mockReader) ListRecent(_ context.Context, _ string, _ int) ([]orderhistory.OrderRecord, error) {
	return m.records, m.err
}
func (m *mockReader) CountThisMonth(_ context.Context, _ string) (int, error) { return 0, nil }
func (m *mockReader) SumThisMonth(_ context.Context, _ string) (int, error)   { return 0, nil }

// mockBuilder は SuggestionBuilder の test mock。
type mockBuilder struct {
	result *BuiltSuggestion
	err    error
}

func (m *mockBuilder) Build(_ context.Context, _ []orderhistory.OrderRecord, _ string) (*BuiltSuggestion, error) {
	return m.result, m.err
}

// recordsDaysAgo は n 件の food 注文履歴を生成する (各 i 日前)。
func recordsDaysAgo(days ...int) []orderhistory.OrderRecord {
	out := make([]orderhistory.OrderRecord, len(days))
	for i, d := range days {
		out[i] = orderhistory.OrderRecord{
			OrderID:   fmt.Sprintf("ORD-%d", i),
			Category:  "food",
			StoreName: "CoCo壱",
			MenuName:  "ポークカレー",
			Amount:    1200,
			OrderedAt: fixedNow.Add(-time.Duration(d) * 24 * time.Hour).Format(time.RFC3339),
		}
	}
	return out
}

func newServiceForTest(reader *mockReader, builder *mockBuilder, store suggestion.SuggestionStore) *Service {
	s := NewService(reader, builder, store)
	s.SetClock(func() time.Time { return fixedNow })
	return s
}

func okBuilt() *BuiltSuggestion {
	return &BuiltSuggestion{
		Plan:           SuggestionPlan{StoreName: "CoCo壱", MenuName: "ポークカレー", Amount: 1200, Category: "food"},
		BedrockAttempt: 1,
	}
}

func TestGetSuggestion_InsufficientHistory(t *testing.T) {
	// 3 件 (< 5) → 非表示 (BR-D01/D02)
	reader := &mockReader{records: recordsDaysAgo(2, 4, 6)}
	s := newServiceForTest(reader, &mockBuilder{result: okBuilt()}, suggestion.NewInmemoryStore())

	got, err := s.GetSuggestion(context.Background(), "u1")
	require.NoError(t, err)
	require.False(t, got.HasSuggestion)
}

func TestGetSuggestion_Sufficient_ReturnsAndSaves(t *testing.T) {
	// 5 件 (>= 5)、直近 3h なし → 提案 + 保存 (BR-D01/D08)
	reader := &mockReader{records: recordsDaysAgo(1, 2, 3, 4, 5)}
	store := suggestion.NewInmemoryStore()
	store.SetClock(func() time.Time { return fixedNow })
	s := newServiceForTest(reader, &mockBuilder{result: okBuilt()}, store)

	got, err := s.GetSuggestion(context.Background(), "u1")
	require.NoError(t, err)
	require.True(t, got.HasSuggestion)
	require.NotEmpty(t, got.SuggestionID)
	require.NotNil(t, got.Plan)
	require.Equal(t, "CoCo壱", got.Plan.StoreName)

	// 保存され ResolveSuggestion で復元できる
	plan, err := s.ResolveSuggestion(context.Background(), got.SuggestionID)
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.Equal(t, 1200, plan.Amount)
}

func TestGetSuggestion_SuppressedByRecentOrder(t *testing.T) {
	// 5 件中 1 件が 1 時間前の food 注文 → 抑制 (BR-D03)
	recs := recordsDaysAgo(2, 3, 4, 5, 6)
	recs = append(recs, orderhistory.OrderRecord{
		OrderID: "ORD-recent", Category: "food", StoreName: "CoCo壱", MenuName: "カレー", Amount: 1200,
		OrderedAt: fixedNow.Add(-1 * time.Hour).Format(time.RFC3339),
	})
	reader := &mockReader{records: recs}
	s := newServiceForTest(reader, &mockBuilder{result: okBuilt()}, suggestion.NewInmemoryStore())

	got, err := s.GetSuggestion(context.Background(), "u1")
	require.NoError(t, err)
	require.False(t, got.HasSuggestion)
}

func TestGetSuggestion_BuilderEmptyPlan_NoSuggestion(t *testing.T) {
	// builder が空 Plan (BR-D06) → 非表示
	reader := &mockReader{records: recordsDaysAgo(1, 2, 3, 4, 5)}
	s := newServiceForTest(reader, &mockBuilder{result: &BuiltSuggestion{FallbackUsed: true}}, suggestion.NewInmemoryStore())

	got, err := s.GetSuggestion(context.Background(), "u1")
	require.NoError(t, err)
	require.False(t, got.HasSuggestion)
}

func TestResolveSuggestion_MissingReturnsNil(t *testing.T) {
	s := newServiceForTest(&mockReader{}, &mockBuilder{}, suggestion.NewInmemoryStore())
	plan, err := s.ResolveSuggestion(context.Background(), "01HXNONEXISTENT")
	require.NoError(t, err)
	require.Nil(t, plan) // 失効・不在 → nil (BR-D10、Unit C 透過フォールバック)
}
