package app

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var availabilityServiceTracer = otel.Tracer("cottage-manager.app.availability-service")

type AvailabilityService interface {
	GetAvailablePeriods(ctx context.Context, name string, period domain.Period) (domain.CottageAvailablePeriod, error)
	GetAvailablePeriodsByCottageType(ctx context.Context, cottageType string, period domain.Period) ([]domain.CottageAvailablePeriod, error)
	IsCottageAvailable(ctx context.Context, cottageName string, period domain.Period) (bool, error)
}

type availabilityService struct {
	cottageService CottageService
	bookingService BookingService
}

func NewAvailabilityService(cs CottageService, bs BookingService) AvailabilityService {
	return &availabilityService{cottageService: cs, bookingService: bs}
}

func (s *availabilityService) GetAvailablePeriods(ctx context.Context, name string, period domain.Period) (domain.CottageAvailablePeriod, error) {
	ctx, span := availabilityServiceTracer.Start(ctx, "AvailabilityService.GetAvailablePeriods")
	defer span.End()
	span.SetAttributes(
		attribute.String("cottage.name", name),
		attribute.String("availability.checkin_date", period.CheckIn.UTC().Format(time.RFC3339)),
		attribute.String("availability.checkout_date", period.CheckOut.UTC().Format(time.RFC3339)),
	)

	slog.DebugContext(ctx, "calculating availability for cottage", "cottage", name, "check_in", period.CheckIn, "check_out", period.CheckOut)

	cottage, err := s.cottageService.GetByName(ctx, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_lookup_failed")
		slog.DebugContext(ctx, "failed to load cottage for availability", "cottage", name, "error", err)
		return domain.CottageAvailablePeriod{}, err
	}
	slog.DebugContext(ctx, "loaded cottage for availability", "cottage", cottage.Name, "bookings_count", len(cottage.Bookings))

	bookings, err := s.bookingService.GetBookings(ctx, cottage.Bookings)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "booking_lookup_failed")
		slog.DebugContext(ctx, "failed to load bookings for availability", "cottage", name, "error", err)
		return domain.CottageAvailablePeriod{}, err
	}
	slog.DebugContext(ctx, "loaded bookings for availability", "cottage", name, "bookings_count", len(bookings))

	periods := cottageVacancies(ctx, bookings, period)
	span.SetAttributes(attribute.Int("availability.periods.count", len(periods)))
	slog.DebugContext(ctx, "calculated availability for cottage", "cottage", name, "available_periods_count", len(periods))

	return domain.CottageAvailablePeriod{Name: cottage.Name, Periods: periods}, nil
}

func (s *availabilityService) GetAvailablePeriodsByCottageType(ctx context.Context, cottageType string, period domain.Period) ([]domain.CottageAvailablePeriod, error) {
	ctx, span := availabilityServiceTracer.Start(ctx, "AvailabilityService.GetAvailablePeriodsByCottageType")
	defer span.End()
	span.SetAttributes(
		attribute.String("cottage.type", cottageType),
		attribute.String("availability.checkin_date", period.CheckIn.UTC().Format(time.RFC3339)),
		attribute.String("availability.checkout_date", period.CheckOut.UTC().Format(time.RFC3339)),
	)

	slog.DebugContext(ctx, "calculating availability for cottage type", "cottage_type", cottageType, "check_in", period.CheckIn, "check_out", period.CheckOut)

	cottages, err := s.cottageService.GetByView(ctx, cottageType)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_lookup_failed")
		slog.DebugContext(ctx, "failed to load cottages by type", "cottage_type", cottageType, "error", err)
		return nil, err
	}
	slog.DebugContext(ctx, "loaded cottages by type", "cottage_type", cottageType, "cottages_count", len(cottages))

	cottageAvailablePeriods := make([]domain.CottageAvailablePeriod, len(cottages))
	for i, cottage := range cottages {
		slog.DebugContext(ctx, "calculating availability for cottage in type query", "index", i, "cottage", cottage.Name)
		availablePeriods, err := s.GetAvailablePeriods(ctx, cottage.Name, period)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "availability_lookup_failed")
			slog.DebugContext(ctx, "failed to calculate availability for cottage in type query", "index", i, "cottage", cottage.Name, "error", err)
			return nil, err
		}

		cottageAvailablePeriods[i] = availablePeriods
		slog.DebugContext(ctx, "calculated availability for cottage in type query", "index", i, "cottage", cottage.Name, "available_periods_count", len(availablePeriods.Periods))
	}

	slog.DebugContext(ctx, "calculated availability for cottage type", "cottage_type", cottageType, "cottages_count", len(cottageAvailablePeriods))
	span.SetAttributes(attribute.Int("cottage.results.count", len(cottageAvailablePeriods)))
	return cottageAvailablePeriods, nil
}

