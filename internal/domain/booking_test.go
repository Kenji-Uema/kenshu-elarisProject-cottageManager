package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewBookingFromDocument_RejectsInvalidStatus(t *testing.T) {
	_, err := NewBookingFromDocument(document.Booking{
		Id:             bson.NewObjectID(),
		MainGuest:      bson.NewObjectID(),
		NumberOfGuests: 2,
		StayPeriod: document.Period{
			CheckIn:  time.Now().UTC(),
			CheckOut: time.Now().UTC().Add(24 * time.Hour),
		},
		CottageName: "Lake House",
		Status:      "INVALID",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var validationErr *validationErrors.ErrValidationConstrain
	if ok := errors.As(err, &validationErr); !ok {
		t.Fatalf("expected validation error, got %T", err)
	}
}
