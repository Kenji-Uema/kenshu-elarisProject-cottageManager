package fakes

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FakeCottageRepo struct {
	GetAllFunc        func(ctx context.Context) ([]document.Cottage, error)
	GetByNameFunc     func(ctx context.Context, name string) (document.Cottage, error)
	GetByViewFunc     func(ctx context.Context, view string) ([]document.Cottage, error)
	GetBookingsIdFunc func(ctx context.Context, name string) ([]bson.ObjectID, error)
	AddBookingFunc    func(ctx context.Context, name string, bookingId bson.ObjectID) error
	DeleteBookingFunc func(ctx context.Context, name string, bookingId bson.ObjectID) error

	GetAllCalls        int
	GetByNameCalls     int
	GetByViewCalls     int
	GetBookingsIdCalls int
	AddBookingCalls    int
	DeleteBookingCalls int
}

func NewFakeCottageRepo() *FakeCottageRepo {
	return &FakeCottageRepo{}
}

func (f *FakeCottageRepo) GetAll(ctx context.Context) ([]document.Cottage, error) {
	f.GetAllCalls++
	if f.GetAllFunc != nil {
		return f.GetAllFunc(ctx)
	}
	return nil, nil
}

func (f *FakeCottageRepo) GetByName(ctx context.Context, name string) (document.Cottage, error) {
	f.GetByNameCalls++
	if f.GetByNameFunc != nil {
		return f.GetByNameFunc(ctx, name)
	}
	return document.Cottage{}, nil
}

func (f *FakeCottageRepo) GetByView(ctx context.Context, view string) ([]document.Cottage, error) {
	f.GetByViewCalls++
	if f.GetByViewFunc != nil {
		return f.GetByViewFunc(ctx, view)
	}
	return nil, nil
}

func (f *FakeCottageRepo) GetBookingsId(ctx context.Context, name string) ([]bson.ObjectID, error) {
	f.GetBookingsIdCalls++
	if f.GetBookingsIdFunc != nil {
		return f.GetBookingsIdFunc(ctx, name)
	}
	return nil, nil
}

func (f *FakeCottageRepo) AddBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	f.AddBookingCalls++
	if f.AddBookingFunc != nil {
		return f.AddBookingFunc(ctx, name, bookingId)
	}
	return nil
}

func (f *FakeCottageRepo) DeleteBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	f.DeleteBookingCalls++
	if f.DeleteBookingFunc != nil {
		return f.DeleteBookingFunc(ctx, name, bookingId)
	}
	return nil
}
