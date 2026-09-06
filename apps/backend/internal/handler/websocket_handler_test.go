package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cifo-monitoring/backend/internal/config"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/cifo-monitoring/backend/internal/ws"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketHandlerMissingToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	hub := ws.NewHub(logger)
	cfg := &config.Config{Environment: "development"}
	authService := service.NewAuthService(cfg, nil, nil, nil, nil, logger)
	handler := NewWebSocketHandler(hub, authService, logger)

	err := handler.HandleWebSocket(c)
	assert.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestWebSocketHandlerInvalidToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws?token=invalid-garbage-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	hub := ws.NewHub(logger)
	cfg := &config.Config{Environment: "production"}
	authService := service.NewAuthService(cfg, nil, nil, nil, nil, logger)
	handler := NewWebSocketHandler(hub, authService, logger)

	err := handler.HandleWebSocket(c)
	assert.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}
