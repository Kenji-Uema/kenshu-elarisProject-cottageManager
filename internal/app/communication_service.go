package app

import (
	"context"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/port"
	"google.golang.org/protobuf/encoding/protojson"
)

const bookingConfirmedRoutingKey = "booking.confirmed"

type CommunicationService interface {
	SendBookingConfirmation()
}

type communicationService struct {
	PaymentConfirmationConsumer port.MqConsumer
	BookingConfirmationProducer port.MqProducer
}

func NewCommunicationService(producer port.MqProducer, consumer port.MqConsumer) CommunicationService {
	return &communicationService{
		PaymentConfirmationConsumer: consumer,
		BookingConfirmationProducer: producer,
	}
}

func (c communicationService) SendBookingConfirmation() {
	ctx := context.Background()
	deliveries, err := c.PaymentConfirmationConsumer.Consume(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to start payment confirmation consumer", "error", err)
		return
	}

	for delivery := range deliveries {
		paymentConfirmation := &dto.PaymentConfirmation{}
		if err := protojson.Unmarshal(delivery.Body, paymentConfirmation); err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal payment confirmation message", "error", err)
			continue
		}

		bookingConfirmation := &dto.BookingConfirmedNotificationEvent{
			Id:            paymentConfirmation.GetId(),
			BookingId:     paymentConfirmation.GetBookingId(),
			BookingStatus: dto.BookingStatus_BOOKING_STATUS_CONFIRMED,
			ConfirmedAt:   paymentConfirmation.GetConfirmedAt(),
		}

		if err := c.BookingConfirmationProducer.Publish(ctx, bookingConfirmation, bookingConfirmedRoutingKey); err != nil {
			slog.ErrorContext(ctx, "failed to publish booking confirmation message", "error", err, "booking_id", bookingConfirmation.GetBookingId())
			continue
		}

		slog.InfoContext(ctx, "booking confirmation published", "booking_id", bookingConfirmation.GetBookingId())
	}
}
