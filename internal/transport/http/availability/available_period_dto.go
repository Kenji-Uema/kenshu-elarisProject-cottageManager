package availability

import (
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
)

type AvailablePeriodDTO struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func fromDomain(period domain.Period) AvailablePeriodDTO {
	return AvailablePeriodDTO{
		From: period.Start,
		To:   period.End,
	}
}
