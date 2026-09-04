package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError custom error struct
type AppError struct {
	Code       string `json:"code"`
	HTTPStatus int    `json:"-"`
	Message    string `json:"-"`
	UserMsg    string `json:"message"`
	Err        error  `json:"-"`
}

// Error returns error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates app error
func New(code string, status int, msg string, userMsg string) *AppError {
	return &AppError{
		Code:       code,
		HTTPStatus: status,
		Message:    msg,
		UserMsg:    userMsg,
	}
}

// Wrap wraps existing error
func Wrap(err error, code string, status int, msg string, userMsg string) *AppError {
	return &AppError{
		Code:       code,
		HTTPStatus: status,
		Message:    msg,
		UserMsg:    userMsg,
		Err:        err,
	}
}

// Predefined app errors
var (
	ErrUnauthorized = New("AUTH_UNAUTHORIZED", http.StatusUnauthorized, "unauthorized access", "Authentication required")
	ErrForbidden    = New("AUTH_FORBIDDEN", http.StatusForbidden, "forbidden action", "Insufficient permissions")
	ErrNotFound     = New("NOT_FOUND", http.StatusNotFound, "resource not found", "Requested resource not found")
	ErrBadRequest   = New("BAD_REQUEST", http.StatusBadRequest, "invalid request", "Invalid input parameters")
	ErrInternal     = New("INTERNAL_ERROR", http.StatusInternalServerError, "internal server error", "An unexpected error occurred")
	ErrRateLimit    = New("RATE_LIMITED", http.StatusTooManyRequests, "rate limit exceeded", "Too many requests, try again later")
)

// FromError converts error
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return Wrap(err, "INTERNAL_ERROR", http.StatusInternalServerError, err.Error(), "An unexpected error occurred")
}
