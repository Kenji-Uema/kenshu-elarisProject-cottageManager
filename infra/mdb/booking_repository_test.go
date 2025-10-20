package mdb

import (
	"context"
	"cottageManager/domain"
	"cottageManager/internal/dbErrors"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Test_bookingRepo_AddBooking(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &bookingRepo{collection: br}

		t.Run("when booking added successfully, should return inserted BookingId", func(t *testing.T) {
			booking := domain.Booking{Id: primitive.NewObjectID()}

			got, err := r.AddBooking(ctx, booking)
			if err != nil {
				t.Errorf("AddBooking() error = %v", err)
			}
			if !reflect.DeepEqual(got, booking.Id) {
				t.Errorf("AddBooking() got = %v, want %v", got, booking.Id)
			}
		})

		t.Run("when booking already exists, should return error", func(t *testing.T) {
			objectId, _ := primitive.ObjectIDFromHex("86a7d3bb21169a393dd1db1b")
			booking := domain.Booking{Id: objectId}

			_, err := r.AddBooking(ctx, booking)

			var unexpectedError *dbErrors.UnexpectedError
			if !errors.As(err, &unexpectedError) {
				t.Errorf("AddBooking() error = %v", err)
			}
		})
	})
}

func Test_bookingCrudRepo_DeleteBooking(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &bookingRepo{collection: br}

		t.Run("when booking deleted successfully, should return true", func(t *testing.T) {
			objectId, _ := primitive.ObjectIDFromHex("86a7d3bb21169a393dd1db1b")
			got, err := r.DeleteBooking(ctx, objectId)

			if err != nil {
				t.Errorf("DeleteBooking() error = %v", err)
			}
			if !reflect.DeepEqual(got, true) {
				t.Errorf("DeleteBooking() got = %v, want %v", got, true)
			}
		})

		t.Run("when try to delete booking that does not exits, should return false", func(t *testing.T) {
			got, err := r.DeleteBooking(ctx, primitive.NewObjectIDFromTimestamp(time.Now()))

			if err != nil {
				t.Errorf("DeleteBooking() error = %v", err)
			}
			if !reflect.DeepEqual(got, false) {
				t.Errorf("DeleteBooking() got = %v, want %v", got, false)
			}
		})
	})
}

func Test_bookingRepo_GetBookings(t *testing.T) {
	setupAndRun(t, func(t *testing.T, ct *mongo.Collection, br *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		r := &bookingRepo{collection: br}

		t.Run("when list of ids is empty, should return error", func(t *testing.T) {
			_, err := r.GetBookings(ctx, []primitive.ObjectID{})

			var target *dbErrors.ValidationError
			if !errors.As(err, &target) {
				t.Errorf("unexpected error type: %v", err)
			}
		})

		t.Run("when searched bookingId does not exist, should return error", func(t *testing.T) {
			_, err := r.GetBookings(ctx, []primitive.ObjectID{primitive.NewObjectIDFromTimestamp(time.Now())})

			var target *dbErrors.MissingBookingsError
			if !errors.As(err, &target) {
				t.Errorf("unexpected error type: %v", err)
			}
		})

		t.Run("when bookingId_A exists but bookingId_B does not, should return error when searching for both", func(t *testing.T) {
			id, _ := primitive.ObjectIDFromHex("86a7d3bb21169a393dd1db1b")
			missingId := primitive.NewObjectIDFromTimestamp(time.Now())

			_, err := r.GetBookings(ctx, []primitive.ObjectID{id, missingId})

			var target *dbErrors.MissingBookingsError
			if !errors.As(err, &target) {
				t.Errorf("unexpected error type: %v", err)
			}
		})

		t.Run("when both bookingId_A and bookingId_B exists, should return both", func(t *testing.T) {
			idA, _ := primitive.ObjectIDFromHex("86a7d3bb21169a393dd1db1b")
			idB, _ := primitive.ObjectIDFromHex("b5fa4aefa370638801dbe2af")

			got, err := r.GetBookings(ctx, []primitive.ObjectID{idA, idB})

			if err != nil {
				t.Errorf("unexpected error type: %v", err)
			}

			gotIds := make([]primitive.ObjectID, len(got))
			for i, b := range got {
				gotIds[i] = b.Id
			}

			if !slices.Contains(gotIds, idA) && slices.Contains(gotIds, idB) {
				t.Errorf("unexpected list of returned bookings")
			}
		})
	})
}
