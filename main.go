package main

import (
	"context"
	"cottageManager/app"
	"cottageManager/infra/logging"
	"cottageManager/infra/mdb"
	"cottageManager/infra/telemetry"
	"cottageManager/transport/http/availability"
	"cottageManager/transport/http/booking"
	"cottageManager/transport/http/cottage"
	"cottageManager/transport/http/health"
	"cottageManager/transport/http/middleware"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
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

	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startCancel()

	mongoDb, err := mdb.NewMongoDb(startCtx, "mongodb://admin:admin123@localhost:32017", "CeladonHotel")
	if err != nil {
		logger.Error("failed to connect to MongoDB", "err", err)
		os.Exit(1)
	}

	cottageRepo := mdb.NewCottageRepo(mongoDb.Database)
	bookingRepo := mdb.NewBookingRepo(mongoDb.Database)

	availabilityService := app.NewAvailabilityService(cottageRepo, bookingRepo)
	cottageService := app.NewCottageService(cottageRepo)
	bookingService := app.NewBookingService(availabilityService, cottageService, bookingRepo)

	availabilityHandler := availability.NewHandler(availabilityService)
	bookingHandler := booking.NewHandler(bookingService)
	cottageHandler := cottage.NewHandler(cottageService)
	healthHandler := health.NewHandler(mongoDb)

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

	// service health check endpoints
	router.GET("/healthz", healthHandler.Health)
	router.GET("/readyz", healthHandler.Readiness)

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

	srv := &http.Server{
		Addr:    "localhost:8080",
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
