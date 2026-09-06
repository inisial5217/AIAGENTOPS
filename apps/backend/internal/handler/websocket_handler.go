package handler

import (
	"log/slog"
	"net/http"

	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/cifo-monitoring/backend/internal/ws"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// WebSocketHandler manages ws connections
type WebSocketHandler struct {
	hub         *ws.Hub
	authService *service.AuthService
	logger      *slog.Logger
	upgrader    websocket.Upgrader
}

// NewWebSocketHandler creates handler instance
func NewWebSocketHandler(hub *ws.Hub, authService *service.AuthService, logger *slog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authService: authService,
		logger:      logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// HandleWebSocket upgrades and handles ws
func (h *WebSocketHandler) HandleWebSocket(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing authentication token")
	}

	claims, err := h.authService.ValidateToken(c.Request().Context(), token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid authentication token")
	}

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Warn("ws upgrade failed", slog.String("error", err.Error()))
		return err
	}

	userID := claims.Subject
	username := claims.PreferredUsername
	if username == "" {
		username = claims.Name
	}

	client := ws.NewClient(h.hub, conn, userID, username, h.logger)
	h.hub.Register(client)

	// start pumps
	go client.WritePump()
	client.ReadPump()

	return nil
}
