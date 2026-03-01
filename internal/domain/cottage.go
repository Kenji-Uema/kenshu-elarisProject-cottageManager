package domain

import (
	"errors"

	"github.com/Kenji-Uema/cottageManager/internal/app/validation"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Cottage struct {
	Id                bson.ObjectID
	Name              string
	View              string
	Details           CottageDetails
	Photos            []string
	PricePerNight     float32
	Bookings          []bson.ObjectID
	CurrentlyOccupied bool
}

type CottageDetails struct {
	Description          string
	FurnitureDescription string
	BathroomDescription  string
	AmenitiesDescription string
}

func NewCottageFromDoc(cottageDoc document.Cottage) (Cottage, error) {
	cottage := Cottage{
		Id:   cottageDoc.Id,
		Name: cottageDoc.Name,
		View: cottageDoc.View,
		Details: CottageDetails{
			Description:          cottageDoc.Details.Description,
			FurnitureDescription: cottageDoc.Details.FurnitureDescription,
			BathroomDescription:  cottageDoc.Details.BathroomDescription,
			AmenitiesDescription: cottageDoc.Details.AmenitiesDescription,
		},
		Photos:            cottageDoc.Photos,
		PricePerNight:     cottageDoc.PricePerNight,
		Bookings:          cottageDoc.Bookings,
		CurrentlyOccupied: cottageDoc.CurrentlyOccupied,
	}

	if err := validation.New().
		NotNilObjectID("Id", cottage.Id).
		NotBlank("Name", cottage.Name).
		NotBlank("view", cottage.View).
		NotZeroValue("photos", cottage.Photos).
		PositiveValue("pricePerNight", cottage.PricePerNight).
		NotZeroValue("bookings", cottage.Bookings).
		Validate(); err != nil {

		detailsErr := validation.New().
			NotBlank("details.description", cottage.Details.Description).
			NotBlank("details.furnitureDescription", cottage.Details.FurnitureDescription).
			NotBlank("details.bathroomDescription", cottage.Details.BathroomDescription).
			NotBlank("details.amenitiesDescription", cottage.Details.AmenitiesDescription).
			Validate()

		err = errors.Join(detailsErr, err)
		return Cottage{}, err
	}

	return cottage, nil
}

func (c Cottage) ToDto() dto.Cottage {
	bookings := make([]string, len(c.Bookings))
	for i, id := range c.Bookings {
		bookings[i] = id.Hex()
	}

	return dto.Cottage{
		Name: c.Name,
		Type: c.View,
		Details: dto.CottageDetails{
			Description:          c.Details.Description,
			View:                 c.View,
			FurnitureDescription: c.Details.FurnitureDescription,
			BathroomDescription:  c.Details.BathroomDescription,
			AmenitiesDescription: c.Details.AmenitiesDescription,
		},
		Photos:        c.Photos,
		PricePerNight: c.PricePerNight,
		Bookings:      bookings,
	}
}
