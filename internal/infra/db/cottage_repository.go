package db

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/app/validation"
	"github.com/Kenji-Uema/cottageManager/internal/config"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	"github.com/Kenji-Uema/cottageManager/internal/port"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type cottageRepo struct {
	collection *mongo.Collection
}

func NewCottageRepo(db *mongo.Database, config config.CottageCollectionConfig) port.CottageRepo {
	return &cottageRepo{collection: db.Collection(config.Name)}
}

func (r *cottageRepo) GetAll(ctx context.Context) ([]document.Cottage, error) {
	cur, err := r.collection.Find(ctx, bson.D{})
	defer func() {
		if err := cur.Close(ctx); err != nil {
			slog.WarnContext(ctx, "failed to close mongo cursor", "error", err)
		}
	}()

	if err != nil {
		slog.ErrorContext(ctx, "find all cottages failed", "error", err)
		return nil, &dbErrors.UnexpectedErr{Msg: "failed to find all cottages", Err: err}
	}

	var cottages []document.Cottage
	if err := cur.All(ctx, &cottages); err != nil {
		slog.ErrorContext(ctx, "failed to decode to document.Cottage", "error", err)
		return nil, &dbErrors.CorruptedDataErr{Err: err}
	}

	return cottages, nil
}

func (r *cottageRepo) GetByView(ctx context.Context, cottageView string) ([]document.Cottage, error) {
	if err := validation.New().NotBlank("cottageView", cottageView).Validate(); err != nil {
		return nil, err
	}

	filter := bson.M{"view": cottageView}

	cur, err := r.collection.Find(ctx, filter)
	defer func() {
		if err := cur.Close(ctx); err != nil {
			slog.WarnContext(ctx, "failed to close mongo cursor", "error", err)
		}
	}()

	if err != nil {
		slog.ErrorContext(ctx, "mongo find cottages by view failed", "error", err, "filter", filter)
		return nil, &dbErrors.UnexpectedErr{Msg: "mongo find cottages by view failed", Err: err}
	}

	var cottages []document.Cottage
	if err := cur.All(ctx, &cottages); err != nil {
		slog.ErrorContext(ctx, "failed to decode cottages", "error", err, "filter", filter)
		return nil, &dbErrors.CorruptedDataErr{Err: err}
	}
	return cottages, nil
}

func (r *cottageRepo) GetByName(ctx context.Context, name string) (document.Cottage, error) {
	if err := validation.New().NotBlank("cottageName", name).Validate(); err != nil {
		return document.Cottage{}, err
	}

	filter := bson.M{"name": name}

	var cottage document.Cottage
	if err := r.collection.FindOne(ctx, filter).Decode(&cottage); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			slog.WarnContext(ctx, "cottage not found", "filter", filter)
			return document.Cottage{}, &dbErrors.CottageNotFoundErr{CottageName: name}
		}

		slog.ErrorContext(ctx, "failed to decode cottage", "error", err, "filter", filter)
		return document.Cottage{}, &dbErrors.CorruptedDataErr{Err: err}
	}

	return cottage, nil
}

func (r *cottageRepo) GetBookingsId(ctx context.Context, cottageName string) ([]bson.ObjectID, error) {
	if err := validation.New().NotBlank("cottageName", cottageName).Validate(); err != nil {
		return nil, err
	}

	var result struct {
		Bookings []bson.ObjectID `bson:"bookings"`
	}

	findResult := r.collection.FindOne(
		ctx,
		bson.M{"name": cottageName},
		options.FindOne().SetProjection(bson.M{
			"bookings": 1,
		}),
	)

	if err := findResult.Decode(&result); err != nil {
		slog.ErrorContext(ctx, "failed to decode bookings", "error", err)
		return nil, &dbErrors.CorruptedDataErr{Err: err}
	}

	return result.Bookings, nil
}

func (r *cottageRepo) AddBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error {
	if err := validation.New().
		NotBlank("cottageName", cottageName).
		NotNilObjectID("bookingId", bookingId).Validate(); err != nil {
		return err
	}

	filter := bson.M{"name": cottageName}
	update := bson.M{"$addToSet": bson.M{"bookings": bookingId}}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add booking to cottage", "error", err, "cottage", cottageName, "booking_id", bookingId.Hex())
		return &dbErrors.UnexpectedErr{Msg: "failed to add booking to cottage", Err: err}
	}

	if res.MatchedCount == 0 {
		return &dbErrors.CottageNotFoundErr{CottageName: cottageName}
	}

	if res.ModifiedCount == 0 {
		return &dbErrors.BookingsNotUpdatedErr{CottageName: cottageName, BookingId: bookingId}
	}

	return nil
}

func (r *cottageRepo) DeleteBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error {
	if err := validation.New().NotBlank("cottageName", cottageName).
		NotNilObjectID("bookingId", bookingId).
		Validate(); err != nil {
		return err
	}

	filter := bson.M{"name": cottageName}
	update := bson.M{"$pull": bson.M{"bookings": bookingId}}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to remove booking from cottage", "error", err, "cottage", cottageName, "booking_id", bookingId.Hex())
		return &dbErrors.UnexpectedErr{Msg: "failed to remove booking from cottage", Err: err}
	}

	if res.MatchedCount == 0 {
		return &dbErrors.CottageNotFoundErr{CottageName: cottageName}
	}

	if res.ModifiedCount == 0 {
		return &dbErrors.BookingsNotUpdatedErr{CottageName: cottageName, BookingId: bookingId}
	}

	return nil
}
