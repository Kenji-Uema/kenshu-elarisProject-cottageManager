package app

import (
	"context"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	mqfakes "github.com/Kenji-Uema/cottageManager/internal/infra/mq/fakes"
	repfakes "github.com/Kenji-Uema/cottageManager/internal/port/fakes"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCommunicationService_SendBookingConfirmation(t *testing.T) {
	t.Run("publishes booking confirmation from payment confirmation", func(t *testing.T) {
		producer := &mqfakes.FakeMqProducer{}
		consumer := &mqfakes.FakeMqConsumer{}
		bookingRepo := repfakes.NewFakeBookingRepo()
		bookingID := bson.NewObjectID()
		mainGuestID := bson.NewObjectID()
		bookingRepo.GetBookingFunc = func(ctx context.Context, id bson.ObjectID) (document.Booking, error) {
			return document.Booking{
				Id:             bookingID,
				MainGuest:      mainGuestID,
				NumberOfGuests: 2,
				StayPeriod: document.Period{
					CheckIn:  time.Date(2026, 3, 11, 15, 0, 0, 0, time.UTC),
					CheckOut: time.Date(2026, 3, 14, 11, 0, 0, 0, time.UTC),
				},
				CottageName: "Rose Cottage",
				Status:      string(domain.BookingStatusConfirmed),
			}, nil
		}

		confirmedAt := timestamppb.New(time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))
		payment := &dto.PaymentConfirmation{
			Id:          "evt-1",
			BookingId:   bookingID.Hex(),
			PayerId:     "payer-1",
			ConfirmedAt: confirmedAt,
		}

		body, err := proto.Marshal(payment)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		consumer.ConsumeFn = func(_ context.Context) (<-chan amqp.Delivery, error) {
			ch := make(chan amqp.Delivery, 1)
			ch <- amqp.Delivery{Body: body}
			close(ch)
			return ch, nil
		}

		svc := NewCommunicationService(producer, consumer, bookingRepo)
		svc.SendBookingConfirmation()

		if bookingRepo.UpdateStatusCalls != 1 {
			t.Fatalf("expected booking status update once, got %d", bookingRepo.UpdateStatusCalls)
		}
		if bookingRepo.LastUpdatedBookingID != bookingID {
			t.Fatalf("expected updated booking id %s, got %s", bookingID.Hex(), bookingRepo.LastUpdatedBookingID.Hex())
		}
		if bookingRepo.LastUpdatedStatus != string(domain.BookingStatusConfirmed) {
			t.Fatalf("expected updated status %q, got %q", domain.BookingStatusConfirmed, bookingRepo.LastUpdatedStatus)
		}
		if producer.PublishCallCount != 1 {
			t.Fatalf("expected publish to be called once, got %d", producer.PublishCallCount)
		}
		if producer.LastPublishedRoutingKey != "guest."+mainGuestID.Hex() {
			t.Fatalf("expected routing key %q, got %q", "guest."+mainGuestID.Hex(), producer.LastPublishedRoutingKey)
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
		if msg.GetGuest().GetGuestId() != mainGuestID.Hex() {
			t.Fatalf("expected guest id %q, got %q", mainGuestID.Hex(), msg.GetGuest().GetGuestId())
		}
		if msg.GetBooking().GetCottageName() != "Rose Cottage" {
			t.Fatalf("expected cottage name %q, got %q", "Rose Cottage", msg.GetBooking().GetCottageName())
		}
		if msg.GetConfirmedAt().AsTime() != confirmedAt.AsTime() {
			t.Fatalf("expected confirmed_at %v, got %v", confirmedAt.AsTime(), msg.GetConfirmedAt().AsTime())
		}
	})

	t.Run("invalid message payload is skipped", func(t *testing.T) {
		producer := &mqfakes.FakeMqProducer{}
		consumer := &mqfakes.FakeMqConsumer{}
		bookingRepo := repfakes.NewFakeBookingRepo()

		consumer.ConsumeFn = func(_ context.Context) (<-chan amqp.Delivery, error) {
			ch := make(chan amqp.Delivery, 1)
			ch <- amqp.Delivery{Body: []byte(`{"invalid_json"`)}
			close(ch)
			return ch, nil
		}

		svc := NewCommunicationService(producer, consumer, bookingRepo)
		svc.SendBookingConfirmation()

		if bookingRepo.UpdateStatusCalls != 0 {
			t.Fatalf("expected booking status to not be updated, got %d", bookingRepo.UpdateStatusCalls)
		}
		if producer.PublishCallCount != 0 {
			t.Fatalf("expected publish to not be called, got %d", producer.PublishCallCount)
		}
	})

	t.Run("invalid booking id is skipped", func(t *testing.T) {
		producer := &mqfakes.FakeMqProducer{}
		consumer := &mqfakes.FakeMqConsumer{}
		bookingRepo := repfakes.NewFakeBookingRepo()

		payment := &dto.PaymentConfirmation{
			Id:        "evt-1",
			BookingId: "not-a-hex-id",
		}

		body, err := proto.Marshal(payment)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		consumer.ConsumeFn = func(_ context.Context) (<-chan amqp.Delivery, error) {
			ch := make(chan amqp.Delivery, 1)
			ch <- amqp.Delivery{Body: body}
			close(ch)
			return ch, nil
		}

		svc := NewCommunicationService(producer, consumer, bookingRepo)
		svc.SendBookingConfirmation()

		if bookingRepo.UpdateStatusCalls != 0 {
			t.Fatalf("expected booking status to not be updated, got %d", bookingRepo.UpdateStatusCalls)
		}
		if producer.PublishCallCount != 0 {
			t.Fatalf("expected publish to not be called, got %d", producer.PublishCallCount)
		}
	})
}
