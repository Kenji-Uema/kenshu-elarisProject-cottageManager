package booking

import (
	"cottageManager/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestDto struct {
	GuestId        string `json:"mainGuest"`
	NumberOfGuests int    `json:"numberOfGuests"`
	CheckInDate    string `json:"checkInDate"`
	CheckOutDate   string `json:"checkOutDate"`
}

type ConfirmationDto struct {
	BookingId string `json:"bookingId"`
}

func (bookingDto *RequestDto) ToDomain(cottageName string) (domain.Booking, error) {
	mainGuestId, err := primitive.ObjectIDFromHex(bookingDto.GuestId)
	if err != nil {
		return domain.Booking{}, err
	}

	checkInDate, err := time.Parse("2006-01-02", bookingDto.CheckInDate)
	if err != nil {
		return domain.Booking{}, err
	}
	checkOutDate, err := time.Parse("2006-01-02", bookingDto.CheckOutDate)
	if err != nil {
		return domain.Booking{}, err
	}

	return domain.Booking{
		MainGuest:      mainGuestId,
		NumberOfGuests: bookingDto.NumberOfGuests,
		StayPeriod: domain.Period{
			Start: checkInDate,
			End:   checkOutDate,
		},
		CottageName: cottageName,
	}, nil
}
