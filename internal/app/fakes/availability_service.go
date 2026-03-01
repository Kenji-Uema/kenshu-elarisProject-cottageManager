package fakes

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
)

type FakeAvailabilityService struct {
	GetAvailablePeriodsFunc              func(ctx context.Context, name string, period domain.Period) (domain.CottageAvailablePeriod, error)
	GetAvailablePeriodsByCottageTypeFunc func(ctx context.Context, cottageType string, period domain.Period) ([]domain.CottageAvailablePeriod, error)
	IsCottageAvailableFunc               func(ctx context.Context, cottageName string, period domain.Period) (bool, error)

	GetAvailablePeriodsCalls              int
	GetAvailablePeriodsByCottageTypeCalls int
	IsCottageAvailableCalls               int
}

func NewFakeAvailabilityService() *FakeAvailabilityService {
	return &FakeAvailabilityService{}
}

func (f *FakeAvailabilityService) GetAvailablePeriods(ctx context.Context, name string, period domain.Period) (domain.CottageAvailablePeriod, error) {
	f.GetAvailablePeriodsCalls++
	if f.GetAvailablePeriodsFunc != nil {
		return f.GetAvailablePeriodsFunc(ctx, name, period)
	}
	return domain.CottageAvailablePeriod{}, nil
}

func (f *FakeAvailabilityService) GetAvailablePeriodsByCottageType(ctx context.Context, cottageType string, period domain.Period) ([]domain.CottageAvailablePeriod, error) {
	f.GetAvailablePeriodsByCottageTypeCalls++
	if f.GetAvailablePeriodsByCottageTypeFunc != nil {
		return f.GetAvailablePeriodsByCottageTypeFunc(ctx, cottageType, period)
	}
	return nil, nil
}

func (f *FakeAvailabilityService) IsCottageAvailable(ctx context.Context, cottageName string, period domain.Period) (bool, error) {
	f.IsCottageAvailableCalls++
	if f.IsCottageAvailableFunc != nil {
		return f.IsCottageAvailableFunc(ctx, cottageName, period)
	}
	return false, nil
}
