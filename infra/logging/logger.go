package logging

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

const timeLayoutWithMillis = "2006-01-02T15:04:05.000Z07:00"

// Setup configures the global structured logger based on environment variables.
// Supported levels: debug, info, warn, error. Format can be "json" or "text" (default json).
func Setup() *slog.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))

	handlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey && attr.Value.Kind() == slog.KindTime {
				attr.Value = slog.StringValue(formatWithMillis(attr.Value.Time()))
			}
			return attr
		},
	}

	var handler slog.Handler
	switch format {
	case "", "json":
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	default:
		// fallback to JSON when the format is unknown
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		fallthrough
	default:
		return slog.LevelInfo
	}
}

func formatWithMillis(t time.Time) string {
	if t.IsZero() {
		return t.Format(timeLayoutWithMillis)
	}

	return t.Truncate(time.Millisecond).Format(timeLayoutWithMillis)
}
