// Package observability は Unit C / 横串の観測性ヘルパを提供する。
//
// Measure は P-OBS-01 の各ステップ計測担当で、関数を context + 名前付きで
// 包むだけでレイテンシログ出力を統一できる generic ヘルパ。
package observability

import (
	"context"
	"log/slog"
	"time"
)

// Measure は fn の実行時間を計測し `<name>_complete` ログを出力する。
//
// 戻り値:
//   - T   : fn の戻り値そのまま
//   - time.Duration : 計測値 (呼出元で LogSummary に蓄積する用途)
//   - error: fn のエラー
//
// fn 内で panic が発生した場合は panic を recover せず再 panic させる
// (上位の gin.Recovery / Lambda runtime に委ねる)。
func Measure[T any](ctx context.Context, name string, fn func() (T, error)) (T, time.Duration, error) {
	started := time.Now()
	result, err := fn()
	elapsed := time.Since(started)
	slog.InfoContext(ctx, name+"_complete",
		"event", name+"_complete",
		"latencyMs", elapsed.Milliseconds(),
		"success", err == nil,
	)
	return result, elapsed, err
}
