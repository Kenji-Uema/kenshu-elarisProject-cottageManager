package app

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/app/fakes"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	fakes2 "github.com/Kenji-Uema/cottageManager/internal/port/fakes"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func validBooking(id bson.ObjectID) domain.Booking {
	return domain.Booking{
		Id:             id,
		MainGuest:      bson.NewObjectID(),
		NumberOfGuests: 2,
		StayPeriod: domain.Period{
			CheckIn:  time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			CheckOut: time.Date(2025, 9, 5, 0, 0, 0, 0, time.UTC),
		},
		CottageName: "A1",
		Status:      "PENDING",
	}
}

func Test_bookingService_GetBookings(t *testing.T) {
	t.Run("success returns mapped bookings", func(t *testing.T) {
		repo := fakes2.NewFakeBookingRepo()
		bookingId := bson.NewObjectID()
		bookingDoc := validBooking(bookingId).ToDocument()
		repo.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]document.Booking, error) {
			return []document.Booking{bookingDoc}, nil
		}

		svc := NewBookingService(nil, repo, nil)
		got, err := svc.GetBookings(context.Background(), []bson.ObjectID{bookingId})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantBooking, _ := domain.NewBookingFromDocument(bookingDoc)
		want := []domain.Booking{wantBooking}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("missing bookings maps to corrupted data error", func(t *testing.T) {
		repo := fakes2.NewFakeBookingRepo()
		repo.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]document.Booking, error) {
			return nil, &dbErrors.MissingBookingsErr{Missing: []bson.ObjectID{bson.NewObjectID()}}
		}

		svc := NewBookingService(nil, repo, nil)
		_, err := svc.GetBookings(context.Background(), []bson.ObjectID{bson.NewObjectID()})
		var expectedErr *appErrors.CorruptedDataError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected CorruptedDataError, got %v", err)
		}
	})
}

func Test_bookingService_AddBooking(t *testing.T) {
	t.Run("success returns booking id", func(t *testing.T) {
		repo := fakes2.NewFakeBookingRepo()
		tx := fakes2.NewFakeTransactionManager()
		cs := fakes.NewFakeCottageService()
		booking := validBooking(bson.NewObjectID())

		id := bson.NewObjectID()
		repo.AddBookingFunc = func(_ context.Context, _ document.Booking) (bson.ObjectID, error) {
			return id, nil
		}
		cs.AddBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		svc := NewBookingService(cs, repo, tx)
		got, err := svc.AddBooking(context.Background(), booking)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != id {
			t.Fatalf("expected %s, got %s", id.Hex(), got.Hex())
		}
	})

	t.Run("cottage not found from cottage service passes through", func(t *testing.T) {
		repo := fakes2.NewFakeBookingRepo()
		tx := fakes2.NewFakeTransactionManager()
		cs := fakes.NewFakeCottageService()
		booking := validBooking(bson.NewObjectID())

		repo.AddBookingFunc = func(_ context.Context, _ document.Booking) (bson.ObjectID, error) {
			return bson.NewObjectID(), nil
		}
		cs.AddBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return &appErrors.CottageNotFound{Err: &dbErrors.CottageNotFoundErr{CottageName: "A1"}}
		}

		svc := NewBookingService(cs, repo, tx)
		_, err := svc.AddBooking(context.Background(), booking)
		var expectedErr *appErrors.CottageNotFound
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected CottageNotFound, got %v", err)
		}
	})

	t.Run("unexpected repo error is wrapped", func(t *testing.T) {
		repo := fakes2.NewFakeBookingRepo()
		tx := fakes2.NewFakeTransactionManager()
		cs := fakes.NewFakeCottageService()
		booking := validBooking(bson.NewObjectID())
		repoErr := errors.New("db error")

		repo.AddBookingFunc = func(_ context.Context, _ document.Booking) (bson.ObjectID, error) {
			return bson.NilObjectID, repoErr
		}

		svc := NewBookingService(cs, repo, tx)
		_, err := svc.AddBooking(context.Background(), booking)
		var expectedErr *appErrors.UnexpectedError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected UnexpectedError, got %v", err)
		}
	})
}

func Test_bookingService_RemoveBooking(t *testing.T) {
	t.Run("booking not found is mapped", func(t *testing.T) {
		repo := fakes2.NewFakeBookingRepo()
		tx := fakes2.NewFakeTransactionManager()
		cs := fakes.NewFakeCottageService()
		bookingId := bson.NewObjectID()

		repo.DeleteBookingFunc = func(_ context.Context, _ bson.ObjectID) error {
			return &dbErrors.BookingNotFoundErr{BookingId: bookingId}
		}

		svc := NewBookingService(cs, repo, tx)
		err := svc.RemoveBooking(context.Background(), "A1", bookingId)
		var expectedErr *appErrors.BookingNotFound
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected BookingNotFound, got %v", err)
		}
	})

	t.Run("success returns nil", func(t *testing.T) {
		repo := fakes2.NewFakeBookingRepo()
		tx := fakes2.NewFakeTransactionManager()
		cs := fakes.NewFakeCottageService()
		bookingId := bson.NewObjectID()

		repo.DeleteBookingFunc = func(_ context.Context, _ bson.ObjectID) error {
			return nil
		}
		cs.RemoveBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		svc := NewBookingService(cs, repo, tx)
		if err := svc.RemoveBooking(context.Background(), "A1", bookingId); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
