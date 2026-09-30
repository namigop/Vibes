package todo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator with a typed helper that returns
// a flat slice of field errors suitable for the HTTP error envelope.
type Validator struct {
	v *validator.Validate
}

// NewValidator returns a Validator using the default tag set.
func NewValidator() *Validator {
	return &Validator{v: validator.New(validator.WithRequiredStructEnabled())}
}

// Validate returns a slice of FieldError for the first validation failure,
// one entry per invalid field. The slice is empty when the input is valid.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate runs the validator on the given struct and converts the error
// into a friendly FieldError list. Returns nil when the struct is valid.
func (val *Validator) Validate(s any) []FieldError {
	err := val.v.Struct(s)
	if err == nil {
		return nil
	}
	var verr validator.ValidationErrors
	if !errors.As(err, &verr) {
		return []FieldError{{Field: "_", Message: err.Error()}}
	}
	out := make([]FieldError, 0, len(verr))
	for _, fe := range verr {
		out = append(out, FieldError{
			Field:   strings.ToLower(fe.Field()),
			Message: friendlyMessage(fe),
		})
	}
	return out
}

func friendlyMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	default:
		return "is invalid"
	}
}
