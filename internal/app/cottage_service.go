package app

import (
	"context"
	"cottageManager/internal/domain"
	"cottageManager/internal/domain/errors/appErrors"
	"cottageManager/internal/port"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CottageService interface {
	GetAll(ctx context.Context) ([]domain.Cottage, error)
	GetByName(ctx context.Context, cottageName string) (domain.Cottage, error)
	AddBooking(ctx context.Context, name string, bookingId primitive.ObjectID) error
	RemoveBooking(ctx context.Context, name string, bookingId primitive.ObjectID) error
}

type cottageService struct {
	cottageRepo port.CottageRepo
}

func NewCottageService(repo port.CottageRepo) CottageService {
	return &cottageService{cottageRepo: repo}
}

func (s *cottageService) GetAll(ctx context.Context) ([]domain.Cottage, error) {
	slog.Debug("retrieving all cottages")
	cottages, err := s.cottageRepo.GetAll(ctx)

	if err != nil {
		slog.Error("failed to retrieve cottages from repository", "error", err)
		return nil, &appErrors.UnexpectedError{Err: err}
	}

	return cottages, nil
}

func (s *cottageService) GetByName(ctx context.Context, name string) (domain.Cottage, error) {
	slog.Debug("retrieving cottage by name", "cottage", name)
	cottage, err := s.cottageRepo.GetByName(ctx, name)

	if err != nil {
		return domain.Cottage{}, &appErrors.CottageNotFound{Err: err}
	}

	return cottage, nil
}

func (s *cottageService) AddBooking(ctx context.Context, name string, bookingId primitive.ObjectID) error {
	slog.Debug("associating booking to cottage", "cottage", name, "booking_id", bookingId.Hex())
	if err := s.cottageRepo.AddBooking(ctx, name, bookingId); err != nil {
		slog.Error("failed to associate booking to cottage", "error", err, "cottage", name, "booking_id", bookingId.Hex())
		return &appErrors.AddBookingToCottageError{Err: err}
	}

	return nil
}

func (s *cottageService) RemoveBooking(ctx context.Context, name string, bookingId primitive.ObjectID) error {
	slog.Debug("detaching booking from cottage", "cottage", name, "booking_id", bookingId.Hex())
	if err := s.cottageRepo.DeleteBooking(ctx, name, bookingId); err != nil {
		slog.Error("failed to detach booking from cottage", "error", err, "cottage", name, "booking_id", bookingId.Hex())
		return &appErrors.RemoveBookingFromCottageError{Err: err}
	}

	return nil
}
