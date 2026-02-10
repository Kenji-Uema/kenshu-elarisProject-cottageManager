package mocks

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
)

type MockAvailabilityService struct {
	GetAvailablePeriodsFunc              func(ctx context.Context, name string, period domain.Period) ([]domain.Period, error)
	GetAvailablePeriodsByCottageTypeFunc func(ctx context.Context, cottageType string, period domain.Period) ([]domain.CottageAvailablePeriod, error)
	IsCottageAvailableFunc               func(ctx context.Context, cottageName string, period domain.Period) (bool, error)

	GetAvailablePeriodsCalls              int
	GetAvailablePeriodsByCottageTypeCalls int
	IsCottageAvailableCalls               int
}

func NewMockAvailabilityService() *MockAvailabilityService {
	return &MockAvailabilityService{}
}

func (m *MockAvailabilityService) GetAvailablePeriods(ctx context.Context, name string, period domain.Period) ([]domain.Period, error) {
	m.GetAvailablePeriodsCalls++
	if m.GetAvailablePeriodsFunc != nil {
		return m.GetAvailablePeriodsFunc(ctx, name, period)
	}
	return nil, nil
}

func (m *MockAvailabilityService) GetAvailablePeriodsByCottageType(ctx context.Context, cottageType string, period domain.Period) ([]domain.CottageAvailablePeriod, error) {
	m.GetAvailablePeriodsByCottageTypeCalls++
	if m.GetAvailablePeriodsByCottageTypeFunc != nil {
		return m.GetAvailablePeriodsByCottageTypeFunc(ctx, cottageType, period)
	}
	return nil, nil
}

func (m *MockAvailabilityService) IsCottageAvailable(ctx context.Context, cottageName string, period domain.Period) (bool, error) {
	m.IsCottageAvailableCalls++
	if m.IsCottageAvailableFunc != nil {
		return m.IsCottageAvailableFunc(ctx, cottageName, period)
	}
	return false, nil
}
