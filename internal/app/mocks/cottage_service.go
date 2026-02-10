package mocks

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockCottageService struct {
	GetAllFunc        func(ctx context.Context) ([]domain.Cottage, error)
	GetByNameFunc     func(ctx context.Context, cottageName string) (domain.Cottage, error)
	AddBookingFunc    func(ctx context.Context, name string, bookingId bson.ObjectID) error
	RemoveBookingFunc func(ctx context.Context, name string, bookingId bson.ObjectID) error

	GetAllCalls        int
	GetByNameCalls     int
	AddBookingCalls    int
	RemoveBookingCalls int
}

func NewMockCottageService() *MockCottageService {
	return &MockCottageService{}
}

func (m *MockCottageService) GetAll(ctx context.Context) ([]domain.Cottage, error) {
	m.GetAllCalls++
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockCottageService) GetByName(ctx context.Context, cottageName string) (domain.Cottage, error) {
	m.GetByNameCalls++
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, cottageName)
	}
	return domain.Cottage{}, nil
}

func (m *MockCottageService) AddBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	m.AddBookingCalls++
	if m.AddBookingFunc != nil {
		return m.AddBookingFunc(ctx, name, bookingId)
	}
	return nil
}

func (m *MockCottageService) RemoveBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	m.RemoveBookingCalls++
	if m.RemoveBookingFunc != nil {
		return m.RemoveBookingFunc(ctx, name, bookingId)
	}
	return nil
}
