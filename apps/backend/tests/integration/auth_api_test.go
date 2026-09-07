package integration_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getBaseURL() string {
	base := os.Getenv("API_BASE_URL")
	if base == "" {
		base = "http://127.0.0.1:8080"
	}
	return base
}

func TestAuthAPI_Endpoints(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := getBaseURL()

	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("%s/healthz", baseURL))
		if err != nil {
			t.Skipf("backend server not reachable at %s: %v", baseURL, err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetMe_Unauthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/auth/me", baseURL), nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("GetMe_DevTokenAdmin", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/auth/me", baseURL), nil)
		req.Header.Set("Authorization", "Bearer dev-token-admin")

		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var bodyMap map[string]map[string]interface{}
		err = json.Unmarshal(body, &bodyMap)
		assert.NoError(t, err)
		user := bodyMap["user"]
		assert.Equal(t, "admin@cifo.local", user["email"])
		assert.Equal(t, "admin", user["role"])
	})

	t.Run("GetMe_DevTokenDevops", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/auth/me", baseURL), nil)
		req.Header.Set("Authorization", "Bearer dev-token-devops")

		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var bodyMap map[string]map[string]interface{}
		err = json.Unmarshal(body, &bodyMap)
		assert.NoError(t, err)
		user := bodyMap["user"]
		assert.Equal(t, "devops@cifo.local", user["email"])
		assert.Equal(t, "devops", user["role"])
	})

	t.Run("GetMe_DevTokenViewer", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/auth/me", baseURL), nil)
		req.Header.Set("Authorization", "Bearer dev-token-viewer")

		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var bodyMap map[string]map[string]interface{}
		err = json.Unmarshal(body, &bodyMap)
		assert.NoError(t, err)
		user := bodyMap["user"]
		assert.Equal(t, "viewer@cifo.local", user["email"])
		assert.Equal(t, "viewer", user["role"])
	})
}
