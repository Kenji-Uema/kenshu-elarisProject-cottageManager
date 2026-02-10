package port

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingRepo interface {
	GetBookings(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error)
	AddBooking(ctx context.Context, booking domain.Booking) (bson.ObjectID, error)
	DeleteBooking(ctx context.Context, id bson.ObjectID) (bool, error)
}

type CottageRepo interface {
	GetAll(ctx context.Context) ([]domain.Cottage, error)
	GetByName(ctx context.Context, name string) (domain.Cottage, error)
	GetByType(ctx context.Context, cottageType string) ([]domain.Cottage, error)
	GetBookingsId(ctx context.Context, name string) ([]bson.ObjectID, error)
	AddBooking(ctx context.Context, name string, bookingId bson.ObjectID) error
	DeleteBooking(ctx context.Context, name string, bookingId bson.ObjectID) error
}
