package app

import (
	"context"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	mqfakes "github.com/Kenji-Uema/cottageManager/internal/infra/mq/fakes"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCommunicationService_SendBookingConfirmation(t *testing.T) {
	t.Run("publishes booking confirmation from payment confirmation", func(t *testing.T) {
		producer := &mqfakes.FakeMqProducer{}
		consumer := &mqfakes.FakeMqConsumer{}

		confirmedAt := timestamppb.New(time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))
		payment := &dto.PaymentConfirmation{
			Id:          "evt-1",
			BookingId:   "booking-1",
			PayerId:     "payer-1",
			ConfirmedAt: confirmedAt,
		}

		body, err := protojson.Marshal(payment)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		consumer.ConsumeFn = func(_ context.Context) (<-chan amqp.Delivery, error) {
			ch := make(chan amqp.Delivery, 1)
			ch <- amqp.Delivery{Body: body}
			close(ch)
			return ch, nil
		}

		svc := NewCommunicationService(producer, consumer)
		svc.SendBookingConfirmation()

		if producer.PublishCallCount != 1 {
			t.Fatalf("expected publish to be called once, got %d", producer.PublishCallCount)
		}
		if producer.LastPublishedRoutingKey != bookingConfirmedRoutingKey {
			t.Fatalf("expected routing key %q, got %q", bookingConfirmedRoutingKey, producer.LastPublishedRoutingKey)
		}

		msg, ok := producer.LastPublishedMessage.(*dto.BookingConfirmedNotificationEvent)
		if !ok {
			t.Fatalf("expected BookingConfirmedNotificationEvent, got %T", producer.LastPublishedMessage)
		}
		if msg.GetId() != payment.GetId() {
			t.Fatalf("expected id %q, got %q", payment.GetId(), msg.GetId())
		}
		if msg.GetBookingId() != payment.GetBookingId() {
			t.Fatalf("expected booking id %q, got %q", payment.GetBookingId(), msg.GetBookingId())
		}
		if msg.GetBookingStatus() != dto.BookingStatus_BOOKING_STATUS_CONFIRMED {
			t.Fatalf("expected booking status CONFIRMED, got %v", msg.GetBookingStatus())
		}
		if msg.GetConfirmedAt().AsTime() != confirmedAt.AsTime() {
			t.Fatalf("expected confirmed_at %v, got %v", confirmedAt.AsTime(), msg.GetConfirmedAt().AsTime())
		}
	})

	t.Run("invalid message payload is skipped", func(t *testing.T) {
		producer := &mqfakes.FakeMqProducer{}
		consumer := &mqfakes.FakeMqConsumer{}

		consumer.ConsumeFn = func(_ context.Context) (<-chan amqp.Delivery, error) {
			ch := make(chan amqp.Delivery, 1)
			ch <- amqp.Delivery{Body: []byte(`{"invalid_json"`)}
			close(ch)
			return ch, nil
		}

		svc := NewCommunicationService(producer, consumer)
		svc.SendBookingConfirmation()

		if producer.PublishCallCount != 0 {
			t.Fatalf("expected publish to not be called, got %d", producer.PublishCallCount)
		}
	})
}
