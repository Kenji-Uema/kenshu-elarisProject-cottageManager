package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/Kenji-Uema/cottageManager/internal/config"
)

func Setup(ctx context.Context, logConfig config.LogConfig, telemetryConfig config.TelemetryConfig) (func(context.Context) error, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	level := parseLevel(logConfig.Level)

	handler, shutdown, err := NewOtelHandler(ctx, telemetryConfig)
	if err != nil {
		slog.Error("failed to initialize OTel log exporter", "error", err)
		handler = nil
		shutdown = nil
	}

	if handler == nil {
		handler = newConsoleHandler(logConfig.Format, level)
	}

	contextHandler := NewContextHandler(handler, level)
	logger := slog.New(contextHandler)
	slog.SetDefault(logger)

	return shutdown, nil
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func newConsoleHandler(format string, level slog.Leveler) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return slog.NewJSONHandler(os.Stdout, opts)
	default:
		return slog.NewTextHandler(os.Stdout, opts)
	}
}
