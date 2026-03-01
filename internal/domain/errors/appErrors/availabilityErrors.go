package appErrors

import (
	"fmt"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
)

type CottageNotAvailableUnexpectedError struct {
	Err error
}

func (e CottageNotAvailableUnexpectedError) Error() string {
	return fmt.Sprintf("Unexpected error happened when checking if cottage is available: %v", e.Err)
}

type InvalidPeriodError struct {
	Period domain.Period
}

func (e InvalidPeriodError) Error() string {
	return fmt.Sprintf("Invalid period: check-out must be after check-in: %v - %v", e.Period.CheckIn, e.Period.CheckOut)
}

type CottageNotAvailableError struct {
	CottageName string
	Period      domain.Period
}

func (e CottageNotAvailableError) Error() string {
	return fmt.Sprintf("Cottage %v is not available for period %v", e.CottageName, e.Period)
}
