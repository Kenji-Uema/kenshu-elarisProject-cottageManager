package app

import (
	"context"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/port"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingService interface {
	AddBooking(ctx context.Context, booking domain.Booking) (string, error)
	RemoveBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error
}

type bookingService struct {
	bookingRepo         port.BookingRepo
	availabilityService AvailabilityService
	cottageService      CottageService
}

func NewBookingService(availabilityService AvailabilityService, cottageService CottageService, bookingRepo port.BookingRepo) BookingService {
	return &bookingService{
		availabilityService: availabilityService,
		cottageService:      cottageService,
		bookingRepo:         bookingRepo,
	}
}

func (s *bookingService) AddBooking(ctx context.Context, booking domain.Booking) (string, error) {
	slog.Debug("adding booking", "cottage", booking.CottageName, "period_start", booking.StayPeriod.Start, "period_end", booking.StayPeriod.End)

	isCottageFree, err := s.availabilityService.IsCottageAvailable(ctx, booking.CottageName, booking.StayPeriod)

	if err != nil {
		slog.Error("failed to validate cottage availability", "error", err, "cottage", booking.CottageName)
		return "", err
	}

	if !isCottageFree {
		slog.Warn("cottage not available for requested period", "cottage", booking.CottageName, "period_start", booking.StayPeriod.Start, "period_end", booking.StayPeriod.End)
		return "", &appErrors.CottageNotAvailableError{CottageName: booking.CottageName, Period: booking.StayPeriod}
	}

	bookingId, err := s.bookingRepo.AddBooking(ctx, booking)
	if err != nil {
		slog.Error("failed to persist booking", "error", err, "cottage", booking.CottageName)
		return "", err
	}

	if err := s.cottageService.AddBooking(ctx, booking.CottageName, bookingId); err != nil {
		slog.Error("failed to attach booking to cottage", "error", err, "cottage", booking.CottageName, "booking_id", bookingId.Hex())
		return "", err
	}

	slog.Info("booking created", "cottage", booking.CottageName, "booking_id", bookingId.Hex())
	return bookingId.Hex(), nil
}

func (s *bookingService) RemoveBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error {
	slog.Debug("removing booking", "cottage", cottageName, "booking_id", bookingId.Hex())

	_, err := s.bookingRepo.DeleteBooking(ctx, bookingId)
	if err != nil {
		slog.Error("failed to delete booking", "error", err, "cottage", cottageName, "booking_id", bookingId.Hex())
		return err
	}

	if err := s.cottageService.RemoveBooking(ctx, cottageName, bookingId); err != nil {
		slog.Error("failed to detach booking from cottage", "error", err, "cottage", cottageName, "booking_id", bookingId.Hex())
		return err
	}

	slog.Info("booking removed", "cottage", cottageName, "booking_id", bookingId.Hex())
	return nil
}
