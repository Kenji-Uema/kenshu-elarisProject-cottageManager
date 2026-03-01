package dto

import (
	"time"
)

type PeriodDto struct {
	CheckIn  time.Time `json:"check_in"`
	CheckOut time.Time `json:"check_out"`
}

type AvailablePeriodDTO struct {
	Name    string      `json:"cottage_name"`
	Periods []PeriodDto `json:"available_periods"`
}
