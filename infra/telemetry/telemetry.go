package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Provider wraps OpenTelemetry wiring so callers can shut it down gracefully.
type Provider struct {
	serviceName string
	shutdown    func(context.Context) error
}

// ServiceName returns the configured service name for telemetry exporters.
func (p Provider) ServiceName() string {
	return p.serviceName
}

// Shutdown flushes telemetry data and releases related resources.
func (p Provider) Shutdown(ctx context.Context) error {
	if p.shutdown == nil {
		return nil
	}
	return p.shutdown(ctx)
}

// Setup configures OpenTelemetry tracing using environment-driven configuration.
// It prefers OTLP/HTTP exporter and falls back to stdout tracing when OTLP is not available.
func Setup(ctx context.Context, logger *slog.Logger) (Provider, error) {
	log := logger
	if log == nil {
		log = slog.Default()
	}

	serviceName := resolveServiceName()
	if disabled, err := shouldDisableSDK(); err != nil {
		log.Warn("invalid OTEL_SDK_DISABLED value", "value", os.Getenv("OTEL_SDK_DISABLED"), "error", err)
	} else if disabled {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		log.Info("telemetry disabled via OTEL_SDK_DISABLED")
		return Provider{serviceName: serviceName}, nil
	}

	res, err := newResource(ctx, serviceName)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to build telemetry resource: %w", err)
	}

	exp, err := newSpanExporter(ctx, log)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if exp != nil {
		opts = append(opts, sdktrace.WithBatcher(exp))
	} else {
		opts = append(opts, sdktrace.WithSampler(sdktrace.NeverSample()))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	log.Info("telemetry initialised", "service", serviceName)

	return Provider{
		serviceName: serviceName,
		shutdown: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	}, nil
}

func newSpanExporter(ctx context.Context, logger *slog.Logger) (sdktrace.SpanExporter, error) {
	exporter := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_TRACES_EXPORTER")))

	switch exporter {
	case "none":
		logger.Info("OTEL tracing disabled", "OTEL_TRACES_EXPORTER", exporter)
		return nil, nil
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "", "otlp":
		exp, err := otlptracehttp.New(ctx)
		if err != nil {
			logger.Warn("failed to create OTLP trace exporter, falling back to stdout", "error", err)
			return stdouttrace.New(stdouttrace.WithPrettyPrint())
		}
		return exp, nil
	default:
		logger.Warn("unsupported OTEL_TRACES_EXPORTER, using stdout exporter", "value", exporter)
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
}

func newResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{semconv.ServiceName(serviceName)}

	if version := strings.TrimSpace(os.Getenv("SERVICE_VERSION")); version != "" {
		attrs = append(attrs, semconv.ServiceVersion(version))
	}

	if env := resolveDeploymentEnvironment(); env != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(env))
	}

	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithAttributes(attrs...),
	)
}

func resolveServiceName() string {
	if fromEnv := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")); fromEnv != "" {
		return fromEnv
	}

	if fromEnv := strings.TrimSpace(os.Getenv("SERVICE_NAME")); fromEnv != "" {
		return fromEnv
	}

	return "cottage-manager"
}

func resolveDeploymentEnvironment() string {
	envVars := []string{"DEPLOYMENT_ENVIRONMENT", "APP_ENV", "ENVIRONMENT"}
	for _, key := range envVars {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func shouldDisableSDK() (bool, error) {
	raw := strings.TrimSpace(os.Getenv("OTEL_SDK_DISABLED"))
	if raw == "" {
		return false, nil
	}

	lowered := strings.ToLower(raw)
	switch lowered {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, errors.New("invalid value for OTEL_SDK_DISABLED")
	}
}
