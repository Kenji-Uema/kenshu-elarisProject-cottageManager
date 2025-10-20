package booking

import (
	"cottageManager/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestDto struct {
	GuestId        string    `json:"mainGuest"`
	NumberOfGuests int       `json:"numberOfGuests"`
	CheckInDate    time.Time `json:"checkInDate" time_format:"2006-01-02"`
	CheckOutDate   time.Time `json:"checkOutDate" time_format:"2006-01-02"`
}

type ConfirmationDto struct {
	BookingId string `json:"bookingId"`
}

func (bookingDto *RequestDto) ToDomain(cottageName string) (domain.Booking, error) {
	mainGuestId, err := primitive.ObjectIDFromHex(bookingDto.GuestId)
	if err != nil {
		return domain.Booking{}, err
	}

	return domain.Booking{
		MainGuest:      mainGuestId,
		NumberOfGuests: bookingDto.NumberOfGuests,
		StayPeriod: domain.Period{
			Start: bookingDto.CheckInDate,
			End:   bookingDto.CheckOutDate,
		},
		CottageName: cottageName,
	}, nil
}
