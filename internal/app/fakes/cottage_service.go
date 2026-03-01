package fakes

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FakeCottageService struct {
	GetAllFunc        func(ctx context.Context) ([]domain.Cottage, error)
	GetByNameFunc     func(ctx context.Context, cottageName string) (domain.Cottage, error)
	GetByViewFunc     func(ctx context.Context, view string) ([]domain.Cottage, error)
	AddBookingFunc    func(ctx context.Context, name string, bookingId bson.ObjectID) error
	RemoveBookingFunc func(ctx context.Context, name string, bookingId bson.ObjectID) error

	GetAllCalls        int
	GetByNameCalls     int
	GetByViewCalls     int
	AddBookingCalls    int
	RemoveBookingCalls int
}

func NewFakeCottageService() *FakeCottageService {
	return &FakeCottageService{}
}

func (f *FakeCottageService) GetAll(ctx context.Context) ([]domain.Cottage, error) {
	f.GetAllCalls++
	if f.GetAllFunc != nil {
		return f.GetAllFunc(ctx)
	}
	return nil, nil
}

func (f *FakeCottageService) GetByName(ctx context.Context, cottageName string) (domain.Cottage, error) {
	f.GetByNameCalls++
	if f.GetByNameFunc != nil {
		return f.GetByNameFunc(ctx, cottageName)
	}
	return domain.Cottage{Name: cottageName}, nil
}

func (f *FakeCottageService) GetByView(ctx context.Context, view string) ([]domain.Cottage, error) {
	f.GetByViewCalls++
	if f.GetByViewFunc != nil {
		return f.GetByViewFunc(ctx, view)
	}
	return nil, nil
}

func (f *FakeCottageService) AddBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	f.AddBookingCalls++
	if f.AddBookingFunc != nil {
		return f.AddBookingFunc(ctx, name, bookingId)
	}
	return nil
}

func (f *FakeCottageService) RemoveBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	f.RemoveBookingCalls++
	if f.RemoveBookingFunc != nil {
		return f.RemoveBookingFunc(ctx, name, bookingId)
	}
	return nil
}
