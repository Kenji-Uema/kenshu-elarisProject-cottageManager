package db

import (
	"context"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
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

func Test_cottageRepo_GetByType(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &cottageRepo{collection: ct}

		t.Run("existing type", func(t *testing.T) {
			got, err := r.GetByType(ctx, "Luxury")
			if err != nil {
				t.Fatalf("GetByType() unexpected error: %v", err)
			}
			if len(got) != 2 {
				t.Errorf("GetByType() count = %d, want %d", len(got), 2)
			}
		})

		t.Run("non-existing type", func(t *testing.T) {
			got, err := r.GetByType(ctx, "Nonexistent")
			if err != nil {
				t.Fatalf("GetByType() unexpected error: %v", err)
			}
			if len(got) != 0 {
				t.Errorf("GetByType() count = %d, want %d", len(got), 0)
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
			if got.Name != "Rose" || got.Details.View != "Sea" {
				t.Errorf("GetByName() got = %+v, want name Rose and details view Sea", got)
			}
		})

		t.Run("non-existing cottage", func(t *testing.T) {
			got, err := r.GetByName(ctx, "Nonexistent")
			if err == nil {
				t.Errorf("GetByName() expected error for missing doc, got nil")
			}
			if !reflect.DeepEqual(got, domain.Cottage{}) {
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
}

func Test_cottageRepo_DeleteBooking(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &cottageRepo{collection: ct}

		newB := bson.NewObjectID()
		if err := r.DeleteBooking(ctx, "Daisy", newB); err != nil {
			t.Fatalf("AddBooking() unexpected error: %v", err)
		}

		daisyCottage, err := r.GetByName(ctx, "Daisy")

		if err != nil {
			t.Fatalf("GetByName() unexpected error: %v", err)
		}

		if slices.Contains(daisyCottage.Bookings, newB) {
			t.Errorf("DeleteBooking() did not remove booking from Daisy")
		}
	})
}
