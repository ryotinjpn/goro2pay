package logging

import (
	"context"
	"io"
	"log/slog"
)

// ContextAwareSlogHandler は context から指定キー (requestId, traceId, userId,
// userAgent, email_hash, action) を抽出して record に attrs を追加する slog.Handler。
// NFR Design P-OBS-01 / LC-AUTH-05 に従う。
// email_hash のみ snake_case (A-NFR-OBS-01 表に従う)、他は camelCase。
type ContextAwareSlogHandler struct {
	inner slog.Handler
}

// NewContextAwareSlogHandler は w を出力先とした JSON Handler を生成する。
// テスト時は w を bytes.Buffer に差し替えて出力内容を検証可能。
func NewContextAwareSlogHandler(w io.Writer, level slog.Level) *ContextAwareSlogHandler {
	inner := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	return &ContextAwareSlogHandler{inner: inner}
}

// Enabled は内部 Handler の Level に従う。
func (h *ContextAwareSlogHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.inner.Enabled(ctx, lvl)
}

// Handle は context から attrs を抽出して record にマージし、内部 Handler に委譲する。
func (h *ContextAwareSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, key := range []ctxKey{CtxKeyRequestID, CtxKeyTraceID, CtxKeyUserID, CtxKeyUserAgent, CtxKeyEmailHash, CtxKeyAction} {
		if v, ok := ctx.Value(key).(string); ok && v != "" {
			record.AddAttrs(slog.String(string(key), v))
		}
	}
	return h.inner.Handle(ctx, record)
}

// WithAttrs は内部 Handler に委譲する。
func (h *ContextAwareSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextAwareSlogHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup は内部 Handler に委譲する。
func (h *ContextAwareSlogHandler) WithGroup(name string) slog.Handler {
	return &ContextAwareSlogHandler{inner: h.inner.WithGroup(name)}
}
