package security

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultClient_GetSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-token", r.Header.Get("X-Vault-Token"))

		switch r.URL.Path {
		case "/v1/sys/health":
			w.WriteHeader(http.StatusOK)
		case "/v1/secret/data/cifo/backend":
			resp := vaultKVResponse{}
			resp.Data.Data = map[string]interface{}{
				"POSTGRES_PASSWORD": "vault_secret_password",
				"PORT":              "8080",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewVaultClient(server.URL, "test-token")

	ctx := context.Background()
	err := client.Health(ctx)
	require.NoError(t, err)

	data, err := client.GetSecret(ctx, "cifo/backend")
	require.NoError(t, err)
	assert.Equal(t, "vault_secret_password", data["POSTGRES_PASSWORD"])

	val, err := client.GetString(ctx, "cifo/backend", "POSTGRES_PASSWORD")
	require.NoError(t, err)
	assert.Equal(t, "vault_secret_password", val)

	_, err = client.GetString(ctx, "cifo/backend", "NON_EXISTENT")
	assert.Error(t, err)
}
