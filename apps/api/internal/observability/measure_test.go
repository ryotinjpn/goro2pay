package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeasure_Success(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	got, _, err := Measure(context.Background(), "bedrock", func() (string, error) {
		return "ok", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", got)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "bedrock_complete", entry["event"])
	assert.Equal(t, true, entry["success"])
}

func TestMeasure_Error(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	expectedErr := errors.New("downstream timeout")
	_, _, err := Measure(context.Background(), "wallet", func() (any, error) {
		return nil, expectedErr
	})
	assert.ErrorIs(t, err, expectedErr)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "wallet_complete", entry["event"])
	assert.Equal(t, false, entry["success"])
}
