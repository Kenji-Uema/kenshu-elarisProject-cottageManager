package dto

type BookingRequestDto struct {
	GuestId        string `json:"mainGuest"`
	NumberOfGuests int    `json:"numberOfGuests"`
	CheckInDate    string `json:"checkInDate"`
	CheckOutDate   string `json:"checkOutDate"`
	GuestName      string `json:"guestName"`
	GuestEmail     string `json:"guestEmail"`
	GuestDocument  string `json:"guestDocument"`
	BillingAddress string `json:"billingAddress"`
}

type ConfirmationDto struct {
	Message   string              `json:"message"`
	BookingId string              `json:"bookingId"`
	Info      ConfirmationInfoDto `json:"info"`
}

type ConfirmationInfoDto struct {
	CottageName    string `json:"cottageName"`
	NumberOfGuests int    `json:"numberOfGuests"`
	CheckInDate    string `json:"checkInDate"`
	CheckOutDate   string `json:"checkOutDate"`
}
