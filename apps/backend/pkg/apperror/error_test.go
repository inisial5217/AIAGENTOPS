package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	err := New("TEST_ERR", http.StatusBadRequest, "internal msg", "user friendly msg")
	assert.Equal(t, "internal msg", err.Error())
	assert.Equal(t, "user friendly msg", err.UserMsg)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)

	wrapped := Wrap(errors.New("root cause"), "WRAPPED", http.StatusInternalServerError, "wrapper", "user msg")
	assert.Contains(t, wrapped.Error(), "root cause")
	assert.Equal(t, "root cause", wrapped.Unwrap().Error())
}

func TestFromError(t *testing.T) {
	stdErr := errors.New("something broke")
	converted := FromError(stdErr)
	assert.Equal(t, "INTERNAL_ERROR", converted.Code)
	assert.Equal(t, http.StatusInternalServerError, converted.HTTPStatus)

	existing := ErrUnauthorized
	assert.Equal(t, ErrUnauthorized, FromError(existing))
	assert.Nil(t, FromError(nil))
}
