package telemetry

import (
	"context"
	"cottageManager/internal/config"
	"errors"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Shutdown func(context.Context) error

func Setup(ctx context.Context, telemetryConfig *config.TelemetryConfig) (Shutdown, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := sanitizeTelemetryConfig(telemetryConfig)
	if cfg == nil {
		return nil, nil
	}

	shutdown, err := initOtel(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return shutdown, nil
}

func initOtel(ctx context.Context, otelConfig *config.TelemetryConfig) (Shutdown, error) {
	endpoint := strings.TrimSpace(otelConfig.ExporterEndpoint)
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(otelConfig.ServiceName),
			semconv.DeploymentEnvironmentKey.String(otelConfig.DeploymentEnv),
			semconv.ServiceVersionKey.String(otelConfig.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	traceExp, err := otlptracegrpc.New(ctx, traceClientOptions(endpoint, otelConfig)...)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp),
	)
	otel.SetTracerProvider(tp)

	metricExp, err := otlpmetricgrpc.New(ctx, metricClientOptions(endpoint, otelConfig)...)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(10*time.Second),
		)),
	)
	otel.SetMeterProvider(mp)

	return func(ctx context.Context) error {
		var errs []error
		if err := mp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := tp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}, nil
}

func traceClientOptions(endpoint string, cfg *config.TelemetryConfig) []otlptracegrpc.Option {
	opts := make([]otlptracegrpc.Option, 0, 3)

	if hasScheme(endpoint) {
		opts = append(opts, otlptracegrpc.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}

	if shouldUseInsecure(endpoint, cfg) {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	return opts
}

func metricClientOptions(endpoint string, cfg *config.TelemetryConfig) []otlpmetricgrpc.Option {
	opts := make([]otlpmetricgrpc.Option, 0, 3)

	if hasScheme(endpoint) {
		opts = append(opts, otlpmetricgrpc.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(endpoint))
	}

	if shouldUseInsecure(endpoint, cfg) {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	return opts
}

func sanitizeTelemetryConfig(in *config.TelemetryConfig) *config.TelemetryConfig {
	if in == nil {
		return nil
	}

	cfg := *in
	cfg.ExporterEndpoint = strings.TrimSpace(cfg.ExporterEndpoint)
	cfg.ServiceName = strings.TrimSpace(cfg.ServiceName)
	cfg.DeploymentEnv = strings.TrimSpace(cfg.DeploymentEnv)
	cfg.ServiceVersion = strings.TrimSpace(cfg.ServiceVersion)

	if cfg.ExporterEndpoint == "" {
		cfg.ExporterEndpoint = "localhost:4317"
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "cottage-manager"
	}
	if cfg.DeploymentEnv == "" {
		cfg.DeploymentEnv = "development"
	}
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "0.0.1"
	}

	return &cfg
}

func shouldUseInsecure(endpoint string, cfg *config.TelemetryConfig) bool {
	if cfg != nil {
		return cfg.UseInsecure
	}

	return !strings.HasPrefix(strings.ToLower(endpoint), "https://")
}

func hasScheme(endpoint string) bool {
	return strings.Contains(endpoint, "://")
}
