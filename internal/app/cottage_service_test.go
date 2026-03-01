package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	"github.com/Kenji-Uema/cottageManager/internal/port/fakes"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func validCottageDoc(name string) document.Cottage {
	return document.Cottage{
		Id:            bson.NewObjectID(),
		Name:          name,
		View:          "sea",
		Photos:        []string{"a.jpg"},
		PricePerNight: 10,
		Bookings: []bson.ObjectID{
			bson.NewObjectID(),
		},
		Details: document.CottageDetails{
			Description:          "desc",
			FurnitureDescription: "furniture",
			BathroomDescription:  "bathroom",
			AmenitiesDescription: "amenities",
		},
	}
}

func Test_cottageService_GetAll(t *testing.T) {
	t.Run("success returns all cottages", func(t *testing.T) {
		repo := fakes.NewFakeCottageRepo()
		docA := validCottageDoc("A1")
		docB := validCottageDoc("B1")
		repo.GetAllFunc = func(_ context.Context) ([]document.Cottage, error) {
			return []document.Cottage{docA, docB}, nil
		}

		wantA, _ := domain.NewCottageFromDoc(docA)
		wantB, _ := domain.NewCottageFromDoc(docB)
		want := []domain.Cottage{wantA, wantB}

		svc := NewCottageService(repo)
		got, err := svc.GetAll(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("repo error maps to unexpected error", func(t *testing.T) {
		repo := fakes.NewFakeCottageRepo()
		repo.GetAllFunc = func(_ context.Context) ([]document.Cottage, error) {
			return nil, errors.New("db down")
		}

		svc := NewCottageService(repo)
		_, err := svc.GetAll(context.Background())
		var expectedErr *appErrors.UnexpectedError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected UnexpectedError, got %v", err)
		}
	})
}

func Test_cottageService_GetByName(t *testing.T) {
	t.Run("success returns cottage", func(t *testing.T) {
		repo := fakes.NewFakeCottageRepo()
		doc := validCottageDoc("A1")
		repo.GetByNameFunc = func(_ context.Context, _ string) (document.Cottage, error) {
			return doc, nil
		}

		want, _ := domain.NewCottageFromDoc(doc)
		svc := NewCottageService(repo)
		got, err := svc.GetByName(context.Background(), "A1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("cottage not found maps to app error", func(t *testing.T) {
		repo := fakes.NewFakeCottageRepo()
		repo.GetByNameFunc = func(_ context.Context, _ string) (document.Cottage, error) {
			return document.Cottage{}, &dbErrors.CottageNotFoundErr{CottageName: "A1"}
		}

		svc := NewCottageService(repo)
		_, err := svc.GetByName(context.Background(), "A1")
		var expectedErr *appErrors.CottageNotFound
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected CottageNotFound, got %v", err)
		}
	})
}

func Test_cottageService_AddBooking(t *testing.T) {
	t.Run("success returns nil", func(t *testing.T) {
		repo := fakes.NewFakeCottageRepo()
		id := bson.NewObjectID()
		repo.AddBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		svc := NewCottageService(repo)
		if err := svc.AddBooking(context.Background(), "A1", id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("bookings not updated maps to add booking app error", func(t *testing.T) {
		repo := fakes.NewFakeCottageRepo()
		id := bson.NewObjectID()
		repo.AddBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return &dbErrors.BookingsNotUpdatedErr{CottageName: "A1", BookingId: id}
		}

		svc := NewCottageService(repo)
		err := svc.AddBooking(context.Background(), "A1", id)
		var expectedErr *appErrors.AddBookingToCottageError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected AddBookingToCottageError, got %v", err)
		}
	})
}

func Test_cottageService_RemoveBooking(t *testing.T) {
	t.Run("success returns nil", func(t *testing.T) {
		repo := fakes.NewFakeCottageRepo()
		id := bson.NewObjectID()
		repo.DeleteBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		svc := NewCottageService(repo)
		if err := svc.RemoveBooking(context.Background(), "A1", id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("bookings not updated maps to remove booking app error", func(t *testing.T) {
		repo := fakes.NewFakeCottageRepo()
		id := bson.NewObjectID()
		repo.DeleteBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return &dbErrors.BookingsNotUpdatedErr{CottageName: "A1", BookingId: id}
		}

		svc := NewCottageService(repo)
		err := svc.RemoveBooking(context.Background(), "A1", id)
		var expectedErr *appErrors.RemoveBookingFromCottageError
		if !errors.As(err, &expectedErr) {
			t.Fatalf("expected RemoveBookingFromCottageError, got %v", err)
		}
	})
}
