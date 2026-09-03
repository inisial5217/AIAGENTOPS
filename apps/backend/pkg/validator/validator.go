package validator

import (
	"fmt"

	"github.com/cifo-monitoring/backend/pkg/apperror"
	"github.com/go-playground/validator/v10"
)

// CustomValidator validator instance wrapper
type CustomValidator struct {
	val *validator.Validate
}

// New creates validator instance
func New() *CustomValidator {
	return &CustomValidator{
		val: validator.New(),
	}
}

// Validate validates struct fields
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.val.Struct(i); err != nil {
		return apperror.Wrap(err, "VALIDATION_FAILED", 400, fmt.Sprintf("validation failed: %v", err), "Invalid input parameters")
	}
	return nil
}
