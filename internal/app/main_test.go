package app

import (
	"log/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	os.Exit(m.Run())
}
