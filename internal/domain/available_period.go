package domain

import (
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
)

type Period struct {
	CheckIn  time.Time
	CheckOut time.Time
}

type CottageAvailablePeriod struct {
	Name    string
	Periods []Period
}

func (p *Period) Normalize() {
	startOfDay := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}

	p.CheckIn = startOfDay(p.CheckIn)
	p.CheckOut = startOfDay(p.CheckOut.AddDate(0, 0, 1)).Add(-time.Nanosecond)
}

func (c *CottageAvailablePeriod) ToDto() dto.AvailablePeriodDTO {
	periods := make([]dto.PeriodDto, len(c.Periods))
	for i, p := range c.Periods {
		periods[i] = dto.PeriodDto{
			CheckIn:  p.CheckIn,
			CheckOut: p.CheckOut,
		}
	}

	return dto.AvailablePeriodDTO{
		Name:    c.Name,
		Periods: periods,
	}
}
