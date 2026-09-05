package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name         string
		userRole     string
		requiredRole string
		expectedCode int
	}{
		{"admin accessing admin", "admin", "admin", http.StatusOK},
		{"admin accessing devops", "admin", "devops", http.StatusOK},
		{"admin accessing viewer", "admin", "viewer", http.StatusOK},
		{"devops accessing admin", "devops", "admin", http.StatusForbidden},
		{"devops accessing devops", "devops", "devops", http.StatusOK},
		{"devops accessing viewer", "devops", "viewer", http.StatusOK},
		{"viewer accessing admin", "viewer", "admin", http.StatusForbidden},
		{"viewer accessing devops", "viewer", "devops", http.StatusForbidden},
		{"viewer accessing viewer", "viewer", "viewer", http.StatusOK},
		{"no role defaults to viewer accessing admin", "", "admin", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.userRole != "" {
				c.Set("user_role", tt.userRole)
			}

			handler := RequireRole(nil, tt.requiredRole)(func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			err := handler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
