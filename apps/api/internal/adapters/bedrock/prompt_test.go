package bedrock

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPrompt_ContainsAllSections(t *testing.T) {
	history := []HistoryItem{
		{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Category: "food", Amount: 1000, DayOfWeek: "Mon"},
	}
	prompt, err := BuildPrompt(history, "Tue", "food")
	require.NoError(t, err)

	assert.Contains(t, prompt, "## 履歴")
	assert.Contains(t, prompt, "## 現在の曜日")
	assert.Contains(t, prompt, "Tue")
	assert.Contains(t, prompt, "## カテゴリ")
	assert.Contains(t, prompt, "## 出力フォーマット")
	assert.Contains(t, prompt, "ゴロゴロ食堂")
}

func TestBuildPrompt_DoesNotIncludePII(t *testing.T) {
	// HistoryItem に userId / email フィールドがないことを構造的に保証する
	history := []HistoryItem{
		{StoreName: "ぐうたら亭", MenuName: "手抜き丼", Category: "food", Amount: 800, DayOfWeek: "Wed"},
	}
	prompt, err := BuildPrompt(history, "Wed", "food")
	require.NoError(t, err)
	assert.NotContains(t, strings.ToLower(prompt), "userid")
	assert.NotContains(t, strings.ToLower(prompt), "email")
}

func TestParsePlanResponse_ValidJSON(t *testing.T) {
	text := `{"storeName":"ダメ屋","menuName":"やる気なしカレー","amount":1200,"category":"food"}`
	plan, err := ParsePlanResponse(text)
	require.NoError(t, err)
	assert.Equal(t, "ダメ屋", plan.StoreName)
	assert.Equal(t, "やる気なしカレー", plan.MenuName)
	assert.Equal(t, 1200, plan.Amount)
	assert.Equal(t, "food", plan.Category)
}

func TestParsePlanResponse_WithSurroundingText(t *testing.T) {
	text := `わかりました、提案します:
{"storeName":"怠惰キッチン","menuName":"何でもよし弁当","amount":1500,"category":"food"}
以上です。`
	plan, err := ParsePlanResponse(text)
	require.NoError(t, err)
	assert.Equal(t, "怠惰キッチン", plan.StoreName)
}

func TestParsePlanResponse_InvalidCases(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"no JSON object", "Sorry, I cannot help."},
		{"missing field", `{"storeName":"X","menuName":"Y","category":"food"}`},
		{"zero amount", `{"storeName":"X","menuName":"Y","amount":0,"category":"food"}`},
		{"empty store", `{"storeName":"","menuName":"Y","amount":100,"category":"food"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePlanResponse(tt.text)
			assert.Error(t, err)
		})
	}
}
