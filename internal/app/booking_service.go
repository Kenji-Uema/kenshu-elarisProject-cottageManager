package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/dbErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"
	"github.com/Kenji-Uema/cottageManager/internal/infra/telemetry"
	"github.com/Kenji-Uema/cottageManager/internal/port"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const createInvoiceRoutingKey = "booking.create-invoice"
const (
	invoiceCurrency = "USD"
	invoiceTaxBps   = 500
)

var bookingServiceTracer = otel.Tracer("cottage-manager.app.booking-service")

type BookingService interface {
	GetBookings(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error)
	AddBooking(ctx context.Context, booking domain.Booking) (bson.ObjectID, error)
	RemoveBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error
}

type bookingService struct {
	bookingRepo    port.BookingRepo
	txManager      port.TransactionManager
	cottageService CottageService
	mqProducer     port.MqProducer
}

func NewBookingService(cottageService CottageService, bookingRepo port.BookingRepo, txManager port.TransactionManager, mqProducer port.MqProducer) BookingService {
	return &bookingService{
		cottageService: cottageService,
		bookingRepo:    bookingRepo,
		txManager:      txManager,
		mqProducer:     mqProducer,
	}
}

func (s *bookingService) GetBookings(ctx context.Context, ids []bson.ObjectID) ([]domain.Booking, error) {
	ctx, span := bookingServiceTracer.Start(ctx, "BookingService.GetBookings")
	defer span.End()
	span.SetAttributes(attribute.Int("booking.ids.count", len(ids)))

	slog.DebugContext(ctx, "retrieving bookings", "booking_ids", ids)

	bookingsDoc, err := s.bookingRepo.GetBookings(ctx, ids)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "booking_repo.get_bookings_failed")
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
		span.RecordError(mappingErr)
		span.SetStatus(codes.Error, "booking_mapping_failed")
		return nil, mappingErr
	}

	span.SetAttributes(attribute.Int("booking.results.count", len(bookings)))
	return bookings, nil
}

