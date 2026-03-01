package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"
	"github.com/Kenji-Uema/cottageManager/internal/port"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingService interface {
	GetBookings(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error)
	AddBooking(ctx context.Context, booking domain.Booking) (bson.ObjectID, error)
	RemoveBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error
}

type bookingService struct {
	bookingRepo    port.BookingRepo
	txManager      port.TransactionManager
	cottageService CottageService
}

func NewBookingService(cottageService CottageService, bookingRepo port.BookingRepo, txManager port.TransactionManager) BookingService {
	return &bookingService{
		cottageService: cottageService,
		bookingRepo:    bookingRepo,
		txManager:      txManager,
	}
}

func (s *bookingService) GetBookings(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error) {
	slog.DebugContext(ctx, "retrieving bookings", "booking_ids", ids)

	bookingsDoc, err := s.bookingRepo.GetBookings(ctx, ids)
	if err != nil {
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return nil, err
		}

		var missingBookingsErr *dbErrors.MissingBookingsErr
		if errors.As(err, &missingBookingsErr) {
			return nil, &appErrors.CorruptedDataError{Err: err}
		}

		var corruptedDataErr *dbErrors.CorruptedDataErr
		if errors.As(err, &corruptedDataErr) {
			return nil, &appErrors.CorruptedDataError{Err: err}
		}

		return nil, &appErrors.UnexpectedError{Err: err}
	}

	bookings := make([]domain.Booking, len(bookingsDoc))
	var mappingErr error
	for i, bookingDoc := range bookingsDoc {
		bookings[i], err = domain.NewBookingFromDocument(bookingDoc)
		mappingErr = errors.Join(mappingErr, err)
	}

	if mappingErr != nil {
		return nil, mappingErr
	}

	return bookings, nil
}

func (s *bookingService) AddBooking(ctx context.Context, booking domain.Booking) (bson.ObjectID, error) {
	cottageName := booking.CottageName
	stayPeriod := booking.StayPeriod
	slog.DebugContext(ctx, "adding booking", "cottage", cottageName, "check_in", stayPeriod.CheckIn, "check_out", stayPeriod.CheckOut)

	transactionRes, err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		bookingId, txErr := s.bookingRepo.AddBooking(txCtx, booking.ToDocument())
		if txErr != nil {
			slog.ErrorContext(txCtx, "failed to persist booking", "error", txErr, "cottage", cottageName)
			return nil, txErr
		}

		if txErr = s.cottageService.AddBooking(txCtx, cottageName, bookingId); txErr != nil {
			slog.ErrorContext(txCtx, "failed to attach booking to cottage", "error", txErr, "cottage", cottageName, "booking_id", bookingId.Hex())
			return nil, txErr
		}

		return bookingId, nil
	})

	if err != nil {
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return bson.NilObjectID, err
		}

		var cottageNotFoundErr *appErrors.CottageNotFound
		if errors.As(err, &cottageNotFoundErr) {
			return bson.NilObjectID, err
		}

		var addBookingToCottageErr *appErrors.AddBookingToCottageError
		if errors.As(err, &addBookingToCottageErr) {
			return bson.NilObjectID, err
		}

		return bson.NilObjectID, &appErrors.UnexpectedError{Err: err}
	}

	bookingId, ok := transactionRes.(bson.ObjectID)
	if !ok {
		return bson.NilObjectID, appErrors.UnexpectedError{Err: fmt.Errorf("transaction returned %T, want bson.ObjectID", transactionRes)}
	}

	slog.InfoContext(ctx, "booking created", "cottage", cottageName, "booking_id", bookingId.Hex())
	return bookingId, nil
}

func (s *bookingService) RemoveBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error {
	slog.DebugContext(ctx, "removing booking", "cottage", cottageName, "booking_id", bookingId.Hex())

	_, err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		txErr := s.bookingRepo.DeleteBooking(txCtx, bookingId)
		if txErr != nil {
			slog.ErrorContext(txCtx, "failed to delete booking", "error", txErr, "cottage", cottageName, "booking_id", bookingId.Hex())
			return nil, txErr
		}

		if txErr = s.cottageService.RemoveBooking(txCtx, cottageName, bookingId); txErr != nil {
			slog.ErrorContext(txCtx, "failed to detach booking from cottage", "error", txErr, "cottage", cottageName, "booking_id", bookingId.Hex())
			return nil, txErr
		}

		return nil, nil
	})

	if err != nil {
		var bookingNotFoundErr *dbErrors.BookingNotFoundErr
		if errors.As(err, &bookingNotFoundErr) {
			return &appErrors.BookingNotFound{BookingId: bookingId}
		}

		return &appErrors.UnexpectedError{Err: err}
	}

	slog.InfoContext(ctx, "booking removed", "cottage", cottageName, "booking_id", bookingId.Hex())
	return nil
}
