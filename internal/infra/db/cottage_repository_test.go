package db

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	"go.mongodb.org/mongo-driver/v2/bson"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Test_cottageRepo_GetAll(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &cottageRepo{collection: ct}
		got, err := r.GetAll(ctx)
		if err != nil {
			t.Fatalf("GetAll() unexpected error: %v", err)
		}

		if len(got) != 3 {
			t.Fatalf("GetAll() length = %d, want %d", len(got), 3)
		}
	})
}

func Test_cottageRepo_GetByView(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &cottageRepo{collection: ct}

		t.Run("existing view", func(t *testing.T) {
			got, err := r.GetByView(ctx, "luxury")
			if err != nil {
				t.Fatalf("GetByView() unexpected error: %v", err)
			}
			if len(got) != 2 {
				t.Errorf("GetByView() count = %d, want %d", len(got), 2)
			}
			for _, cottage := range got {
				if cottage.View != "luxury" {
					t.Errorf("GetByView() returned cottage with view = %q, want %q", cottage.View, "luxury")
				}
			}
		})

		t.Run("non-existing view", func(t *testing.T) {
			got, err := r.GetByView(ctx, "Nonexistent")
			if err != nil {
				t.Fatalf("GetByView() unexpected error: %v", err)
			}
			if len(got) != 0 {
				t.Errorf("GetByView() count = %d, want %d", len(got), 0)
			}
		})
	})
}

func Test_cottageRepo_GetByName(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &cottageRepo{collection: ct}

		t.Run("existing cottage", func(t *testing.T) {
			got, err := r.GetByName(ctx, "Rose")
			if err != nil {
				t.Fatalf("GetByName() unexpected error: %v", err)
			}
			if got.Name != "Rose" || got.View != "luxury" {
				t.Errorf("GetByName() got = %+v, want name Rose and view luxury", got)
			}
			if got.CleaningStatus != document.CleaningStatusFullyCleaned {
				t.Errorf("GetByName() cleaning status = %q, want %q", got.CleaningStatus, document.CleaningStatusFullyCleaned)
			}
			if got.Key.Holder != document.KeyHolderCottage {
				t.Errorf("GetByName() key holder = %q, want %q", got.Key.Holder, document.KeyHolderCottage)
			}
		})

		t.Run("non-existing cottage", func(t *testing.T) {
			got, err := r.GetByName(ctx, "Nonexistent")
			if err == nil {
				t.Errorf("GetByName() expected error for missing doc, got nil")
			}
			if !reflect.DeepEqual(got, document.Cottage{}) {
				t.Errorf("GetByName() got = %+v, want empty cottage", got)
			}
		})
	})
}

func Test_cottageRepo_GetBookingsId(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &cottageRepo{collection: ct}

		got, err := r.GetBookingsId(ctx, "Rose")
		if err != nil {
			t.Fatalf("GetBookingsId() unexpected error: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("GetBookingsId() length = %d, want %d", len(got), 3)
		}
	})
}

func Test_cottageRepo_AddBooking(t *testing.T) {
	t.Run("add booking to existing cottage succeeds", func(t *testing.T) {
		setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			r := &cottageRepo{collection: ct}

			newB := bson.NewObjectID()
			if err := r.AddBooking(ctx, "Daisy", newB); err != nil {
				t.Fatalf("AddBooking() unexpected error: %v", err)
			}

			daisyCottage, err := r.GetByName(ctx, "Daisy")

			if err != nil {
				t.Fatalf("GetByName() unexpected error: %v", err)
			}

			if !slices.Contains(daisyCottage.Bookings, newB) {
				t.Errorf("AddBooking() did not append booking to Daisy")
			}
		})
	})

	t.Run("strict duplicate add returns error", func(t *testing.T) {
		setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			r := &cottageRepo{collection: ct}

			existing, err := bson.ObjectIDFromHex("86a7d3bb21169a393dd1db1b")
			if err != nil {
				t.Fatalf("ObjectIDFromHex() unexpected error: %v", err)
			}
			err = r.AddBooking(ctx, "Rose", existing)
			if err == nil {
				t.Fatalf("AddBooking() expected error for duplicate booking id, got nil")
			}
			var target *dbErrors.BookingsNotUpdatedErr
			if !errors.As(err, &target) {
				t.Fatalf("AddBooking() error type = %T, want %T", err, target)
			}
		})
	})

	t.Run("missing cottage returns not found error", func(t *testing.T) {
		setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			r := &cottageRepo{collection: ct}

			err := r.AddBooking(ctx, "MissingCottage", bson.NewObjectID())
			if err == nil {
				t.Fatalf("AddBooking() expected error for missing cottage, got nil")
			}
			var target *dbErrors.CottageNotFoundErr
			if !errors.As(err, &target) {
				t.Fatalf("AddBooking() error type = %T, want %T", err, target)
			}
		})
	})
}

func Test_cottageRepo_DeleteBooking(t *testing.T) {
	t.Run("delete missing booking returns not updated error", func(t *testing.T) {
		setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			r := &cottageRepo{collection: ct}

			newB := bson.NewObjectID()
			err := r.DeleteBooking(ctx, "Daisy", newB)
			if err == nil {
				t.Fatalf("DeleteBooking() expected error when booking is missing, got nil")
			}
			var target *dbErrors.BookingsNotUpdatedErr
			if !errors.As(err, &target) {
				t.Fatalf("DeleteBooking() error type = %T, want %T", err, target)
			}
		})
	})

	t.Run("delete existing booking succeeds", func(t *testing.T) {
		setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			r := &cottageRepo{collection: ct}

			existing, err := bson.ObjectIDFromHex("86a7d3bb21169a393dd1db1b")
			if err != nil {
				t.Fatalf("ObjectIDFromHex() unexpected error: %v", err)
			}
			if err := r.DeleteBooking(ctx, "Rose", existing); err != nil {
				t.Fatalf("DeleteBooking() unexpected error: %v", err)
			}

			roseCottage, err := r.GetByName(ctx, "Rose")

			if err != nil {
				t.Fatalf("GetByName() unexpected error: %v", err)
			}

			if slices.Contains(roseCottage.Bookings, existing) {
				t.Errorf("DeleteBooking() did not remove booking from Rose")
			}
		})
	})

	t.Run("missing cottage returns not found error", func(t *testing.T) {
		setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			r := &cottageRepo{collection: ct}

			err := r.DeleteBooking(ctx, "MissingCottage", bson.NewObjectID())
			if err == nil {
				t.Fatalf("DeleteBooking() expected error for missing cottage, got nil")
			}
			var target *dbErrors.CottageNotFoundErr
			if !errors.As(err, &target) {
				t.Fatalf("DeleteBooking() error type = %T, want %T", err, target)
			}
		})
	})
}
