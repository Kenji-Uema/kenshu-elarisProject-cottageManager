package mocks

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockBookingService struct {
	AddBookingFunc    func(ctx context.Context, booking domain.Booking) (string, error)
	RemoveBookingFunc func(ctx context.Context, cottageName string, bookingId bson.ObjectID) error

	AddBookingCalls    int
	RemoveBookingCalls int
}

func NewMockBookingService() *MockBookingService {
	return &MockBookingService{}
}

func (m *MockBookingService) AddBooking(ctx context.Context, booking domain.Booking) (string, error) {
	m.AddBookingCalls++
	if m.AddBookingFunc != nil {
		return m.AddBookingFunc(ctx, booking)
	}
	return "", nil
}

func (m *MockBookingService) RemoveBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error {
	m.RemoveBookingCalls++
	if m.RemoveBookingFunc != nil {
		return m.RemoveBookingFunc(ctx, cottageName, bookingId)
	}
	return nil
}
