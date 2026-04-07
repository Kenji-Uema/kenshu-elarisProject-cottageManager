package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const appTracerName = "cottage-manager.app"

func StartAppSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(appTracerName).Start(ctx, name, opts...)
}

func RecordSpanError(span trace.Span, err error, status string) {
	if span == nil || err == nil {
		return
	}

	span.RecordError(err)
	if status == "" {
		status = err.Error()
	}
	span.SetStatus(codes.Error, status)
}
