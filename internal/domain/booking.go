package domain

import (
	"strings"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/app/validation"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	Id             bson.ObjectID
	MainGuest      bson.ObjectID
	NumberOfGuests int
	StayPeriod     Period
	CottageName    string
	Status         BookingStatus
	Payer          BookingPayer
}

type BookingPayer struct {
	Name           string
	Email          string
	DocumentNumber string
	BillingAddress string
}

func NewBookingFromDto(dto dto.BookingRequestDto, cottageName string) (Booking, error) {
	mainGuest, err := bson.ObjectIDFromHex(dto.GuestId)
	if err != nil {
		return Booking{}, err
	}

	checkInDate, err := time.Parse("2006-01-02", dto.CheckInDate)
	if err != nil {
		return Booking{}, err
	}

	checkOutDate, err := time.Parse("2006-01-02", dto.CheckOutDate)
	if err != nil {
		return Booking{}, err
	}

	booking := Booking{
		MainGuest:      mainGuest,
		NumberOfGuests: dto.NumberOfGuests,
		StayPeriod:     Period{CheckIn: checkInDate, CheckOut: checkOutDate},
		CottageName:    cottageName,
		Status:         BookingStatusPending,
		Payer: BookingPayer{
			Name:           dto.GuestName,
			Email:          dto.GuestEmail,
			DocumentNumber: dto.GuestDocument,
			BillingAddress: dto.BillingAddress,
		},
	}

	if err := validation.New().
		NotNilObjectID("mainGuest", booking.MainGuest).
		PositiveValue("numberOfGuests", booking.NumberOfGuests).
		ValidPeriod(booking.StayPeriod.CheckIn, booking.StayPeriod.CheckOut).
		NotBlank("cottageName", booking.CottageName).
		NotBlank("status", string(booking.Status)).Validate(); err != nil {
		return Booking{}, err
	}

	if !booking.Status.IsValid() {
		return Booking{}, &validationErrors.ErrValidationConstrain{
			Field:   "status",
			Message: "must be one of PENDING, CONFIRMED, CANCELLED",
		}
	}

	return booking, nil
}

func NewBookingFromDocument(doc document.Booking) (Booking, error) {
	status := ParseBookingStatus(doc.Status)

	booking := Booking{
		Id:             doc.Id,
		MainGuest:      doc.MainGuest,
		NumberOfGuests: doc.NumberOfGuests,
		StayPeriod: Period{
			CheckIn:  doc.StayPeriod.CheckIn,
			CheckOut: doc.StayPeriod.CheckOut,
		},
		CottageName: doc.CottageName,
		Status:      status,
	}

	if err := validation.New().
		NotNilObjectID("bookingId", booking.Id).
		NotNilObjectID("mainGuest", booking.MainGuest).
		PositiveValue("numberOfGuests", booking.NumberOfGuests).
		ValidPeriod(booking.StayPeriod.CheckIn, booking.StayPeriod.CheckOut).
		NotBlank("cottageName", booking.CottageName).
		NotBlank("status", string(booking.Status)).Validate(); err != nil {

		return Booking{}, err
	}

	if !booking.Status.IsValid() {
		return Booking{}, &validationErrors.ErrValidationConstrain{
			Field:   "status",
			Message: "must be one of PENDING, CONFIRMED, CANCELLED",
		}
	}

	return booking, nil
}

func (b Booking) ToDocument() document.Booking {
	return document.Booking{
		Id:             b.Id,
		MainGuest:      b.MainGuest,
		NumberOfGuests: b.NumberOfGuests,
		StayPeriod: document.Period{
			CheckIn:  b.StayPeriod.CheckIn,
			CheckOut: b.StayPeriod.CheckOut,
		},
		CottageName: b.CottageName,
		Status:      b.Status.StorageValue(),
	}
}

func (s BookingStatus) IsValid() bool {
	switch ParseBookingStatus(string(s)) {
	case BookingStatusPending, BookingStatusConfirmed, BookingStatusCancelled:
		return true
	default:
		return false
	}
}

func (s BookingStatus) StorageValue() string {
	return strings.ToLower(string(ParseBookingStatus(string(s))))
}

func ParseBookingStatus(status string) BookingStatus {
	return BookingStatus(strings.ToUpper(status))
}
