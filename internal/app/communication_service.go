package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/infra/telemetry"
	"github.com/Kenji-Uema/cottageManager/internal/port"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.opentelemetry.io/otel/attribute"
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
		deliveryCtx, deliverySpan := telemetry.StartConsumerSpan(ctx, delivery, "payment-confirmed")
		processCtx, processSpan := telemetry.StartAppSpan(deliveryCtx, "CommunicationService.SendBookingConfirmation.Process")

		paymentConfirmation := &dto.PaymentConfirmation{}
		unmarshalCtx, unmarshalSpan := telemetry.StartAppSpan(processCtx, "CommunicationService.SendBookingConfirmation.UnmarshalPaymentConfirmation")
		if err := proto.Unmarshal(delivery.Body, paymentConfirmation); err != nil {
			telemetry.RecordSpanError(unmarshalSpan, err, "payment_confirmation_unmarshal_failed")
			telemetry.RecordDeliveryError(deliverySpan, err)
			slog.ErrorContext(unmarshalCtx, "failed to unmarshal payment confirmation message", "error", err)
			c.nackDelivery(unmarshalCtx, delivery, false)
			unmarshalSpan.End()
			processSpan.End()
			deliverySpan.End()
			continue
		}
		unmarshalSpan.End()

		bookingID, err := bson.ObjectIDFromHex(paymentConfirmation.GetBookingId())
		if err != nil {
			telemetry.RecordSpanError(processSpan, err, "payment_confirmation_invalid_booking_id")
			telemetry.RecordDeliveryError(deliverySpan, err)
			slog.ErrorContext(processCtx, "invalid booking id in payment confirmation", "error", err, "booking_id", paymentConfirmation.GetBookingId())
			c.nackDelivery(processCtx, delivery, false)
			processSpan.End()
			deliverySpan.End()
			continue
		}
		processSpan.SetAttributes(attribute.String("booking.id", bookingID.Hex()))

		updateCtx, updateSpan := telemetry.StartAppSpan(processCtx, "CommunicationService.SendBookingConfirmation.UpdateBookingStatus")
		if err := c.BookingRepo.UpdateStatus(updateCtx, bookingID, string(domain.BookingStatusConfirmed)); err != nil {
			telemetry.RecordSpanError(updateSpan, err, "booking_repo.update_status_failed")
			telemetry.RecordSpanError(processSpan, err, "booking_repo.update_status_failed")
			telemetry.RecordDeliveryError(deliverySpan, err)
			slog.ErrorContext(updateCtx, "failed to update booking status from payment confirmation", "error", err, "booking_id", paymentConfirmation.GetBookingId())
			c.nackDelivery(updateCtx, delivery, true)
			updateSpan.End()
			processSpan.End()
			deliverySpan.End()
			continue
		}
		updateSpan.End()

		loadCtx, loadSpan := telemetry.StartAppSpan(processCtx, "CommunicationService.SendBookingConfirmation.LoadBooking")
		booking, err := c.BookingRepo.GetBooking(loadCtx, bookingID)
		if err != nil {
			telemetry.RecordSpanError(loadSpan, err, "booking_repo.get_booking_failed")
			telemetry.RecordSpanError(processSpan, err, "booking_repo.get_booking_failed")
			telemetry.RecordDeliveryError(deliverySpan, err)
			slog.ErrorContext(loadCtx, "failed to load booking after payment confirmation", "error", err, "booking_id", paymentConfirmation.GetBookingId())
			c.nackDelivery(loadCtx, delivery, true)
			loadSpan.End()
			processSpan.End()
			deliverySpan.End()
			continue
		}
		loadSpan.End()

		buildCtx, buildSpan := telemetry.StartAppSpan(processCtx, "CommunicationService.SendBookingConfirmation.BuildMessage")
		bookingConfirmation := buildBookingConfirmation(paymentConfirmation, booking)
		buildSpan.End()

		routingKey := fmt.Sprintf("%s%s", guestRoutingKeyPrefix, booking.MainGuest.Hex())
		publishCtx, publishSpan := telemetry.StartAppSpan(buildCtx, "CommunicationService.SendBookingConfirmation.PublishBookingConfirmation")
		publishSpan.SetAttributes(
			attribute.String("booking.id", bookingConfirmation.GetBookingId()),
			attribute.String("messaging.rabbitmq.routing_key", routingKey),
		)
		if err := c.BookingConfirmationProducer.Publish(publishCtx, bookingConfirmation, routingKey); err != nil {
			telemetry.RecordSpanError(publishSpan, err, "mq.publish_booking_confirmation_failed")
			telemetry.RecordSpanError(processSpan, err, "mq.publish_booking_confirmation_failed")
			telemetry.RecordDeliveryError(deliverySpan, err)
			slog.ErrorContext(publishCtx, "failed to publish booking confirmation message", "error", err, "booking_id", bookingConfirmation.GetBookingId())
			c.nackDelivery(publishCtx, delivery, true)
			publishSpan.End()
			processSpan.End()
			deliverySpan.End()
			continue
		}
		publishSpan.End()

		slog.InfoContext(processCtx, "booking confirmation published", "booking_id", bookingConfirmation.GetBookingId(), "routing_key", routingKey)
		c.ackDelivery(processCtx, delivery)
		processSpan.End()
		deliverySpan.End()
	}
}

func (c communicationService) ackDelivery(ctx context.Context, delivery amqp.Delivery) {
	if err := delivery.Ack(false); err != nil {
		slog.ErrorContext(ctx, "failed to ack payment confirmation delivery", "delivery_tag", delivery.DeliveryTag, "error", err)
	}
}

func (c communicationService) nackDelivery(ctx context.Context, delivery amqp.Delivery, requeue bool) {
	if err := delivery.Nack(false, requeue); err != nil {
		slog.ErrorContext(ctx, "failed to nack payment confirmation delivery", "delivery_tag", delivery.DeliveryTag, "requeue", requeue, "error", err)
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
