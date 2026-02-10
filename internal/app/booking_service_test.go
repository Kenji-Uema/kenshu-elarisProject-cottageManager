package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/app/mocks"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	portmocks "github.com/Kenji-Uema/cottageManager/internal/port/mocks"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var initBookingServiceMocks = func() (*mocks.MockAvailabilityService, *mocks.MockCottageService, *portmocks.MockBookingRepo) {
	am := mocks.NewMockAvailabilityService()
	cm := mocks.NewMockCottageService()
	br := portmocks.NewMockBookingRepo()

	return am, cm, br
}

func Test_bookingService_AddBooking(t *testing.T) {
	t.Run("when cottage is not free, should return error", func(t *testing.T) {
		am, _, _ := initBookingServiceMocks()

		booking := domain.Booking{CottageName: "A1"}

		am.IsCottageAvailableFunc = func(_ context.Context, _ string, _ domain.Period) (bool, error) {
			return false, nil
		}
		// No further calls expected

		svc := NewBookingService(am, nil, nil)
		_, err := svc.AddBooking(context.Background(), booking)

		var expectedErr *appErrors.CottageNotAvailableError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("when IsCottageAvailable returns error, should propagate", func(t *testing.T) {
		am, _, _ := initBookingServiceMocks()

		booking := domain.Booking{CottageName: "A1"}
		expErr := appErrors.CottageNotAvailableUnexpectedError{Err: fmt.Errorf("cottage not available")}
		am.IsCottageAvailableFunc = func(_ context.Context, _ string, _ domain.Period) (bool, error) {
			return false, expErr
		}

		svc := NewBookingService(am, nil, nil)
		_, err := svc.AddBooking(context.Background(), booking)

		if !errors.Is(err, expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when AddBooking in repo fails, should propagate", func(t *testing.T) {
		am, _, br := initBookingServiceMocks()

		booking := domain.Booking{CottageName: "A1"}

		am.IsCottageAvailableFunc = func(_ context.Context, _ string, _ domain.Period) (bool, error) {
			return true, nil
		}

		expErr := dbErrors.UnexpectedError{Err: errors.New("db error")}
		br.AddBookingFunc = func(_ context.Context, _ domain.Booking) (bson.ObjectID, error) {
			return bson.NilObjectID, &expErr
		}

		svc := NewBookingService(am, nil, br)
		_, err := svc.AddBooking(context.Background(), booking)

		if !errors.Is(err, &expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when cottageService.AddBooking fails, should propagate", func(t *testing.T) {
		am, cm, br := initBookingServiceMocks()

		booking := domain.Booking{CottageName: "A1"}

		am.IsCottageAvailableFunc = func(_ context.Context, _ string, _ domain.Period) (bool, error) {
			return true, nil
		}

		id := bson.NewObjectID()
		br.AddBookingFunc = func(_ context.Context, _ domain.Booking) (bson.ObjectID, error) {
			return id, nil
		}

		expErr := dbErrors.UnexpectedError{Err: errors.New("db error")}
		cm.AddBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return &expErr
		}

		svc := NewBookingService(am, cm, br)
		_, err := svc.AddBooking(context.Background(), booking)
		if !errors.Is(err, &expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when all succeed, should return booking id hex", func(t *testing.T) {
		am, cm, br := initBookingServiceMocks()

		booking := domain.Booking{CottageName: "A1"}

		am.IsCottageAvailableFunc = func(_ context.Context, _ string, _ domain.Period) (bool, error) {
			return true, nil
		}

		id := bson.NewObjectIDFromTimestamp(time.Now())
		br.AddBookingFunc = func(_ context.Context, _ domain.Booking) (bson.ObjectID, error) {
			return id, nil
		}

		cm.AddBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		svc := NewBookingService(am, cm, br)

		got, err := svc.AddBooking(context.Background(), booking)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != id.Hex() {
			t.Fatalf("expected %s, got %s", id.Hex(), got)
		}
	})
}

func Test_bookingService_RemoveBooking(t *testing.T) {
	t.Run("when bookingRepo.DeleteBooking fails, should propagate", func(t *testing.T) {
		_, _, br := initBookingServiceMocks()

		expErr := dbErrors.UnexpectedError{Err: errors.New("db error")}
		br.DeleteBookingFunc = func(_ context.Context, _ bson.ObjectID) (bool, error) {
			return false, &expErr
		}

		svc := NewBookingService(nil, nil, br)

		err := svc.RemoveBooking(context.Background(), "A1", bson.NewObjectIDFromTimestamp(time.Now()))
		if !errors.Is(err, &expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when cottageService.RemoveBooking fails, should propagate", func(t *testing.T) {
		_, cm, br := initBookingServiceMocks()

		br.DeleteBookingFunc = func(_ context.Context, _ bson.ObjectID) (bool, error) {
			return true, nil
		}

		expErr := dbErrors.UnexpectedError{Err: errors.New("db error")}
		cm.RemoveBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return &expErr
		}

		svc := NewBookingService(nil, cm, br)

		err := svc.RemoveBooking(context.Background(), "A1", bson.NewObjectIDFromTimestamp(time.Now()))
		if !errors.Is(err, &expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when all succeed, should return nil", func(t *testing.T) {
		_, cm, br := initBookingServiceMocks()

		br.DeleteBookingFunc = func(_ context.Context, _ bson.ObjectID) (bool, error) {
			return true, nil
		}
		cm.RemoveBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		svc := NewBookingService(nil, cm, br)
		if err := svc.RemoveBooking(context.Background(), "A1", bson.NewObjectIDFromTimestamp(time.Now())); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
