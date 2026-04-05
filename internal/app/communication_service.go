package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/port"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const guestRoutingKeyPrefix = "guest."

type CommunicationService interface {
	SendBookingConfirmation()
}

type communicationService struct {
	PaymentConfirmationConsumer port.MqConsumer
	BookingConfirmationProducer port.MqProducer
	BookingRepo                 port.BookingRepo
}

func NewCommunicationService(producer port.MqProducer, consumer port.MqConsumer, bookingRepo port.BookingRepo) CommunicationService {
	return &communicationService{
		PaymentConfirmationConsumer: consumer,
		BookingConfirmationProducer: producer,
		BookingRepo:                 bookingRepo,
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
		if err := proto.Unmarshal(delivery.Body, paymentConfirmation); err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal payment confirmation message", "error", err)
			continue
		}

		bookingID, err := bson.ObjectIDFromHex(paymentConfirmation.GetBookingId())
		if err != nil {
			slog.ErrorContext(ctx, "invalid booking id in payment confirmation", "error", err, "booking_id", paymentConfirmation.GetBookingId())
			continue
		}
		if err := c.BookingRepo.UpdateStatus(ctx, bookingID, string(domain.BookingStatusConfirmed)); err != nil {
			slog.ErrorContext(ctx, "failed to update booking status from payment confirmation", "error", err, "booking_id", paymentConfirmation.GetBookingId())
			continue
		}

		booking, err := c.BookingRepo.GetBooking(ctx, bookingID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to load booking after payment confirmation", "error", err, "booking_id", paymentConfirmation.GetBookingId())
			continue
		}

		bookingConfirmation := buildBookingConfirmation(paymentConfirmation, booking)

		routingKey := fmt.Sprintf("%s%s", guestRoutingKeyPrefix, booking.MainGuest.Hex())
		if err := c.BookingConfirmationProducer.Publish(ctx, bookingConfirmation, routingKey); err != nil {
			slog.ErrorContext(ctx, "failed to publish booking confirmation message", "error", err, "booking_id", bookingConfirmation.GetBookingId())
			continue
		}

		slog.InfoContext(ctx, "booking confirmation published", "booking_id", bookingConfirmation.GetBookingId(), "routing_key", routingKey)
	}
}

func buildBookingConfirmation(paymentConfirmation *dto.PaymentConfirmation, booking document.Booking) *dto.BookingConfirmedNotificationEvent {
	return &dto.BookingConfirmedNotificationEvent{
		Id:            paymentConfirmation.GetId(),
		BookingId:     paymentConfirmation.GetBookingId(),
		BookingStatus: dto.BookingStatus_BOOKING_STATUS_CONFIRMED,
		Guest: &dto.Guest{
			GuestId: booking.MainGuest.Hex(),
		},
		Booking: &dto.BookingSummary{
			CottageName:    booking.CottageName,
			CheckIn:        timestamppb.New(booking.StayPeriod.CheckIn),
			CheckOut:       timestamppb.New(booking.StayPeriod.CheckOut),
			Nights:         int32(booking.StayPeriod.CheckOut.Sub(booking.StayPeriod.CheckIn).Hours() / 24),
			NumberOfGuests: int32(booking.NumberOfGuests),
		},
		ConfirmedAt: paymentConfirmation.GetConfirmedAt(),
	}
}
