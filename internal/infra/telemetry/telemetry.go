package telemetry

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Shutdown func(context.Context) error

func Setup(ctx context.Context, cfg config.TelemetryConfig) (Shutdown, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	shutdown, err := initOtel(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return shutdown, nil
}

func initOtel(ctx context.Context, otelConfig config.TelemetryConfig) (Shutdown, error) {
	endpoint := strings.TrimSpace(otelConfig.OTLPEndpoint)
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	// TODO check this
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(""),
			semconv.DeploymentEnvironmentKey.String(""),
			semconv.ServiceVersionKey.String(""),
		),
	)
	if err != nil {
		return nil, err
	}

	traceExp, err := otlptracegrpc.New(ctx, traceClientOptions(endpoint)...)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp),
	)
	otel.SetTracerProvider(tp)

	metricExp, err := otlpmetricgrpc.New(ctx, metricClientOptions(endpoint)...)
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

func traceClientOptions(endpoint string) []otlptracegrpc.Option {
	opts := make([]otlptracegrpc.Option, 0, 3)

	if hasScheme(endpoint) {
		opts = append(opts, otlptracegrpc.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}

	if shouldUseInsecure(endpoint) {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	return opts
}

func metricClientOptions(endpoint string) []otlpmetricgrpc.Option {
	opts := make([]otlpmetricgrpc.Option, 0, 3)

	if hasScheme(endpoint) {
		opts = append(opts, otlpmetricgrpc.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(endpoint))
	}

	if shouldUseInsecure(endpoint) {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	return opts
}

func shouldUseInsecure(endpoint string) bool {
	return !strings.HasPrefix(strings.ToLower(endpoint), "https://")
}

func hasScheme(endpoint string) bool {
	return strings.Contains(endpoint, "://")
}
