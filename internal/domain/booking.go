package domain

import (
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/app/validation"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Booking struct {
	Id             bson.ObjectID
	MainGuest      bson.ObjectID
	NumberOfGuests int
	StayPeriod     Period
	CottageName    string
	Status         string
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
		Status:         "PENDING",
	}

	if err := validation.New().
		NotNilObjectID("mainGuest", booking.MainGuest).
		PositiveValue("numberOfGuests", booking.NumberOfGuests).
		ValidPeriod(booking.StayPeriod.CheckIn, booking.StayPeriod.CheckOut).
		NotBlank("cottageName", booking.CottageName).
		NotBlank("status", booking.Status).Validate(); err != nil {

		return Booking{}, err
	}

	return booking, nil
}

func NewBookingFromDocument(doc document.Booking) (Booking, error) {
	booking := Booking{
		Id:             doc.Id,
		MainGuest:      doc.MainGuest,
		NumberOfGuests: doc.NumberOfGuests,
		StayPeriod: Period{
			CheckIn:  doc.StayPeriod.CheckIn,
			CheckOut: doc.StayPeriod.CheckOut,
		},
		CottageName: doc.CottageName,
		Status:      doc.Status,
	}

	if err := validation.New().
		NotNilObjectID("bookingId", booking.Id).
		NotNilObjectID("mainGuest", booking.MainGuest).
		PositiveValue("numberOfGuests", booking.NumberOfGuests).
		ValidPeriod(booking.StayPeriod.CheckIn, booking.StayPeriod.CheckOut).
		NotBlank("cottageName", booking.CottageName).
		NotBlank("status", booking.Status).Validate(); err != nil {

		return Booking{}, err
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
		Status:      b.Status,
	}
}
