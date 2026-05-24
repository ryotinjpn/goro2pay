package order

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogSummary_LogComplete_Emits11Fields(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	s := NewLogSummary(context.Background())
	s.SetIdempotencyKey("01HZIDEM")
	s.SetIdempotent(false)
	s.SetOrderID("01HZORDER")
	s.SetBedrockLatencyMs(1234)
	s.SetBedrockAttempt(1)
	s.SetFallbackTriggered(false)
	s.SetCategory("food")
	s.SetAmount(1000)
	s.SetStoreName("ゴロゴロ食堂")
	s.SetMenuName("おまかせ定食")
	s.SetHistoryCount(7)
	s.SetSource("button")

	s.LogComplete()

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "place_order_complete", entry["event"])
	assert.Equal(t, "01HZIDEM", entry["idempotencyKey"])
	assert.Equal(t, "01HZORDER", entry["orderId"])
	assert.EqualValues(t, 1234, entry["bedrockLatencyMs"])
	assert.EqualValues(t, 1, entry["bedrockAttempt"])
	assert.Equal(t, false, entry["fallbackTriggered"])
	assert.Equal(t, "food", entry["category"])
	assert.EqualValues(t, 1000, entry["amount"])
	assert.Equal(t, "ゴロゴロ食堂", entry["storeName"])
	assert.Equal(t, "おまかせ定食", entry["menuName"])
	assert.EqualValues(t, 7, entry["historyCount"])
	assert.Equal(t, "button", entry["source"])
}

// TestLogSummary_StructHasNoBedrockBodyFields は NFRC-C24 構造的 PII 防御を検証する。
//
// LogSummary struct の field に Bedrock のプロンプト本文 / レスポンス本文を
// 表す名前 (prompt / response / messages 等) が含まれないことを reflect で確認する。
func TestLogSummary_StructHasNoBedrockBodyFields(t *testing.T) {
	v := reflect.TypeOf(LogSummary{})
	forbidden := []string{"prompt", "response", "messages", "responsebody", "promptbody"}
	for i := 0; i < v.NumField(); i++ {
		fname := strings.ToLower(v.Field(i).Name)
		for _, banned := range forbidden {
			assert.NotEqual(t, banned, fname, "LogSummary must not contain Bedrock body field: %s", fname)
		}
	}
}
