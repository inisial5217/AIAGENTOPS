package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_New(t *testing.T) {
	// test new error
	err := New("TEST_ERR", http.StatusBadRequest, "internal msg", "user msg")
	assert.NotNil(t, err)
	assert.Equal(t, "TEST_ERR", err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, "internal msg", err.Error())
	assert.Equal(t, "user msg", err.UserMsg)
}

func TestAppError_Wrap(t *testing.T) {
	// test wrap error
	raw := errors.New("raw db error")
	wrapped := Wrap(raw, "DB_ERR", http.StatusInternalServerError, "db failed", "database error")
	assert.NotNil(t, wrapped)
	assert.Equal(t, raw, wrapped.Unwrap())
	assert.Contains(t, wrapped.Error(), "db failed: raw db error")
}

func TestFromError(t *testing.T) {
	// test from error
	assert.Nil(t, FromError(nil))

	raw := errors.New("something went wrong")
	converted := FromError(raw)
	assert.NotNil(t, converted)
	assert.Equal(t, "INTERNAL_ERROR", converted.Code)
	assert.Equal(t, http.StatusInternalServerError, converted.HTTPStatus)
}
