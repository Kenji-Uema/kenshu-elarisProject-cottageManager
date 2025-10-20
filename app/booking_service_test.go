package app

import (
	"context"
	appmocks "cottageManager/app/mocks"
	"cottageManager/domain"
	"cottageManager/internal/appErrors"
	"cottageManager/internal/dbErrors"
	portmocks "cottageManager/port/mocks"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var initBookingServiceMocks = func(ctrl *gomock.Controller) (*appmocks.MockAvailabilityService, *appmocks.MockCottageService, *portmocks.MockBookingRepo) {
	am := appmocks.NewMockAvailabilityService(ctrl)
	cm := appmocks.NewMockCottageService(ctrl)
	br := portmocks.NewMockBookingRepo(ctrl)

	return am, cm, br
}

func Test_bookingService_AddBooking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("when cottage is not free, should return error", func(t *testing.T) {
		am, _, _ := initBookingServiceMocks(ctrl)

		am.EXPECT().IsCottageAvailable(gomock.Any(), "A1", gomock.Any()).Return(false, nil)
		// No further calls expected

		svc := NewBookingService(am, nil, nil)
		_, err := svc.AddBooking(context.Background(), "A1", domain.Booking{})

		var expectedErr *appErrors.CottageNotAvailableError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("when IsCottageAvailable returns error, should propagate", func(t *testing.T) {
		am, _, _ := initBookingServiceMocks(ctrl)

		expErr := appErrors.CottageNotAvailableUnexpectedError{Err: fmt.Errorf("cottage not available")}
		am.EXPECT().IsCottageAvailable(gomock.Any(), "A1", domain.Booking{}.StayPeriod).Return(false, expErr)

		svc := NewBookingService(am, nil, nil)
		_, err := svc.AddBooking(context.Background(), "A1", domain.Booking{})

		if !errors.Is(err, expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when AddBooking in repo fails, should propagate", func(t *testing.T) {
		am, _, br := initBookingServiceMocks(ctrl)

		am.EXPECT().IsCottageAvailable(gomock.Any(), "A1", domain.Period{}).Return(true, nil)

		expErr := dbErrors.UnexpectedError{Err: errors.New("db error")}
		br.EXPECT().AddBooking(gomock.Any(), domain.Booking{}).Return(primitive.NilObjectID, &expErr)

		svc := NewBookingService(am, nil, br)
		_, err := svc.AddBooking(context.Background(), "A1", domain.Booking{})

		if !errors.Is(err, &expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when cottageService.AddBooking fails, should propagate", func(t *testing.T) {
		am, cm, br := initBookingServiceMocks(ctrl)

		am.EXPECT().IsCottageAvailable(gomock.Any(), "A1", domain.Period{}).Return(true, nil)

		id := primitive.NewObjectID()
		br.EXPECT().AddBooking(gomock.Any(), domain.Booking{}).Return(id, nil)

		expErr := dbErrors.UnexpectedError{Err: errors.New("db error")}
		cm.EXPECT().AddBooking(gomock.Any(), "A1", id).Return(&expErr)

		svc := NewBookingService(am, cm, br)
		_, err := svc.AddBooking(context.Background(), "A1", domain.Booking{})
		if !errors.Is(err, &expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when all succeed, should return booking id hex", func(t *testing.T) {
		am, cm, br := initBookingServiceMocks(ctrl)

		am.EXPECT().IsCottageAvailable(gomock.Any(), "A1", domain.Period{}).Return(true, nil)

		id := primitive.NewObjectIDFromTimestamp(time.Now())
		br.EXPECT().AddBooking(gomock.Any(), domain.Booking{}).Return(id, nil)

		cm.EXPECT().AddBooking(gomock.Any(), "A1", id).Return(nil)

		svc := NewBookingService(am, cm, br)

		got, err := svc.AddBooking(context.Background(), "A1", domain.Booking{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != id.Hex() {
			t.Fatalf("expected %s, got %s", id.Hex(), got)
		}
	})
}

func Test_bookingService_RemoveBooking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("when bookingRepo.DeleteBooking fails, should propagate", func(t *testing.T) {
		_, _, br := initBookingServiceMocks(ctrl)

		expErr := dbErrors.UnexpectedError{Err: errors.New("db error")}
		br.EXPECT().DeleteBooking(gomock.Any(), gomock.Any()).Return(false, &expErr)

		svc := NewBookingService(nil, nil, br)

		err := svc.RemoveBooking(context.Background(), "A1", primitive.NewObjectIDFromTimestamp(time.Now()))
		if !errors.Is(err, &expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when cottageService.RemoveBooking fails, should propagate", func(t *testing.T) {
		_, cm, br := initBookingServiceMocks(ctrl)

		br.EXPECT().DeleteBooking(gomock.Any(), gomock.Any()).Return(true, nil)

		expErr := dbErrors.UnexpectedError{Err: errors.New("db error")}
		cm.EXPECT().RemoveBooking(gomock.Any(), "A1", gomock.Any()).Return(&expErr)

		svc := NewBookingService(nil, cm, br)

		err := svc.RemoveBooking(context.Background(), "A1", primitive.NewObjectIDFromTimestamp(time.Now()))
		if !errors.Is(err, &expErr) {
			t.Fatalf("expected %v, got %v", expErr, err)
		}
	})

	t.Run("when all succeed, should return nil", func(t *testing.T) {
		_, cm, br := initBookingServiceMocks(ctrl)

		br.EXPECT().DeleteBooking(gomock.Any(), gomock.Any()).Return(true, nil)
		cm.EXPECT().RemoveBooking(gomock.Any(), "A1", gomock.Any()).Return(nil)

		svc := NewBookingService(nil, cm, br)
		if err := svc.RemoveBooking(context.Background(), "A1", primitive.NewObjectIDFromTimestamp(time.Now())); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
