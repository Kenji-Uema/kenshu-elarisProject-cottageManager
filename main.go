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
	"github.com/Kenji-Uema/cottageManager/internal/infra/mq"
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

	shutdownTelemetry, err := telemetry.Init(ctx, configs.AppConfig.Telemetry, configs.AppConfig)
	exitOnError(ctx, "failed to setup telemetry", err)

	mongoDb, err := db.NewMongoDbFromConfig(ctx, configs.MongoConfig.Conn)
	exitOnError(ctx, "failed to connect to MongoDB", err)

	rabbitMqClient, err := mq.NewRabbitMqConnection(ctx, configs.RabbitMqConfig.Conn)
	exitOnError(ctx, "failed to connect to RabbitMQ", err)

	cottageRepo := db.NewCottageRepo(mongoDb.Database, configs.MongoConfig.Collections.Cottage)
	bookingRepo := db.NewBookingRepo(mongoDb.Database, configs.MongoConfig.Collections.Booking)
	txManager := db.NewMongoTxManager(mongoDb.Client)

	invoiceProducer, err := mq.NewRabbitmqProducer(rabbitMqClient, configs.RabbitMqConfig.Publishers.CreateInvoice.Publish)
	exitOnError(ctx, "failed to create invoice producer", err)
	err = invoiceProducer.DeclareExchange(configs.RabbitMqConfig.Publishers.CreateInvoice.Exchange)
	exitOnError(ctx, "failed to declare invoice exchange", err)

	notificationProducer, err := mq.NewRabbitmqProtoProducer(rabbitMqClient, configs.RabbitMqConfig.Publishers.BookingConfirmation.Publish)
	exitOnError(ctx, "failed to create notification producer", err)
	err = notificationProducer.DeclareExchange(configs.RabbitMqConfig.Publishers.BookingConfirmation.Exchange)
	exitOnError(ctx, "failed to declare notification exchange", err)

	paymentConsumer, err := mq.NewRabbitmqConsumer(rabbitMqClient, configs.RabbitMqConfig.Consumers.PaymentConfirmed.Consume)
	exitOnError(ctx, "failed to create payment consumer", err)
	err = paymentConsumer.DeclareQueue(ctx, configs.RabbitMqConfig.Consumers.PaymentConfirmed.Queue)
	exitOnError(ctx, "failed to declare payment queue", err)
	err = paymentConsumer.BindQueue(ctx, configs.RabbitMqConfig.Consumers.PaymentConfirmed.Binding)
	exitOnError(ctx, "failed to bind payment queue", err)

	cottageService := app.NewCottageService(cottageRepo)
	bookingService := app.NewBookingService(cottageService, bookingRepo, txManager, invoiceProducer)
	availabilityService := app.NewAvailabilityService(cottageService, bookingService)
	communicationService := app.NewCommunicationService(notificationProducer, paymentConsumer, bookingRepo)

	go communicationService.SendBookingConfirmation()

	httpServer := http.NewHttpServer(configs.AppConfig.Server)
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

	if err := rabbitMqClient.Close(); err != nil {
		slog.ErrorContext(ctx, "close rabbitmq connection", "error", err)
	}

	if err := invoiceProducer.CloseChannel(); err != nil {
		slog.ErrorContext(ctx, "close invoice producer", "error", err)
	}

	if err := notificationProducer.CloseChannel(); err != nil {
		slog.ErrorContext(ctx, "close notification producer", "error", err)
	}

	if err := paymentConsumer.CloseChannel(); err != nil {
		slog.ErrorContext(ctx, "close payment consumer", "error", err)
	}
}

func exitOnError(ctx context.Context, errMsg string, err error) {
	if err != nil {
		slog.ErrorContext(ctx, errMsg, "error", err)
		os.Exit(1)
	}
}
