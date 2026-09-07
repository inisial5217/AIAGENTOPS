package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	e := echo.New()
	e.Use(SecurityHeaders())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	err := SecurityHeaders()(h)(c)
	assert.NoError(t, err)

	headers := rec.Header()
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	assert.Equal(t, "default-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'self';", headers.Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=31536000; includeSubDomains; preload", headers.Get("Strict-Transport-Security"))
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
	assert.Equal(t, "geolocation=(), camera=(), microphone=(), payment=()", headers.Get("Permissions-Policy"))
}
