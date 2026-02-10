package db

import (
	"context"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/config"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	"github.com/Kenji-Uema/cottageManager/internal/port"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type bookingRepo struct {
	collection *mongo.Collection
}

func NewBookingRepo(db *mongo.Database, config config.BookingCollectionConfig) port.BookingRepo {
	return &bookingRepo{collection: db.Collection(config.Name)}
}

func (r *bookingRepo) GetBookings(ctx context.Context, ids []primitive.ObjectID) ([]domain.Booking, error) {
	if len(ids) == 0 {
		return []domain.Booking{}, nil
	}

	cur, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	defer func() {
		if err := cur.Close(ctx); err != nil {
			slog.Warn("failed to close mongo cursor", "error", err)
		}
	}()

	if err != nil {
		slog.Error("failed to fetch bookings", "error", err, "ids", ids)
		return nil, &dbErrors.UnexpectedError{Err: err}
	}

	var bookings []domain.Booking
	if err := cur.All(ctx, &bookings); err != nil {
		slog.Error("failed to decode bookings", "error", err)
		return nil, &dbErrors.UnexpectedError{Err: err}
	}

	// Check if all requested IDs were returned
	if len(bookings) != len(ids) {
		// optional: figure out which IDs are missing
		found := make(map[primitive.ObjectID]struct{}, len(bookings))
		for _, b := range bookings {
			found[b.Id] = struct{}{}
		}

		var missing []primitive.ObjectID
		for _, id := range ids {
			if _, ok := found[id]; !ok {
				missing = append(missing, id)
			}
		}

		slog.Warn("missing bookings detected", "missing", missing)
		return nil, &dbErrors.MissingBookingsError{Missing: missing}
	}

	return bookings, nil
}

func (r *bookingRepo) AddBooking(ctx context.Context, booking domain.Booking) (primitive.ObjectID, error) {
	result, err := r.collection.InsertOne(ctx, booking)

	if err != nil {
		slog.Error("failed to insert booking", "error", err)
		return primitive.NilObjectID, &dbErrors.UnexpectedError{Err: err}
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *bookingRepo) DeleteBooking(ctx context.Context, id primitive.ObjectID) (bool, error) {
	deleted, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		slog.Error("failed to delete booking", "error", err, "booking_id", id.Hex())
		return false, &dbErrors.UnexpectedError{Err: err}
	}

	return deleted.DeletedCount != 0, nil
}
