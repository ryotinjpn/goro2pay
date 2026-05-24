package order

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogBedrockRetry(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	defer slog.SetDefault(prev)

	LogBedrockRetry(context.Background(), 2, "ThrottlingException", 1500)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, "bedrock_retry", entry["event"])
	assert.EqualValues(t, 2, entry["attempt"])
	assert.Equal(t, "ThrottlingException", entry["errorClass"])
	assert.EqualValues(t, 1500, entry["elapsedMs"])
}

func TestLogFallbackTriggered(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	defer slog.SetDefault(prev)

	LogFallbackTriggered(context.Background(), "bedrock_double_failure", 5, "build_from_history")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, "fallback_triggered", entry["event"])
	assert.Equal(t, "bedrock_double_failure", entry["reason"])
	assert.EqualValues(t, 5, entry["historyCount"])
	assert.Equal(t, "build_from_history", entry["fallbackType"])
}
