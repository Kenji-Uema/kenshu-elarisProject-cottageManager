package main

import (
	"context"
	"cottageManager/internal/app"
	"cottageManager/internal/config"
	"cottageManager/internal/infra/db"
	"cottageManager/internal/infra/logging"
	"cottageManager/internal/infra/telemetry"
	"cottageManager/internal/transport/http/availability"
	"cottageManager/internal/transport/http/booking"
	"cottageManager/internal/transport/http/cottage"
	"cottageManager/internal/transport/http/health"
	"cottageManager/internal/transport/http/middleware"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	ginConfig := config.LoadConfig[config.GinConfig]()
	mongoConfig := config.LoadConfig[config.MongoDbConfig]()
	cottageConfig := config.LoadConfig[config.CottageCollectionConfig]()
	bookingConfig := config.LoadConfig[config.BookingCollectionConfig]()

	logger, telemetryProvider, err := telemetrySetup()

	mongoDb := mongoDbSetup(mongoConfig, err, logger)

	router := ginSetup(telemetryProvider)
	routerSetup(mongoDb, router, cottageConfig, bookingConfig)
	ginSpinUP(ginConfig, router, logger)
}

func routerSetup(mongoDb *db.Db, router *gin.Engine, cotttageConf *config.CottageCollectionConfig, bookingConf *config.BookingCollectionConfig) {
	cottageRepo := db.NewCottageRepo(mongoDb.Database, cotttageConf)
	bookingRepo := db.NewBookingRepo(mongoDb.Database, bookingConf)

	availabilityService := app.NewAvailabilityService(cottageRepo, bookingRepo)
	cottageService := app.NewCottageService(cottageRepo)
	bookingService := app.NewBookingService(availabilityService, cottageService, bookingRepo)

	availabilityHandler := availability.NewHandler(availabilityService)
	bookingHandler := booking.NewHandler(bookingService)
	cottageHandler := cottage.NewHandler(cottageService)
	healthHandler := health.NewHandler(mongoDb)

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
	// delete a booking for a given cottage
	router.DELETE("/cottage/:name/booking/:bookingId", bookingHandler.RemoveBooking)
}

func ginSetup(telemetryProvider telemetry.Provider) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware(telemetryProvider.ServiceName()))
	router.Use(middleware.RequestLogger())

	router.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	return router
}

func ginSpinUP(ginConfig *config.GinConfig, router *gin.Engine, logger *slog.Logger) {
	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%v", ginConfig.Port),
		Handler: router,
	}

	go func() {
		logger.Info("http server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("server exited gracefully")
}

func mongoDbSetup(config *config.MongoDbConfig, err error, logger *slog.Logger) *db.Db {
	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startCancel()

	mongoDb, err := db.NewMongoDb(startCtx,
		fmt.Sprintf("mongodb://%s:%s@%s:%s", config.User, config.Password, config.Url, config.Port), config.Db)
	if err != nil {
		logger.Error("failed to connect to MongoDB", "err", err)
		os.Exit(1)
	}
	return mongoDb
}

func telemetrySetup() (*slog.Logger, telemetry.Provider, error) {
	logger := logging.Setup()

	telemetryProvider, err := telemetry.Setup(context.Background(), logger)
	if err != nil {
		logger.Error("failed to initialise telemetry", "err", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetryProvider.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown telemetry", "err", err)
		}
	}()
	return logger, telemetryProvider, err
}
