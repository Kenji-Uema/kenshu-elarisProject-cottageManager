package fakes

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FakeBookingService struct {
	GetBookingsFunc   func(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error)
	AddBookingFunc    func(ctx context.Context, booking domain.Booking) (bson.ObjectID, error)
	RemoveBookingFunc func(ctx context.Context, cottageName string, bookingId bson.ObjectID) error

	GetBookingsCalls   int
	AddBookingCalls    int
	RemoveBookingCalls int
}

func NewFakeBookingService() *FakeBookingService {
	return &FakeBookingService{}
}

func (f *FakeBookingService) GetBookings(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error) {
	f.GetBookingsCalls++
	if f.GetBookingsFunc != nil {
		return f.GetBookingsFunc(ctx, ids)
	}
	return nil, nil
}

func (f *FakeBookingService) AddBooking(ctx context.Context, booking domain.Booking) (bson.ObjectID, error) {
	f.AddBookingCalls++
	if f.AddBookingFunc != nil {
		return f.AddBookingFunc(ctx, booking)
	}
	return bson.NilObjectID, nil
}

func (f *FakeBookingService) RemoveBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error {
	f.RemoveBookingCalls++
	if f.RemoveBookingFunc != nil {
		return f.RemoveBookingFunc(ctx, cottageName, bookingId)
	}
	return nil
}
