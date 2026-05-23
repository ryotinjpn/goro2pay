package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextAwareSlogHandler_EmitsAttrsFromContext(t *testing.T) {
	var buf bytes.Buffer
	handler := NewContextAwareSlogHandler(&buf, slog.LevelInfo)
	logger := slog.New(handler)

	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxKeyRequestID, "req-123")
	ctx = context.WithValue(ctx, CtxKeyTraceID, "trace-abc")
	ctx = context.WithValue(ctx, CtxKeyUserID, "user-sub-1")
	ctx = context.WithValue(ctx, CtxKeyUserAgent, "Mozilla/5.0")
	ctx = context.WithValue(ctx, CtxKeyAction, "logout")

	logger.InfoContext(ctx, "user logout")

	// JSON ログとして出力されているはず
	var entry map[string]any
	err := json.Unmarshal(buf.Bytes(), &entry)
	assert.NoError(t, err)

	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "user logout", entry["msg"])
	assert.Equal(t, "req-123", entry["requestId"])
	assert.Equal(t, "trace-abc", entry["traceId"])
	assert.Equal(t, "user-sub-1", entry["userId"])
	assert.Equal(t, "Mozilla/5.0", entry["userAgent"])
	assert.Equal(t, "logout", entry["action"])
}

func TestContextAwareSlogHandler_OmitsMissingKeys(t *testing.T) {
	var buf bytes.Buffer
	handler := NewContextAwareSlogHandler(&buf, slog.LevelInfo)
	logger := slog.New(handler)

	// 認証必須 endpoint だが email_hash は context にない (P-SEC-02)
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxKeyRequestID, "req-1")
	ctx = context.WithValue(ctx, CtxKeyUserID, "sub-1")

	logger.InfoContext(ctx, "test")

	var entry map[string]any
	err := json.Unmarshal(buf.Bytes(), &entry)
	assert.NoError(t, err)

	assert.Equal(t, "req-1", entry["requestId"])
	assert.Equal(t, "sub-1", entry["userId"])
	_, hasEmailHash := entry["email_hash"]
	assert.False(t, hasEmailHash, "email_hash should be absent when not set in context")
}

func TestContextAwareSlogHandler_RespectsLogLevel(t *testing.T) {
	var buf bytes.Buffer
	handler := NewContextAwareSlogHandler(&buf, slog.LevelInfo)
	logger := slog.New(handler)

	ctx := context.Background()
	logger.DebugContext(ctx, "should not appear")
	logger.InfoContext(ctx, "should appear")

	output := buf.String()
	assert.NotContains(t, output, "should not appear")
	assert.Contains(t, output, "should appear")
}
