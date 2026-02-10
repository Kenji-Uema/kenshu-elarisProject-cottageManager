package port

import (
	"context"
	"cottageManager/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingRepo interface {
	GetBookings(ctx context.Context, ids []primitive.ObjectID) ([]domain.Booking, error)
	AddBooking(ctx context.Context, booking domain.Booking) (primitive.ObjectID, error)
	DeleteBooking(ctx context.Context, id primitive.ObjectID) (bool, error)
}

type CottageRepo interface {
	GetAll(ctx context.Context) ([]domain.Cottage, error)
	GetByName(ctx context.Context, name string) (domain.Cottage, error)
	GetByType(ctx context.Context, cottageType string) ([]domain.Cottage, error)
	GetBookingsId(ctx context.Context, name string) ([]primitive.ObjectID, error)
	AddBooking(ctx context.Context, name string, bookingId primitive.ObjectID) error
	DeleteBooking(ctx context.Context, name string, bookingId primitive.ObjectID) error
}
