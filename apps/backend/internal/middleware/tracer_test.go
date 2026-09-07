package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestTracerMiddlewareNewTrace verifies trace creation
func TestTracerMiddlewareNewTrace(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	h := func(ctx echo.Context) error {
		handlerCalled = true
		traceID := ctx.Get("trace_id")
		assert.NotEmpty(t, traceID)
		return ctx.String(http.StatusOK, "ok")
	}

	mw := TracerMiddleware("test-service")
	err := mw(h)(c)

	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.NotEmpty(t, rec.Header().Get(HeaderXTraceID))
	assert.NotEmpty(t, rec.Header().Get(HeaderTraceParent))
}

// TestTracerMiddlewarePropagation verifies w3c propagation
func TestTracerMiddlewarePropagation(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test/propagate", nil)
	existingTraceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	existingSpanID := "00f067aa0ba902b7"
	req.Header.Set(HeaderTraceParent, "00-"+existingTraceID+"-"+existingSpanID+"-01")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(ctx echo.Context) error {
		traceID := ctx.Get("trace_id")
		assert.Equal(t, existingTraceID, traceID)
		return ctx.String(http.StatusOK, "ok")
	}

	mw := TracerMiddleware("test-service")
	err := mw(h)(c)

	assert.NoError(t, err)
	assert.Equal(t, existingTraceID, rec.Header().Get(HeaderXTraceID))
}

// TestTracerMiddlewareError verifies error recording
func TestTracerMiddlewareError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test/error", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedErr := errors.New("handler failed")
	h := func(ctx echo.Context) error {
		return expectedErr
	}

	mw := TracerMiddleware("test-service")
	err := mw(h)(c)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NotEmpty(t, rec.Header().Get(HeaderXTraceID))
}
