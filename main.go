package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Kenji-Uema/cottageManager/docs"
	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/config"
	"github.com/Kenji-Uema/cottageManager/internal/infra/db"
	"github.com/Kenji-Uema/cottageManager/internal/infra/logging"
	"github.com/Kenji-Uema/cottageManager/internal/infra/telemetry"
	"github.com/Kenji-Uema/cottageManager/internal/transport/http"
)

// @title Cottage Manager API
// @version 1.0
// @description API for managing cottages, availability, and bookings.
// @BasePath /
func main() {
	slog.SetDefault(logging.NewLogger())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	configs, err := config.LoadConfigs()
	exitOnError(ctx, "failed to load configs", err)

	shutdownTelemetry, err := telemetry.Init(ctx, configs.TelemetryConfig, configs.AppConfig)
	exitOnError(ctx, "failed to setup telemetry", err)

	mongoDb, err := db.NewMongoDbFromConfig(ctx, configs.MongoConfig)
	exitOnError(ctx, "failed to connect to MongoDB", err)

	cottageRepo := db.NewCottageRepo(mongoDb.Database, configs.CottageCollectionConfig)
	bookingRepo := db.NewBookingRepo(mongoDb.Database, configs.BookingCollectionConfig)
	txManager := db.NewMongoTxManager(mongoDb.Client)

	cottageService := app.NewCottageService(cottageRepo)
	bookingService := app.NewBookingService(cottageService, bookingRepo, txManager)
	availabilityService := app.NewAvailabilityService(cottageService, bookingService)

	httpServer := http.NewHttpServer(configs.ServerConfig)
	httpServer.SetupRoutes(availabilityService, bookingService, cottageService, mongoDb)
	httpServer.SetServer()
	go httpServer.Run(ctx)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.InfoContext(shutdownCtx, "shutdown signal received; shutting down")

	if err := shutdownTelemetry(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "telemetry shutdown", "error", err)
	}

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "server shutdown failed", "err", err)
	}

	if err := mongoDb.Close(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "mongo shutdown failed", "err", err)
	}
}

func exitOnError(ctx context.Context, errMsg string, err error) {
	if err != nil {
		slog.ErrorContext(ctx, errMsg, "error", err)
		os.Exit(1)
	}
}
