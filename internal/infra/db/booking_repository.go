package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/app/validation"
	"github.com/Kenji-Uema/cottageManager/internal/config"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	"github.com/Kenji-Uema/cottageManager/internal/port"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type bookingRepo struct {
	collection *mongo.Collection
}

func NewBookingRepo(db *mongo.Database, config config.BookingCollectionConfig) port.BookingRepo {
	return &bookingRepo{collection: db.Collection(config.Name)}
}

func (r *bookingRepo) GetBookings(ctx context.Context, ids []bson.ObjectID) ([]document.Booking, error) {
	if err := validation.New().
		NotZeroValue("ids", ids).
		NoDuplicates("ids", ids).Validate(); err != nil {
		return nil, err
	}

	cur, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	defer func() {
		if err := cur.Close(ctx); err != nil {
			slog.WarnContext(ctx, "failed to close mongo cursor", "error", err)
		}
	}()

	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch bookings", "error", err, "ids", ids)
		return nil, &dbErrors.UnexpectedErr{Msg: "failed to fetch bookings", Err: err}
	}

	var bookings []document.Booking
	if err := cur.All(ctx, &bookings); err != nil {
		slog.ErrorContext(ctx, "failed to decode bookings", "error", err)
		return nil, &dbErrors.CorruptedDataErr{Err: err}
	}

	// Check if all requested IDs were returned
	if len(bookings) != len(ids) {
		found := make(map[bson.ObjectID]struct{}, len(bookings))
		for _, b := range bookings {
			found[b.Id] = struct{}{}
		}

		var missing []bson.ObjectID
		for _, id := range ids {
			if _, ok := found[id]; !ok {
				missing = append(missing, id)
			}
		}

		slog.ErrorContext(ctx, "missing bookings detected", "missing", missing)
		return nil, &dbErrors.MissingBookingsErr{Missing: missing}
	}

	return bookings, nil
}

func (r *bookingRepo) AddBooking(ctx context.Context, booking document.Booking) (bson.ObjectID, error) {
	if err := validation.New().NotZeroValue("booking", booking).Validate(); err != nil {
		return bson.NilObjectID, err
	}

	res, err := r.collection.InsertOne(ctx, booking)

	if err != nil {
		slog.ErrorContext(ctx, "failed to insert booking", "error", err, "booking", booking)
		return bson.NilObjectID, &dbErrors.UnexpectedErr{Msg: "failed to insert booking", Err: err}
	}

	bookingId, ok := res.InsertedID.(bson.ObjectID)
	if !ok {
		return bson.NilObjectID, &dbErrors.UnexpectedErr{
			Msg: "unexpected booking bookingId type",
			Err: fmt.Errorf("insertedID type %T", res.InsertedID),
		}
	}

	return bookingId, nil
}

func (r *bookingRepo) DeleteBooking(ctx context.Context, id bson.ObjectID) error {
	if err := validation.New().NotNilObjectID("id", id).Validate(); err != nil {
		return err
	}

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete booking", "error", err, "booking_id", id.Hex())
		return &dbErrors.UnexpectedErr{Msg: "failed to delete booking", Err: err}
	}

	if res.DeletedCount == 0 {
		return &dbErrors.BookingNotFoundErr{BookingId: id}
	}

	return nil
}
