package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	portmocks "github.com/Kenji-Uema/cottageManager/internal/port/mocks"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func Test_cottageService_GetAll(t *testing.T) {
	t.Run("success: returns all cottages", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo()
		want := []domain.Cottage{{Name: "A1"}, {Name: "B1"}}

		repo.GetAllFunc = func(_ context.Context) ([]domain.Cottage, error) {
			return want, nil
		}

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
		repo := portmocks.NewMockCottageRepo()
		dbError := dbErrors.UnexpectedError{Err: errors.New("db error")}

		repo.GetAllFunc = func(_ context.Context) ([]domain.Cottage, error) {
			return nil, &dbError
		}

		svc := NewCottageService(repo)
		_, err := svc.GetAll(context.Background())

		var expectedErr *appErrors.UnexpectedError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func Test_cottageService_GetByName(t *testing.T) {
	t.Run("success: returns cottage", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo()
		want := domain.Cottage{Name: "A1"}

		repo.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return want, nil
		}

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
		repo := portmocks.NewMockCottageRepo()

		dbError := errors.New("db error")
		repo.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, dbError
		}

		svc := NewCottageService(repo)
		_, err := svc.GetByName(context.Background(), "A1")

		var expectedErr *appErrors.CottageNotFound
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func Test_cottageService_AddBooking(t *testing.T) {
	t.Run("success: should return nil", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo()
		id := bson.NewObjectID()

		repo.AddBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		svc := NewCottageService(repo)
		if err := svc.AddBooking(context.Background(), "A1", id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error: should Return AddBookingToCottageError", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo()
		id := bson.NewObjectID()

		dbError := dbErrors.UnexpectedError{Err: errors.New("db error")}
		repo.AddBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return &dbError
		}

		svc := NewCottageService(repo)
		err := svc.AddBooking(context.Background(), "A1", id)

		var expectedErr *appErrors.AddBookingToCottageError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func Test_cottageService_RemoveBooking(t *testing.T) {
	t.Run("success: should return nil", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo()
		id := bson.NewObjectID()

		repo.DeleteBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		svc := NewCottageService(repo)
		if err := svc.RemoveBooking(context.Background(), "A1", id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error: should return RemoveBookingFromCottage error", func(t *testing.T) {
		repo := portmocks.NewMockCottageRepo()
		id := bson.NewObjectID()

		dbError := dbErrors.UnexpectedError{Err: errors.New("db error")}
		repo.DeleteBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return &dbError
		}

		svc := NewCottageService(repo)
		err := svc.RemoveBooking(context.Background(), "A1", id)

		var expectedErr *appErrors.RemoveBookingFromCottageError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}
