package logging

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const slogFields string = "slog_fields"

type ContextHandler struct {
	slog.Handler
	level slog.Leveler
}

func NewContextHandler(handler slog.Handler, level slog.Leveler) *ContextHandler {
	return &ContextHandler{
		Handler: handler,
		level:   level,
	}
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.Handler == nil {
		return errors.New("no slog handler configured")
	}

	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok && len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}

	if traceAttrs := otelAttrs(ctx); len(traceAttrs) > 0 {
		r.AddAttrs(traceAttrs...)
	}

	return h.Handler.Handle(ctx, r)
}

func (h *ContextHandler) Enabled(_ context.Context, l slog.Level) bool {
	if h.Handler == nil {
		return false
	}

	if h.level == nil {
		return true
	}

	return l >= h.level.Level()
}

func AppendCtx(parent context.Context, attr slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(slogFields).([]slog.Attr); ok {
		out := make([]slog.Attr, 0, len(v)+1)
		out = append(out, v...)
		out = append(out, attr)

		return context.WithValue(parent, slogFields, out)
	}

	return context.WithValue(parent, slogFields, []slog.Attr{attr})
}

func otelAttrs(ctx context.Context) []slog.Attr {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return nil
	}

	attrs := make([]slog.Attr, 0, 2)
	attrs = append(attrs, slog.String("trace_id", spanCtx.TraceID().String()))
	attrs = append(attrs, slog.String("span_id", spanCtx.SpanID().String()))

	return attrs
}
