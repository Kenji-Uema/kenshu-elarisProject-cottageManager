package dbErrors

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UnexpectedError struct {
	Err error
}

func (e *UnexpectedError) Error() string {
	return fmt.Sprintf("unexpected error: %v", e.Err)
}

type ValidationError struct {
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid request: %v", e.Reason)
}

type MissingBookingsError struct {
	Missing []bson.ObjectID
}

func (e *MissingBookingsError) Error() string {
	return fmt.Sprintf("not all booking ids found, missing: %v", e.Missing)
}
