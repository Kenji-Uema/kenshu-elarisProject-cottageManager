package fakes

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FakeBookingRepo struct {
	GetBookingsFunc   func(ctx context.Context, ids []bson.ObjectID) ([]document.Booking, error)
	AddBookingFunc    func(ctx context.Context, booking document.Booking) (bson.ObjectID, error)
	DeleteBookingFunc func(ctx context.Context, id bson.ObjectID) error

	GetBookingsCalls   int
	AddBookingCalls    int
	DeleteBookingCalls int
}

func NewFakeBookingRepo() *FakeBookingRepo {
	return &FakeBookingRepo{}
}

func (f *FakeBookingRepo) GetBookings(ctx context.Context, ids []bson.ObjectID) ([]document.Booking, error) {
	f.GetBookingsCalls++
	if f.GetBookingsFunc != nil {
		return f.GetBookingsFunc(ctx, ids)
	}
	return nil, nil
}

func (f *FakeBookingRepo) AddBooking(ctx context.Context, booking document.Booking) (bson.ObjectID, error) {
	f.AddBookingCalls++
	if f.AddBookingFunc != nil {
		return f.AddBookingFunc(ctx, booking)
	}
	return bson.NilObjectID, nil
}

func (f *FakeBookingRepo) DeleteBooking(ctx context.Context, id bson.ObjectID) error {
	f.DeleteBookingCalls++
	if f.DeleteBookingFunc != nil {
		return f.DeleteBookingFunc(ctx, id)
	}
	return nil
}
