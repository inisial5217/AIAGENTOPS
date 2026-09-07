package integration_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDockerAPI_Endpoints(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := getBaseURL()

	t.Run("ListContainers_Authenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/docker/containers", baseURL), nil)
		req.Header.Set("Authorization", "Bearer dev-token-admin")

		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var respObj struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		err = json.Unmarshal(body, &respObj)
		assert.NoError(t, err)
		assert.NotEmpty(t, respObj.Data, "expected active docker containers in running environment")
	})

	t.Run("GetDockerSystemInfo", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/docker/system", baseURL), nil)
		req.Header.Set("Authorization", "Bearer dev-token-admin")

		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var sys map[string]interface{}
		err = json.Unmarshal(body, &sys)
		assert.NoError(t, err)
		assert.Contains(t, sys, "containers_total")
	})

	t.Run("ListImages", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/docker/images", baseURL), nil)
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
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/docker/containers", baseURL), nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("backend not reachable: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
