package logging

import (
	"context"
	"cottageManager/internal/config"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	sdlog "go.opentelemetry.io/otel/sdk/log"
)

func NewOtelHandler(ctx context.Context, cfg *config.TelemetryConfig) (slog.Handler, func(context.Context) error, error) {
	opts := make([]otlploggrpc.Option, 0, 3)

	opts = append(opts, otlploggrpc.WithEndpoint(cfg.ExporterEndpoint))
	if cfg.UseInsecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	exporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create OTLP log exporter: %w", err)
	}

	processor := sdlog.NewBatchProcessor(exporter)
	logProvider := sdlog.NewLoggerProvider(sdlog.WithProcessor(processor))
	global.SetLoggerProvider(logProvider)

	otelHandler := otelslog.NewHandler(cfg.ServiceName,
		otelslog.WithLoggerProvider(logProvider),
		otelslog.WithSource(true),
	)

	return otelHandler, logProvider.Shutdown, nil
}
