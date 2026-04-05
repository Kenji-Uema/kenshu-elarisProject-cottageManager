package fakes

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FakeBookingRepo struct {
	GetBookingsFunc           func(ctx context.Context, ids []bson.ObjectID) ([]document.Booking, error)
	GetBookingFunc            func(ctx context.Context, id bson.ObjectID) (document.Booking, error)
	HasOverlappingBookingFunc func(ctx context.Context, cottageName string, period domain.Period) (bool, error)
	AddBookingFunc            func(ctx context.Context, booking document.Booking) (bson.ObjectID, error)
	UpdateStatusFunc          func(ctx context.Context, id bson.ObjectID, status string) error
	DeleteBookingFunc         func(ctx context.Context, id bson.ObjectID) error

	GetBookingsCalls           int
	GetBookingCalls            int
	HasOverlappingBookingCalls int
	AddBookingCalls            int
	UpdateStatusCalls          int
	DeleteBookingCalls         int

	LastUpdatedBookingID bson.ObjectID
	LastUpdatedStatus    string
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

func (f *FakeBookingRepo) GetBooking(ctx context.Context, id bson.ObjectID) (document.Booking, error) {
	f.GetBookingCalls++
	if f.GetBookingFunc != nil {
		return f.GetBookingFunc(ctx, id)
	}
	return document.Booking{}, nil
}

func (f *FakeBookingRepo) HasOverlappingBooking(ctx context.Context, cottageName string, period domain.Period) (bool, error) {
	f.HasOverlappingBookingCalls++
	if f.HasOverlappingBookingFunc != nil {
		return f.HasOverlappingBookingFunc(ctx, cottageName, period)
	}
	return false, nil
}

func (f *FakeBookingRepo) AddBooking(ctx context.Context, booking document.Booking) (bson.ObjectID, error) {
	f.AddBookingCalls++
	if f.AddBookingFunc != nil {
		return f.AddBookingFunc(ctx, booking)
	}
	return bson.NilObjectID, nil
}

func (f *FakeBookingRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, status string) error {
	f.UpdateStatusCalls++
	f.LastUpdatedBookingID = id
	f.LastUpdatedStatus = status
	if f.UpdateStatusFunc != nil {
		return f.UpdateStatusFunc(ctx, id, status)
	}
	return nil
}

func (f *FakeBookingRepo) DeleteBooking(ctx context.Context, id bson.ObjectID) error {
	f.DeleteBookingCalls++
	if f.DeleteBookingFunc != nil {
		return f.DeleteBookingFunc(ctx, id)
	}
	return nil
}
