package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validator is a global instance of the go-playground/validator
var validate = validator.New()

// ValidateStruct validates a struct using the go-playground/validator
// It returns an error if validation fails, wrapping the original error with a custom error type.
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}
