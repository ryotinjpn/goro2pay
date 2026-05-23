package order

import (
	"context"
	"log/slog"
)

// EventLogger は P-OBS-03 の WARNING ログ層を担う。
//
// PlaceOrder のサマリログ (LogSummary, INFO) とは別経路で、リトライ発動 /
// フォールバック発動の "瞬間" を WARN として記録する。CloudWatch Logs metric
// filter (`level = "WARN" and event = "bedrock_retry"`) で抽出して
// NFRC-C13-2 / NFRC-C13-3 アラームに供給する。

// LogBedrockRetry は Bedrock リトライ発動時に WARN ログを 1 件出力する。
//
// errorClass は RetryClassifier 経由で判定したエラー型名 (例 "ThrottlingException")。
func LogBedrockRetry(ctx context.Context, attempt int, errorClass string, elapsedMs int64) {
	slog.WarnContext(ctx, "bedrock_retry",
		"event", "bedrock_retry",
		"attempt", attempt,
		"errorClass", errorClass,
		"elapsedMs", elapsedMs,
	)
}

// LogFallbackTriggered はフォールバック発動時に WARN ログを 1 件出力する。
//
// reason は "bedrock_double_failure" | "permanent_error"、
// fallbackType は "build_from_history" | "default"。
func LogFallbackTriggered(ctx context.Context, reason string, historyCount int, fallbackType string) {
	slog.WarnContext(ctx, "fallback_triggered",
		"event", "fallback_triggered",
		"reason", reason,
		"historyCount", historyCount,
		"fallbackType", fallbackType,
	)
}
