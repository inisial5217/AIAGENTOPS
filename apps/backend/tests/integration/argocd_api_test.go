package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArgoCDAPI_Endpoints(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := getBaseURL()

	t.Run("GetOverview_Authenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/argocd/overview", baseURL), nil)
		req.Header.Set("Authorization", "Bearer dev-token-viewer")

		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ListApplications_Authenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/argocd/applications", baseURL), nil)
		req.Header.Set("Authorization", "Bearer dev-token-devops")

		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Unauthorized_Access", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/argocd/applications", baseURL), nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
