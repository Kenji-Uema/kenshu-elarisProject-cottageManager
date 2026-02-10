package mocks

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockBookingRepo struct {
	GetBookingsFunc   func(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error)
	AddBookingFunc    func(ctx context.Context, booking domain.Booking) (bson.ObjectID, error)
	DeleteBookingFunc func(ctx context.Context, id bson.ObjectID) (bool, error)

	GetBookingsCalls   int
	AddBookingCalls    int
	DeleteBookingCalls int
}

func NewMockBookingRepo() *MockBookingRepo {
	return &MockBookingRepo{}
}

func (m *MockBookingRepo) GetBookings(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error) {
	m.GetBookingsCalls++
	if m.GetBookingsFunc != nil {
		return m.GetBookingsFunc(ctx, ids)
	}
	return nil, nil
}

func (m *MockBookingRepo) AddBooking(ctx context.Context, booking domain.Booking) (bson.ObjectID, error) {
	m.AddBookingCalls++
	if m.AddBookingFunc != nil {
		return m.AddBookingFunc(ctx, booking)
	}
	return bson.NilObjectID, nil
}

func (m *MockBookingRepo) DeleteBooking(ctx context.Context, id bson.ObjectID) (bool, error) {
	m.DeleteBookingCalls++
	if m.DeleteBookingFunc != nil {
		return m.DeleteBookingFunc(ctx, id)
	}
	return false, nil
}

type MockCottageRepo struct {
	GetAllFunc        func(ctx context.Context) ([]domain.Cottage, error)
	GetByNameFunc     func(ctx context.Context, name string) (domain.Cottage, error)
	GetByTypeFunc     func(ctx context.Context, cottageType string) ([]domain.Cottage, error)
	GetBookingsIdFunc func(ctx context.Context, name string) ([]bson.ObjectID, error)
	AddBookingFunc    func(ctx context.Context, name string, bookingId bson.ObjectID) error
	DeleteBookingFunc func(ctx context.Context, name string, bookingId bson.ObjectID) error

	GetAllCalls        int
	GetByNameCalls     int
	GetByTypeCalls     int
	GetBookingsIdCalls int
	AddBookingCalls    int
	DeleteBookingCalls int
}

func NewMockCottageRepo() *MockCottageRepo {
	return &MockCottageRepo{}
}

func (m *MockCottageRepo) GetAll(ctx context.Context) ([]domain.Cottage, error) {
	m.GetAllCalls++
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockCottageRepo) GetByName(ctx context.Context, name string) (domain.Cottage, error) {
	m.GetByNameCalls++
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, name)
	}
	return domain.Cottage{}, nil
}

func (m *MockCottageRepo) GetByType(ctx context.Context, cottageType string) ([]domain.Cottage, error) {
	m.GetByTypeCalls++
	if m.GetByTypeFunc != nil {
		return m.GetByTypeFunc(ctx, cottageType)
	}
	return nil, nil
}

func (m *MockCottageRepo) GetBookingsId(ctx context.Context, name string) ([]bson.ObjectID, error) {
	m.GetBookingsIdCalls++
	if m.GetBookingsIdFunc != nil {
		return m.GetBookingsIdFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockCottageRepo) AddBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	m.AddBookingCalls++
	if m.AddBookingFunc != nil {
		return m.AddBookingFunc(ctx, name, bookingId)
	}
	return nil
}

func (m *MockCottageRepo) DeleteBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	m.DeleteBookingCalls++
	if m.DeleteBookingFunc != nil {
		return m.DeleteBookingFunc(ctx, name, bookingId)
	}
	return nil
}
