package appErrors

import "fmt"

type CottageNotFound struct {
	Err error
}

func (e CottageNotFound) Error() string {
	return fmt.Sprintf("Cottage not found: %v", e.Err)
}

type AddBookingToCottageError struct {
	Err error
}

func (e AddBookingToCottageError) Error() string {
	return fmt.Sprintf("Could not add booking: %v", e.Err)
}

type RemoveBookingFromCottageError struct {
	Err error
}

func (e RemoveBookingFromCottageError) Error() string {
	return fmt.Sprintf("Could not remove booking: %v", e.Err)
}

type UnexpectedError struct {
	Err error
}

func (e UnexpectedError) Error() string {
	return fmt.Sprintf("Unexpected error: %v", e.Err)
}