func (s *bookingService) AddBooking(ctx context.Context, booking domain.Booking) (bson.ObjectID, error) {
	ctx, span := bookingServiceTracer.Start(ctx, "BookingService.AddBooking")
	defer span.End()
	cottageName := booking.CottageName
	stayPeriod := booking.StayPeriod
	span.SetAttributes(
		attribute.String("booking.cottage_name", cottageName),
		attribute.String("booking.main_guest_id", booking.MainGuest.Hex()),
		attribute.String("booking.checkin_date", stayPeriod.CheckIn.UTC().Format(time.RFC3339)),
		attribute.String("booking.checkout_date", stayPeriod.CheckOut.UTC().Format(time.RFC3339)),
		attribute.Int("booking.number_of_guests", booking.NumberOfGuests),
	)
	slog.DebugContext(ctx, "adding booking", "cottage", cottageName, "check_in", stayPeriod.CheckIn, "check_out", stayPeriod.CheckOut)

	transactionRes, err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		overlapCtx, overlapSpan := telemetry.StartAppSpan(txCtx, "BookingService.AddBooking.CheckOverlap")
		hasOverlap, txErr := s.bookingRepo.HasOverlappingBooking(overlapCtx, cottageName, stayPeriod)
		if txErr != nil {
			telemetry.RecordSpanError(overlapSpan, txErr, "booking_repo.has_overlapping_booking_failed")
		}
		overlapSpan.End()
		if txErr != nil {
			slog.ErrorContext(txCtx, "failed to validate booking overlap", "error", txErr, "cottage", cottageName, "check_in", stayPeriod.CheckIn, "check_out", stayPeriod.CheckOut)
			return nil, txErr
		}
		if hasOverlap {
			slog.WarnContext(txCtx, "booking rejected due to overlap", "cottage", cottageName, "check_in", stayPeriod.CheckIn, "check_out", stayPeriod.CheckOut)
			return nil, &appErrors.CottageNotAvailableError{CottageName: cottageName, Period: stayPeriod}
		}

		saveCtx, saveSpan := telemetry.StartAppSpan(txCtx, "BookingService.AddBooking.PersistBooking")
		bookingId, txErr := s.bookingRepo.AddBooking(saveCtx, booking.ToDocument())
		if txErr != nil {
			telemetry.RecordSpanError(saveSpan, txErr, "booking_repo.add_booking_failed")
		}
		saveSpan.End()
		if txErr != nil {
			slog.ErrorContext(txCtx, "failed to persist booking", "error", txErr, "cottage", cottageName)
			return nil, txErr
		}

		attachCtx, attachSpan := telemetry.StartAppSpan(txCtx, "BookingService.AddBooking.AttachToCottage")
		txErr = s.cottageService.AddBooking(attachCtx, cottageName, bookingId)
		if txErr != nil {
			telemetry.RecordSpanError(attachSpan, txErr, "cottage_service.add_booking_failed")
		}
		attachSpan.End()
		if txErr != nil {
			slog.ErrorContext(txCtx, "failed to attach booking to cottage", "error", txErr, "cottage", cottageName, "booking_id", bookingId.Hex())
			return nil, txErr
		}

		if s.mqProducer != nil {
			loadCtx, loadSpan := telemetry.StartAppSpan(txCtx, "BookingService.AddBooking.LoadCottageForInvoice")
			cottage, txErr := s.cottageService.GetByName(loadCtx, cottageName)
			if txErr != nil {
				telemetry.RecordSpanError(loadSpan, txErr, "cottage_service.get_by_name_failed")
			}
			loadSpan.End()
			if txErr != nil {
				slog.ErrorContext(txCtx, "failed to load cottage for invoice payload", "error", txErr, "cottage", cottageName, "booking_id", bookingId.Hex())
				return nil, txErr
			}

			_, msgSpan := telemetry.StartAppSpan(txCtx, "BookingService.AddBooking.CreateInvoiceMessage")
			createInvoiceReq := createInvoiceMessage(booking, cottage, bookingId, time.Now().UTC())
			msgSpan.End()

			publishCtx, publishSpan := telemetry.StartAppSpan(txCtx, "BookingService.AddBooking.PublishInvoiceRequest")
			txErr = s.mqProducer.Publish(publishCtx, createInvoiceReq, createInvoiceRoutingKey)
			if txErr != nil {
				telemetry.RecordSpanError(publishSpan, txErr, "mq.publish_create_invoice_failed")
			}
			publishSpan.End()
			if txErr != nil {
				slog.ErrorContext(txCtx, "failed to publish create-invoice request", "error", txErr, "booking_id", bookingId.Hex())
				return nil, txErr
			}
		}

		return bookingId, nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "add_booking_failed")
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			return bson.NilObjectID, err
		}

		var cottageNotFoundErr *appErrors.CottageNotFound
		if errors.As(err, &cottageNotFoundErr) {
			return bson.NilObjectID, err
		}
		var cottageNotAvailableErr *appErrors.CottageNotAvailableError
		if errors.As(err, &cottageNotAvailableErr) {
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
		err := appErrors.UnexpectedError{Err: fmt.Errorf("transaction returned %T, want bson.ObjectID", transactionRes)}
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid_transaction_result")
		return bson.NilObjectID, err
	}

	span.SetAttributes(attribute.String("booking.id", bookingId.Hex()))
	slog.InfoContext(ctx, "booking created", "cottage", cottageName, "booking_id", bookingId.Hex())
	return bookingId, nil
}

