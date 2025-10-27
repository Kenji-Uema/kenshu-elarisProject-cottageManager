package app

import (
	"context"
	"cottageManager/internal/domain"
	"cottageManager/internal/domain/errors/appErrors"
	"cottageManager/internal/port"
	"log/slog"
	"sort"
	"time"
)

type AvailabilityService interface {
	GetAvailablePeriods(ctx context.Context, name string, period domain.Period) ([]domain.Period, error)
	GetAvailablePeriodsByCottageType(ctx context.Context, cottageType string, period domain.Period) ([]domain.CottageAvailablePeriod, error)
	IsCottageAvailable(ctx context.Context, cottageName string, period domain.Period) (bool, error)
}

type availabilityService struct {
	cottageRepo port.CottageRepo
	bookingRepo port.BookingRepo
}

func NewAvailabilityService(cr port.CottageRepo, br port.BookingRepo) AvailabilityService {
	return &availabilityService{cottageRepo: cr, bookingRepo: br}
}

func (s *availabilityService) GetAvailablePeriods(ctx context.Context, name string, period domain.Period) ([]domain.Period, error) {
	slog.Debug("calculating availability for cottage", "cottage", name, "period_start", period.Start, "period_end", period.End)
	cottage, err := s.cottageRepo.GetByName(ctx, name)
	if err != nil {
		slog.Error("failed to load cottage for availability", "error", err, "cottage", name)
		return nil, err
	}

	bookings, err := s.bookingRepo.GetBookings(ctx, cottage.Bookings)
	if err != nil {
		slog.Error("failed to load bookings for availability", "error", err, "cottage", name)
		return nil, err
	}

	return cottageVacancies(bookings, period), nil
}

func (s *availabilityService) GetAvailablePeriodsByCottageType(ctx context.Context, cottageType string, period domain.Period) ([]domain.CottageAvailablePeriod, error) {
	slog.Debug("calculating availability for cottage type", "cottage_type", cottageType, "period_start", period.Start, "period_end", period.End)
	cottages, err := s.cottageRepo.GetByType(ctx, cottageType)

	if err != nil {
		slog.Error("failed to load cottages by type", "error", err, "cottage_type", cottageType)
		return nil, err
	}

	var cottageAvailablePeriods []domain.CottageAvailablePeriod
	for _, cottage := range cottages {
		bookings, err := s.bookingRepo.GetBookings(ctx, cottage.Bookings)
		if err != nil {
			slog.Error("failed to load bookings for cottage", "error", err, "cottage", cottage.Name)
			return nil, err
		}

		cottageAvailablePeriods = append(cottageAvailablePeriods, domain.CottageAvailablePeriod{
			Name:    cottage.Name,
			Periods: cottageVacancies(bookings, period),
		})
	}

	return cottageAvailablePeriods, nil
}

func (s *availabilityService) IsCottageAvailable(ctx context.Context, cottageName string, period domain.Period) (bool, error) {
	slog.Debug("checking cottage availability", "cottage", cottageName, "period_start", period.Start, "period_end", period.End)
	cottage, err := s.cottageRepo.GetByName(ctx, cottageName)

	if err != nil {
		slog.Error("failed to load cottage while checking availability", "error", err, "cottage", cottageName)
		return false, err
	}

	bookings, err := s.bookingRepo.GetBookings(ctx, cottage.Bookings)

	if err != nil {
		slog.Error("failed to load bookings while checking availability", "error", err, "cottage", cottageName)
		return false, &appErrors.CottageNotAvailableUnexpectedError{Err: err}
	}

	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].StayPeriod.Start.Before(bookings[j].StayPeriod.Start)
	})

	for _, b := range bookings {
		if !b.StayPeriod.Start.Before(period.End) {
			break
		}
		if period.Start.Before(b.StayPeriod.End) && period.End.After(b.StayPeriod.Start) {
			return false, nil
		}
	}
	return true, nil
}

func cottageVacancies(bookings []domain.Booking, period domain.Period) []domain.Period {
	period.Normalize()
	for i := range bookings {
		bookings[i].StayPeriod.Normalize()
	}

	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].StayPeriod.Start.Before(bookings[j].StayPeriod.Start)
	})

	available := make([]domain.Period, 0, len(bookings)+1)
	nextFreeStart := period.Start

	for _, b := range bookings {
		stay := b.StayPeriod

		if !stay.Start.Before(period.End) {
			break
		}
		if !stay.End.After(period.Start) {
			continue
		}

		gapEnd := stay.Start.Add(-time.Nanosecond)
		if nextFreeStart.Before(gapEnd) {
			available = append(available, domain.Period{Start: nextFreeStart, End: gapEnd})
		}

		if stay.End.After(nextFreeStart) {
			nextFreeStart = stay.End.Add(time.Nanosecond)
			if nextFreeStart.After(period.End) {
				nextFreeStart = period.End
			}
		}
	}

	if nextFreeStart.Before(period.End) {
		available = append(available, domain.Period{Start: nextFreeStart, End: period.End})
	}

	return available
}
