package app

import (
	"context"
	"errors"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	portmocks "github.com/Kenji-Uema/cottageManager/internal/port/mocks"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_cottageService_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success: returns all cottages", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo(ctrl)
		want := []domain.Cottage{{Name: "A1"}, {Name: "B1"}}

		repo.EXPECT().GetAll(gomock.Any()).Return(want, nil)

		svc := NewCottageService(repo)
		got, err := svc.GetAll(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("error: return cottageNotFound error", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo(ctrl)
		dbError := dbErrors.UnexpectedError{Err: errors.New("db error")}

		repo.EXPECT().GetAll(gomock.Any()).Return(nil, &dbError)

		svc := NewCottageService(repo)
		_, err := svc.GetAll(context.Background())

		var expectedErr *appErrors.UnexpectedError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func Test_cottageService_GetByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success: returns cottage", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo(ctrl)
		want := domain.Cottage{Name: "A1"}

		repo.EXPECT().GetByName(gomock.Any(), "A1").Return(want, nil)

		svc := NewCottageService(repo)
		got, err := svc.GetByName(context.Background(), "A1")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("error: returns cottageNotFound error", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo(ctrl)

		dbError := errors.New("db error")
		repo.EXPECT().GetByName(gomock.Any(), "A1").Return(domain.Cottage{}, dbError)

		svc := NewCottageService(repo)
		_, err := svc.GetByName(context.Background(), "A1")

		var expectedErr *appErrors.CottageNotFound
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func Test_cottageService_AddBooking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success: should return nil", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo(ctrl)
		id := primitive.NewObjectID()

		repo.EXPECT().AddBooking(gomock.Any(), "A1", id).Return(nil)

		svc := NewCottageService(repo)
		if err := svc.AddBooking(context.Background(), "A1", id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error: should Return AddBookingToCottageError", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo(ctrl)
		id := primitive.NewObjectID()

		dbError := dbErrors.UnexpectedError{Err: errors.New("db error")}
		repo.EXPECT().AddBooking(gomock.Any(), "A1", id).Return(&dbError)

		svc := NewCottageService(repo)
		err := svc.AddBooking(context.Background(), "A1", id)

		var expectedErr *appErrors.AddBookingToCottageError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func Test_cottageService_RemoveBooking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success: should return nil", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo(ctrl)
		id := primitive.NewObjectID()

		repo.EXPECT().DeleteBooking(gomock.Any(), "A1", id).Return(nil)

		svc := NewCottageService(repo)
		if err := svc.RemoveBooking(context.Background(), "A1", id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error: should return RemoveBookingFromCottage error", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo(ctrl)
		id := primitive.NewObjectID()

		dbError := dbErrors.UnexpectedError{Err: errors.New("db error")}
		repo.EXPECT().DeleteBooking(gomock.Any(), "A1", id).Return(&dbError)

		svc := NewCottageService(repo)
		err := svc.RemoveBooking(context.Background(), "A1", id)

		var expectedErr *appErrors.RemoveBookingFromCottageError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}