func (s *availabilityService) IsCottageAvailable(ctx context.Context, cottageName string, period domain.Period) (bool, error) {
	ctx, span := availabilityServiceTracer.Start(ctx, "AvailabilityService.IsCottageAvailable")
	defer span.End()
	span.SetAttributes(
		attribute.String("cottage.name", cottageName),
		attribute.String("availability.checkin_date", period.CheckIn.UTC().Format(time.RFC3339)),
		attribute.String("availability.checkout_date", period.CheckOut.UTC().Format(time.RFC3339)),
	)

	slog.DebugContext(ctx, "checking cottage availability", "cottage", cottageName, "check_in", period.CheckIn, "check_out", period.CheckOut)

	cottage, err := s.cottageService.GetByName(ctx, cottageName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_lookup_failed")
		slog.DebugContext(ctx, "failed to load cottage while checking availability", "cottage", cottageName, "error", err)
		return false, err
	}
	slog.DebugContext(ctx, "loaded cottage while checking availability", "cottage", cottageName, "bookings_count", len(cottage.Bookings))

	bookings, err := s.bookingService.GetBookings(ctx, cottage.Bookings)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "booking_lookup_failed")
		slog.DebugContext(ctx, "failed to load bookings while checking availability", "cottage", cottageName, "error", err)
		return false, err
	}
	slog.DebugContext(ctx, "loaded bookings while checking availability", "cottage", cottageName, "bookings_count", len(bookings))

	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].StayPeriod.CheckIn.Before(bookings[j].StayPeriod.CheckIn)
	})

	for _, b := range bookings {
		if !b.StayPeriod.CheckIn.Before(period.CheckOut) {
			slog.DebugContext(ctx, "no overlap possible anymore; booking starts after requested period", "cottage", cottageName, "booking_check_in", b.StayPeriod.CheckIn, "check_out", period.CheckOut)
			break
		}
		if period.CheckIn.Before(b.StayPeriod.CheckOut) && period.CheckOut.After(b.StayPeriod.CheckIn) {
			span.SetAttributes(attribute.Bool("cottage.available", false))
			slog.DebugContext(ctx, "cottage is not available due to overlap", "cottage", cottageName, "booking_check_in", b.StayPeriod.CheckIn, "booking_check_out", b.StayPeriod.CheckOut, "check_in", period.CheckIn, "check_out", period.CheckOut)
			return false, nil
		}
	}

	span.SetAttributes(attribute.Bool("cottage.available", true))
	slog.DebugContext(ctx, "cottage is available for requested period", "cottage", cottageName, "check_in", period.CheckIn, "check_out", period.CheckOut)
	return true, nil
}

func cottageVacancies(ctx context.Context, bookings []domain.Booking, period domain.Period) []domain.Period {
	slog.DebugContext(ctx, "starting cottage vacancies calculation",
		"bookings_count", len(bookings), "check_in", period.CheckIn, "check_out", period.CheckOut)

	period.Normalize()
	for i := range bookings {
		bookings[i].StayPeriod.Normalize()
	}
	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].StayPeriod.CheckIn.Before(bookings[j].StayPeriod.CheckIn)
	})

	slog.DebugContext(ctx, "search period",
		"period", period)
	slog.DebugContext(ctx, "cottage vacancies sorted",
		"period", period, "bookings", bookings)

	availablePeriods := make([]domain.Period, 0, len(bookings)+1)
	nextFreeStart := period.CheckIn

	for _, b := range bookings {
		stay := b.StayPeriod

		if !stay.CheckIn.Before(period.CheckOut) {
			slog.DebugContext(ctx, "stopping bookings scan; stay starts after period end",
				"stay_check_in", stay.CheckIn, "check_out", period.CheckOut)
			break
		}
		if !stay.CheckOut.After(period.CheckIn) {
			slog.DebugContext(ctx, "skipping stay; stay ends before search period starts",
				"stay_check_out", stay.CheckOut, "check_in", period.CheckIn)
			continue
		}

		gapEnd := stay.CheckIn.Add(-time.Nanosecond)
		if nextFreeStart.Before(gapEnd) {
			availablePeriods = append(availablePeriods, domain.Period{CheckIn: nextFreeStart, CheckOut: gapEnd})
			slog.DebugContext(ctx, "found gap",
				"gap_check_in", nextFreeStart, "gap_check_out", gapEnd)
		}

		if stay.CheckOut.After(nextFreeStart) {
			nextFreeStart = stay.CheckOut.Add(time.Nanosecond)
			if nextFreeStart.After(period.CheckOut) {
				nextFreeStart = period.CheckOut
			}
			slog.DebugContext(ctx, "advanced next free start",
				"next_free_start", nextFreeStart)
		}
	}

	if nextFreeStart.Before(period.CheckOut) {
		availablePeriods = append(availablePeriods, domain.Period{CheckIn: nextFreeStart, CheckOut: period.CheckOut})
		slog.DebugContext(ctx, "added trailing gap",
			"gap_check_in", nextFreeStart, "gap_check_out", period.CheckOut)
	}

	slog.DebugContext(ctx, "finished cottage vacancies calculation",
		"available_periods_count", len(availablePeriods), "available_periods", availablePeriods)
	return availablePeriods
}