func (s *bookingService) RemoveBooking(ctx context.Context, cottageName string, bookingId bson.ObjectID) error {
	ctx, span := bookingServiceTracer.Start(ctx, "BookingService.RemoveBooking")
	defer span.End()
	span.SetAttributes(
		attribute.String("booking.cottage_name", cottageName),
		attribute.String("booking.id", bookingId.Hex()),
	)

	slog.DebugContext(ctx, "removing booking", "cottage", cottageName, "booking_id", bookingId.Hex())

	_, err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		deleteCtx, deleteSpan := telemetry.StartAppSpan(txCtx, "BookingService.RemoveBooking.DeleteBooking")
		txErr := s.bookingRepo.DeleteBooking(deleteCtx, bookingId)
		if txErr != nil {
			telemetry.RecordSpanError(deleteSpan, txErr, "booking_repo.delete_booking_failed")
		}
		deleteSpan.End()
		if txErr != nil {
			slog.ErrorContext(txCtx, "failed to delete booking", "error", txErr, "cottage", cottageName, "booking_id", bookingId.Hex())
			return nil, txErr
		}

		detachCtx, detachSpan := telemetry.StartAppSpan(txCtx, "BookingService.RemoveBooking.DetachFromCottage")
		txErr = s.cottageService.RemoveBooking(detachCtx, cottageName, bookingId)
		if txErr != nil {
			telemetry.RecordSpanError(detachSpan, txErr, "cottage_service.remove_booking_failed")
		}
		detachSpan.End()
		if txErr != nil {
			slog.ErrorContext(txCtx, "failed to detach booking from cottage", "error", txErr, "cottage", cottageName, "booking_id", bookingId.Hex())
			return nil, txErr
		}

		return nil, nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "remove_booking_failed")
		var bookingNotFoundErr *dbErrors.BookingNotFoundErr
		if errors.As(err, &bookingNotFoundErr) {
			return &appErrors.BookingNotFound{BookingId: bookingId}
		}

		return &appErrors.UnexpectedError{Err: err}
	}

	slog.InfoContext(ctx, "booking removed", "cottage", cottageName, "booking_id", bookingId.Hex())
	return nil
}

func createInvoiceMessage(booking domain.Booking, cottage domain.Cottage, bookingID bson.ObjectID, issuedAt time.Time) *dto.CreateInvoicePaymentRequest {
	nights := int32(booking.StayPeriod.CheckOut.Sub(booking.StayPeriod.CheckIn) / (24 * time.Hour))
	if nights < 1 {
		nights = 1
	}

	valuePerNight := priceToMinorUnits(cottage.PricePerNight)
	total := valuePerNight * int64(nights)
	taxTotal := calculateTax(total)

	return &dto.CreateInvoicePaymentRequest{
		IdempotencyKey: bookingID.Hex(),
		BookingId:      bookingID.Hex(),
		PayerId:        booking.MainGuest.Hex(),
		IssuedAt:       timestamppb.New(issuedAt),
		DueAt:          timestamppb.New(issuedAt.Add(48 * time.Hour)),
		Payer: &dto.Payer{
			Name:           booking.Payer.Name,
			Email:          booking.Payer.Email,
			DocumentNumber: booking.Payer.DocumentNumber,
			BillingAddress: booking.Payer.BillingAddress,
		},
		Booking: &dto.BookingSnapshot{
			CottageName:    booking.CottageName,
			Nights:         nights,
			NumberOfGuests: int32(booking.NumberOfGuests),
			ValuePerNight: &dto.Money{
				Amount:   valuePerNight,
				Currency: invoiceCurrency,
			},
		},
		Total: &dto.Money{
			Amount:   total,
			Currency: invoiceCurrency,
		},
		TaxTotal: &dto.Money{
			Amount:   taxTotal,
			Currency: invoiceCurrency,
		},
		DiscountTotal: &dto.Money{
			Amount:   0,
			Currency: invoiceCurrency,
		},
	}
}

func priceToMinorUnits(amount float32) int64 {
	return int64(math.Round(float64(amount) * 100))
}

func calculateTax(total int64) int64 {
	return int64(math.Round(float64(total*invoiceTaxBps) / 10000))
}
