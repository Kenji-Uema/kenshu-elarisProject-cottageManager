package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"
	"github.com/Kenji-Uema/cottageManager/internal/port"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var cottageServiceTracer = otel.Tracer("cottage-manager.app.cottage-service")

type CottageService interface {
	GetAll(ctx context.Context) ([]domain.Cottage, error)
	GetByView(ctx context.Context, view string) ([]domain.Cottage, error)
	GetByName(ctx context.Context, cottageName string) (domain.Cottage, error)
	AddBooking(ctx context.Context, name string, bookingId bson.ObjectID) error
	RemoveBooking(ctx context.Context, name string, bookingId bson.ObjectID) error
}

type cottageService struct {
	cottageRepo port.CottageRepo
}

func NewCottageService(repo port.CottageRepo) CottageService {
	return &cottageService{cottageRepo: repo}
}

func (s *cottageService) GetAll(ctx context.Context) ([]domain.Cottage, error) {
	ctx, span := cottageServiceTracer.Start(ctx, "CottageService.GetAll")
	defer span.End()

	slog.DebugContext(ctx, "retrieving all cottagesDoc")

	cottagesDoc, err := s.cottageRepo.GetAll(ctx)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_repo.get_all_failed")
		slog.ErrorContext(ctx, "failed to retrieve cottagesDoc from repository", "error", err)
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return nil, err
		}

		return nil, &appErrors.UnexpectedError{Err: err}
	}

	cottages := make([]domain.Cottage, len(cottagesDoc))
	for i, c := range cottagesDoc {
		cottage, validationErr := domain.NewCottageFromDoc(c)

		err = errors.Join(validationErr, err)
		cottages[i] = cottage
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_mapping_failed")
		slog.ErrorContext(ctx, "failed to convert document.Cottage to domain.Cottage", "error", err)

		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return nil, &appErrors.CorruptedDataError{Err: err}
		}

		return nil, &appErrors.UnexpectedError{Err: err}
	}

	span.SetAttributes(attribute.Int("cottage.results.count", len(cottages)))
	return cottages, nil
}

func (s *cottageService) GetByView(ctx context.Context, view string) ([]domain.Cottage, error) {
	ctx, span := cottageServiceTracer.Start(ctx, "CottageService.GetByView")
	defer span.End()
	span.SetAttributes(attribute.String("cottage.view", view))

	slog.DebugContext(ctx, "retrieving cottages by view", "view", view)

	cottagesDoc, err := s.cottageRepo.GetByView(ctx, view)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_repo.get_by_view_failed")
		slog.ErrorContext(ctx, "failed to retrieve cottagesDoc by view from repository", "error", err, "view", view)
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return nil, err
		}

		return nil, &appErrors.UnexpectedError{Err: err}
	}

	cottages := make([]domain.Cottage, len(cottagesDoc))
	for i, c := range cottagesDoc {
		cottage, validationErr := domain.NewCottageFromDoc(c)

		err = errors.Join(validationErr, err)
		cottages[i] = cottage
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_mapping_failed")
		slog.ErrorContext(ctx, "failed to convert cottagesDoc by view to domain.Cottage", "error", err, "view", view)

		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return nil, &appErrors.CorruptedDataError{Err: err}
		}

		return nil, &appErrors.UnexpectedError{Err: err}
	}

	span.SetAttributes(attribute.Int("cottage.results.count", len(cottages)))
	return cottages, nil
}

func (s *cottageService) GetByName(ctx context.Context, name string) (domain.Cottage, error) {
	ctx, span := cottageServiceTracer.Start(ctx, "CottageService.GetByName")
	defer span.End()
	span.SetAttributes(attribute.String("cottage.name", name))

	slog.DebugContext(ctx, "retrieving cottage by name", "cottage", name)

	cottageDoc, err := s.cottageRepo.GetByName(ctx, name)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_repo.get_by_name_failed")
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return domain.Cottage{}, err
		}
		var cottageNotFoundErr *dbErrors.CottageNotFoundErr
		if errors.As(err, &cottageNotFoundErr) {
			return domain.Cottage{}, &appErrors.CottageNotFound{Err: err}
		}
		var corruptedDataErr *dbErrors.CorruptedDataErr
		if errors.As(err, &corruptedDataErr) {
			return domain.Cottage{}, &appErrors.CorruptedDataError{Err: err}
		}

		return domain.Cottage{}, &appErrors.UnexpectedError{Err: err}
	}

	cottage, err := domain.NewCottageFromDoc(cottageDoc)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_mapping_failed")
		slog.ErrorContext(ctx, "failed to convert cottageDoc to domain.Cottage", "error", err)

		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return domain.Cottage{}, &appErrors.CorruptedDataError{Err: err}
		}

		return domain.Cottage{}, &appErrors.UnexpectedError{Err: err}
	}

	return cottage, nil
}

func (s *cottageService) AddBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	ctx, span := cottageServiceTracer.Start(ctx, "CottageService.AddBooking")
	defer span.End()
	span.SetAttributes(
		attribute.String("cottage.name", name),
		attribute.String("booking.id", bookingId.Hex()),
	)

	slog.DebugContext(ctx, "associating booking to cottage", "cottage", name, "booking_id", bookingId.Hex())

	if err := s.cottageRepo.AddBooking(ctx, name, bookingId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_repo.add_booking_failed")
		slog.ErrorContext(ctx, "failed to associate booking to cottage", "error", err, "cottage", name, "booking_id", bookingId.Hex())
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return err
		}

		var cottageNotFoundErr *dbErrors.CottageNotFoundErr
		if errors.As(err, &cottageNotFoundErr) {
			return &appErrors.CottageNotFound{Err: err}
		}

		var bookingNotUpdatedErr *dbErrors.BookingsNotUpdatedErr
		if errors.As(err, &bookingNotUpdatedErr) {
			return &appErrors.AddBookingToCottageError{Err: err}
		}

		return &appErrors.UnexpectedError{Err: err}
	}

	return nil
}

func (s *cottageService) RemoveBooking(ctx context.Context, name string, bookingId bson.ObjectID) error {
	ctx, span := cottageServiceTracer.Start(ctx, "CottageService.RemoveBooking")
	defer span.End()
	span.SetAttributes(
		attribute.String("cottage.name", name),
		attribute.String("booking.id", bookingId.Hex()),
	)

	slog.DebugContext(ctx, "detaching booking from cottage", "cottage", name, "booking_id", bookingId.Hex())

	if err := s.cottageRepo.DeleteBooking(ctx, name, bookingId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cottage_repo.remove_booking_failed")
		slog.ErrorContext(ctx, "failed to detach booking from cottage", "error", err, "cottage", name, "booking_id", bookingId.Hex())
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return err
		}

		var cottageNotFoundErr *dbErrors.CottageNotFoundErr
		if errors.As(err, &cottageNotFoundErr) {
			return &appErrors.CottageNotFound{Err: err}
		}

		var bookingNotUpdatedErr *dbErrors.BookingsNotUpdatedErr
		if errors.As(err, &bookingNotUpdatedErr) {
			return &appErrors.RemoveBookingFromCottageError{Err: err}
		}

		return &appErrors.UnexpectedError{Err: err}
	}

	return nil
}
