package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/config"
	"github.com/Kenji-Uema/cottageManager/internal/infra/db"
	"github.com/Kenji-Uema/cottageManager/internal/infra/logging"
	"github.com/Kenji-Uema/cottageManager/internal/infra/telemetry"
	httphandler "github.com/Kenji-Uema/cottageManager/internal/transport/http"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func exitOnError(errMsg string, err error) {
	if err != nil {
		slog.Error(errMsg, "error", err)
		os.Exit(1)
	}
}

func main() {
	ctx := context.Background()
	configs, err := config.LoadConfigs()
	exitOnError("failed to load configs", err)

	loggingShutdown, err := logging.Setup(ctx, configs.LogConfig, configs.TelemetryConfig)
	exitOnError("failed to setup logging", err)

	telemetryShutdown, err := telemetry.Setup(ctx, configs.TelemetryConfig)
	exitOnError("failed to setup telemetry", err)

	mongoDb := mongoDbSetup(configs.MongoConfig)

	router := ginSetup()
	routerSetup(mongoDb, router, configs.CottageCollectionConfig, configs.BookingCollectionConfig)
	server := ginSpinUP(configs.AppConfig, router)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server shutdown failed", "err", err)
	}

	if err := mongoDb.Close(shutdownCtx); err != nil {
		slog.Error("mongo shutdown failed", "err", err)
	}

	if telemetryShutdown != nil {
		if err := telemetryShutdown(shutdownCtx); err != nil {
			slog.Error("otel shutdown failed", "err", err)
		}
	}

	slog.Info("server exited gracefully")

	if loggingShutdown != nil {
		if err := loggingShutdown(shutdownCtx); err != nil {
			slog.Error("failed to flush logs: %v\n", err)
		}
	}
}

func routerSetup(mongoDb *db.Db, router *gin.Engine, cottageConf config.CottageCollectionConfig, bookingConf config.BookingCollectionConfig) {
	cottageRepo := db.NewCottageRepo(mongoDb.Database, cottageConf)
	bookingRepo := db.NewBookingRepo(mongoDb.Database, bookingConf)

	availabilityService := app.NewAvailabilityService(cottageRepo, bookingRepo)
	cottageService := app.NewCottageService(cottageRepo)
	bookingService := app.NewBookingService(availabilityService, cottageService, bookingRepo)

	availabilityHandler := httphandler.NewAvailabilityHandler(availabilityService)
	bookingHandler := httphandler.NewBookingHandler(bookingService)
	cottageHandler := httphandler.NewCottageHandler(cottageService)
	healthHandler := httphandler.NewHandler(mongoDb)

	// service health check endpoints
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Readiness)

	// return the list with details of all cottages
	router.GET("/cottages", cottageHandler.GetAll)
	// return the cottage for the given name
	router.GET("/cottage/:name", cottageHandler.GetByName)

	// return a list of available periods in the given date range
	router.GET("/cottage/:name/available-dates", availabilityHandler.GetAvailablePeriods)
	// return a list of available periods for every cottage of a given type in the given date range
	router.GET("/cottage/type/:cottageType/available-dates", availabilityHandler.GetAvailablePeriodsByCottageType)

	// register a booking for a cottage
	router.POST("/cottage/:name/booking", bookingHandler.AddBooking)
	// cancel a booking for a given cottage
	router.DELETE("/cottage/:name/booking/:bookingId", bookingHandler.RemoveBooking)
}

func ginSetup() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware("cottageManager"))
	router.Use(httphandler.RequestLogger())

	router.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	return router
}

func ginSpinUP(ginConfig config.AppConfig, router *gin.Engine) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", ginConfig.Host, ginConfig.Port),
		Handler: router,
	}

	go func() {
		slog.Info("http server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
		}
	}()

	return srv
}

func mongoDbSetup(config config.MongoConfig) *db.Db {
	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startCancel()

	mongoUri := fmt.Sprintf("mongodb://%s:%s@%s", config.Username, config.Password, config.Host)

	mongoDb, err := db.NewMongoDb(startCtx, mongoUri, config.Database)
	if err != nil {
		slog.ErrorContext(startCtx, "failed to connect to MongoDB", "err", err)
		os.Exit(1)
	}

	slog.InfoContext(startCtx, "connected to MongoDB")

	return mongoDb
}
