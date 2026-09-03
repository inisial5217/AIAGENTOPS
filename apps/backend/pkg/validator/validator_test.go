package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type sampleRequest struct {
	Email string `validate:"required,email"`
	Role  string `validate:"required,oneof=admin devops viewer"`
}

func TestValidator_ValidateSuccess(t *testing.T) {
	// test valid struct
	v := New()
	req := sampleRequest{
		Email: "admin@cifo.id",
		Role:  "admin",
	}
	err := v.Validate(req)
	assert.NoError(t, err)
}

func TestValidator_ValidateFailure(t *testing.T) {
	// test invalid struct
	v := New()
	req := sampleRequest{
		Email: "not-an-email",
		Role:  "superadmin",
	}
	err := v.Validate(req)
	assert.Error(t, err)
}
