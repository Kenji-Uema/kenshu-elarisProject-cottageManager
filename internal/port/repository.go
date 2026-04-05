package port

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TransactionManager interface {
	WithTransaction(ctx context.Context, callback func(ctx context.Context) (any, error)) (any, error)
}

type BookingRepo interface {
	GetBookings(ctx context.Context, ids []bson.ObjectID) ([]document.Booking, error)
	GetBooking(ctx context.Context, id bson.ObjectID) (document.Booking, error)
	HasOverlappingBooking(ctx context.Context, cottageName string, period domain.Period) (bool, error)
	AddBooking(ctx context.Context, booking document.Booking) (bson.ObjectID, error)
	UpdateStatus(ctx context.Context, id bson.ObjectID, status string) error
	DeleteBooking(ctx context.Context, id bson.ObjectID) error
}

type CottageRepo interface {
	GetAll(ctx context.Context) ([]document.Cottage, error)
	GetByName(ctx context.Context, name string) (document.Cottage, error)
	GetByView(ctx context.Context, cottageType string) ([]document.Cottage, error)
	GetBookingsId(ctx context.Context, cottageName string) ([]bson.ObjectID, error)
	AddBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error
	DeleteBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error
}
