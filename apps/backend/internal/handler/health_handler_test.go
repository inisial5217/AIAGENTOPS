package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLiveness(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHealthHandler(nil, nil, "")
	err := h.Liveness(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res["status"])
}

func TestReadinessUnhealthyWhenNil(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHealthHandler(nil, nil, "")
	err := h.Readiness(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}
