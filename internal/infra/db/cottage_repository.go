package db

import (
	"context"
	"cottageManager/internal/config"
	"cottageManager/internal/domain"
	"cottageManager/internal/domain/errors/appErrors"
	"cottageManager/internal/domain/errors/dbErrors"
	"cottageManager/internal/port"
	"errors"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cottageRepo struct {
	collection *mongo.Collection
}

func NewCottageRepo(db *mongo.Database, config config.CottageCollectionConfig) port.CottageRepo {
	return &cottageRepo{collection: db.Collection(config.Name)}
}

func (r *cottageRepo) GetAll(ctx context.Context) ([]domain.Cottage, error) {
	cur, err := r.collection.Find(ctx, bson.D{})
	defer func() {
		if err := cur.Close(ctx); err != nil {
			slog.Warn("failed to close mongo cursor", "error", err)
		}
	}()

	if err != nil {
		slog.Error("mongo find failed", "error", err)
		return nil, &dbErrors.UnexpectedError{Err: err}
	}

	var cottages []domain.Cottage
	for cur.Next(ctx) {
		var r domain.Cottage
		if err := cur.Decode(&r); err != nil {
			slog.Error("failed to decode cottage document", "error", err)
			return nil, &dbErrors.UnexpectedError{Err: err}
		}
		cottages = append(cottages, r)
	}

	if err := cur.Err(); err != nil {
		slog.Error("cursor iteration failed", "error", err)
		return nil, &dbErrors.UnexpectedError{Err: err}
	}

	return cottages, nil
}

func (r *cottageRepo) GetByType(ctx context.Context, cottageType string) ([]domain.Cottage, error) {
	filter := bson.M{"view": cottageType}

	cur, err := r.collection.Find(ctx, filter)
	defer func() {
		if err := cur.Close(ctx); err != nil {
			slog.Warn("failed to close mongo cursor", "error", err)
		}
	}()

	if err != nil {
		slog.Error("mongo find failed", "error", err, "filter", filter)
		return nil, &dbErrors.UnexpectedError{Err: err}
	}

	var cottages []domain.Cottage
	if err := cur.All(ctx, &cottages); err != nil {
		slog.Error("failed to decode cottages", "error", err, "filter", filter)
		return nil, &dbErrors.UnexpectedError{Err: err}
	}
	return cottages, nil
}

func (r *cottageRepo) GetByName(ctx context.Context, name string) (domain.Cottage, error) {
	filter := bson.M{"name": name}

	var cottage domain.Cottage
	err := r.collection.FindOne(ctx, filter).Decode(&cottage)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Cottage{}, &appErrors.CottageNotFound{Err: err}
		}
		slog.Error("failed to find cottage", "error", err, "filter", filter)
		return domain.Cottage{}, &dbErrors.UnexpectedError{Err: err}
	}

	return cottage, nil
}

func (r *cottageRepo) GetBookingsId(ctx context.Context, name string) ([]primitive.ObjectID, error) {
	var result struct {
		Bookings []primitive.ObjectID `bson:"bookings"`
	}

	err := r.collection.FindOne(
		ctx,
		bson.M{"name": name},
		options.FindOne().SetProjection(bson.M{
			"bookings": 1,
			"_id":      0,
		}),
	).Decode(&result)

	if err != nil {
		return nil, &dbErrors.UnexpectedError{Err: err}
	}

	return result.Bookings, nil
}

func (r *cottageRepo) AddBooking(ctx context.Context, name string, bookingId primitive.ObjectID) error {
	filter := bson.M{"name": name}
	update := bson.M{"$push": bson.M{"bookings": bookingId}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.Error("failed to add booking to cottage", "error", err, "cottage", name, "booking_id", bookingId.Hex())
		return &dbErrors.UnexpectedError{Err: err}
	}

	return nil
}

func (r *cottageRepo) DeleteBooking(ctx context.Context, name string, bookingId primitive.ObjectID) error {
	filter := bson.M{"name": name}
	update := bson.M{"$pull": bson.M{"bookings": bookingId}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.Error("failed to remove booking from cottage", "error", err, "cottage", name, "booking_id", bookingId.Hex())
		return &dbErrors.UnexpectedError{Err: err}
	}

	return nil
}
