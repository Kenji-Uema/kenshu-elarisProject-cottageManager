package cottage

import (
	"cottageManager/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Dto struct {
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Details       DetailsDto `json:"details"`
	Photos        []string   `json:"photos"`
	PricePerNight float32    `json:"price_per_night"`
	Bookings      []string   `json:"bookings"`
}

type DetailsDto struct {
	Description          string `json:"description"`
	View                 string `json:"view"`
	FurnitureDescription string `json:"furniture_description"`
	BathroomDescription  string `json:"bathroom_description"`
	AmenitiesDescription string `json:"amenities_description"`
}

func FromCottageDomainToDto(cottage domain.Cottage) Dto {
	return Dto{
		Name:          cottage.Name,
		Type:          cottage.View,
		Details:       fromCottageDetailsDomainToDto(cottage.Details),
		Photos:        cottage.Photos,
		PricePerNight: cottage.PricePerNight,
		Bookings:      bookingsToHex(cottage.Bookings),
	}
}

func fromCottageDetailsDomainToDto(cottageDetails domain.CottageDetails) DetailsDto {
	return DetailsDto{
		Description:          cottageDetails.Description,
		View:                 cottageDetails.View,
		FurnitureDescription: cottageDetails.FurnitureDescription,
		BathroomDescription:  cottageDetails.BathroomDescription,
		AmenitiesDescription: cottageDetails.AmenitiesDescription,
	}
}

func bookingsToHex(ids []primitive.ObjectID) []string {
	result := make([]string, len(ids))

	for i, id := range ids {
		result[i] = id.Hex()
	}

	return result
}
