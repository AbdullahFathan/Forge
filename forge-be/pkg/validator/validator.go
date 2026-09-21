package validator

import (
	"github.com/go-playground/validator/v10"

	"workspace/pkg/apperr"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

func Struct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var fields []FieldError
	if ves, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ves {
			fields = append(fields, FieldError{
				Field:   fe.Field(),
				Tag:     fe.Tag(),
				Message: fe.Error(),
			})
		}
	}
	return apperr.ErrValidation.WithDetails(fields)
}
