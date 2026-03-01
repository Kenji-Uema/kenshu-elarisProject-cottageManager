package validation

import (
	"reflect"

	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (v *Validator) NotNilObjectID(field string, id bson.ObjectID) *Validator {
	v.steps = append(v.steps, func() error {
		if id == bson.NilObjectID {
			return &validationErrors.ErrValidationConstrain{Field: field, Message: "must not be nil"}
		}
		return nil
	})
	return v
}

func (v *Validator) NoDuplicates(field string, list any) *Validator {
	v.steps = append(v.steps, func() error {
		rv := reflect.ValueOf(list)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return &validationErrors.ErrValidationConstrain{
				Field:   field,
				Message: "must be a slice/array",
			}
		}

		for i := 0; i < rv.Len(); i++ {
			for j := i + 1; j < rv.Len(); j++ {
				if reflect.DeepEqual(rv.Index(i).Interface(), rv.Index(j).Interface()) {
					return &validationErrors.ErrValidationConstrain{
						Field:   field,
						Message: "must not contain duplicates",
					}
				}
			}
		}
		return nil
	})

	return v
}
