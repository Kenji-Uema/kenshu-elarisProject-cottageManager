package document

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Cottage struct {
	Id                bson.ObjectID   `bson:"_id,omitempty"`
	Name              string          `bson:"name"`
	View              string          `bson:"view"`
	Details           CottageDetails  `bson:"details"`
	Photos            []string        `bson:"photos"`
	PricePerNight     float32         `bson:"price_per_night"`
	Bookings          []bson.ObjectID `bson:"bookings"`
	CurrentlyOccupied bool            `bson:"currently_occupied"`
}

type CottageDetails struct {
	Description          string `bson:"description"`
	FurnitureDescription string `bson:"furniture_description"`
	BathroomDescription  string `bson:"bathroom_description"`
	AmenitiesDescription string `bson:"amenities_description"`
}
