package bindings

type CottageNameURI struct {
	Name string `uri:"name" binding:"required"`
}

type CottageTypeURI struct {
	CottageType string `uri:"cottageType" binding:"required"`
}

type CottageNameAndBookingIdURI struct {
	CottageNameURI
	BookingId string `uri:"bookingId" binding:"required"`
}
