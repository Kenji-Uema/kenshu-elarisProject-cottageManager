package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/config"
	"github.com/Kenji-Uema/cottageManager/internal/infra/db"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Server struct {
	Router *gin.Engine
	server *http.Server
	config config.ServerConfig
}

func NewHttpServer(config config.ServerConfig) *Server {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware("cottageManager"))

	return &Server{config: config, Router: router}
}

func (s *Server) SetServer() {
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:           s.Router,
		ReadHeaderTimeout: time.Duration(s.config.ReadHeaderTimeoutInSeconds) * time.Second,
		ReadTimeout:       time.Duration(s.config.ReadTimeoutInSeconds) * time.Second,
		WriteTimeout:      time.Duration(s.config.WriteTimeoutInSeconds) * time.Second,
		IdleTimeout:       time.Duration(s.config.IdleTimeoutInSeconds) * time.Second,
	}
	s.server = server
}

func (s *Server) Run(ctx context.Context) {
	slog.InfoContext(ctx, "http server listening", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.ErrorContext(ctx, "server error", "err", err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) SetupRoutes(
	availabilityService app.AvailabilityService,
	bookingService app.BookingService,
	cottageService app.CottageService,
	mongoDb *db.Db) {

	availabilityHandler := NewAvailabilityHandler(availabilityService)
	bookingHandler := NewBookingHandler(bookingService, availabilityService)
	cottageHandler := NewCottageHandler(cottageService)
	probeHandler := NewProbeHandler(mongoDb)

	// bookingService health check endpoints
	s.Router.GET("/healthz", probeHandler.Heath)
	s.Router.GET("/readyz", probeHandler.Ready)

	// return the list with details of all cottages
	s.Router.GET("/cottages", cottageHandler.GetAll)
	// return the cottage for the given name
	s.Router.GET("/cottage/:name", cottageHandler.GetByName)
	// return the list of cottages given the type
	s.Router.GET("/cottage/view/:view", cottageHandler.GetByView)

	// return a list of available periods in the given date range
	s.Router.GET("/cottage/:name/available-dates", availabilityHandler.GetAvailablePeriods)
	// return a list of available periods for every cottage of a given type in the given date range
	s.Router.GET("/cottage/type/:cottageType/available-dates", availabilityHandler.GetAvailablePeriodsByCottageType)

	// register a booking for a cottage
	s.Router.POST("/cottage/:name/booking", bookingHandler.AddBooking)
	// cancel a booking for a given cottage
	s.Router.DELETE("/cottage/:name/booking/:bookingId", bookingHandler.RemoveBooking)

	// swagger
	s.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
