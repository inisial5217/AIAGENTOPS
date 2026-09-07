package security

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cifo-monitoring/backend/pkg/apperror"
)

// VaultClient reads secrets
type VaultClient interface {
	GetSecret(ctx context.Context, path string) (map[string]interface{}, error)
	GetString(ctx context.Context, path string, key string) (string, error)
	Health(ctx context.Context) error
}

type vaultKVResponse struct {
	Data struct {
		Data map[string]interface{} `json:"data"`
	} `json:"data"`
}

type vaultClientImpl struct {
	addr       string
	token      string
	httpClient *http.Client
	mu         sync.RWMutex
	cache      map[string]map[string]interface{}
}

// NewVaultClient creates client
func NewVaultClient(addr, token string) VaultClient {
	if addr == "" {
		addr = "http://127.0.0.1:8200"
	}
	addr = strings.TrimRight(addr, "/")

	return &vaultClientImpl{
		addr:  addr,
		token: token,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache: make(map[string]map[string]interface{}),
	}
}

// Health checks vault
func (c *vaultClientImpl) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/sys/health", c.addr), nil)
	if err != nil {
		return err
	}

	if c.token != "" {
		req.Header.Set("X-Vault-Token", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apperror.NewInternal(fmt.Sprintf("vault unreachable: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 429 && resp.StatusCode != 473 {
		return apperror.NewInternal(fmt.Sprintf("vault unhealthy: %d", resp.StatusCode))
	}

	return nil
}

// GetSecret fetches secret
func (c *vaultClientImpl) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	c.mu.RLock()
	if data, exists := c.cache[path]; exists {
		c.mu.RUnlock()
		return data, nil
	}
	c.mu.RUnlock()

	cleanPath := strings.TrimPrefix(path, "/")
	if !strings.HasPrefix(cleanPath, "secret/data/") {
		if strings.HasPrefix(cleanPath, "secret/") {
			cleanPath = strings.Replace(cleanPath, "secret/", "secret/data/", 1)
		} else {
			cleanPath = "secret/data/" + cleanPath
		}
	}

	url := fmt.Sprintf("%s/v1/%s", c.addr, cleanPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if c.token != "" {
		req.Header.Set("X-Vault-Token", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("vault request failed: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperror.NewNotFound(fmt.Sprintf("vault secret %s", path))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperror.NewInternal(fmt.Sprintf("vault returned status: %d", resp.StatusCode))
	}

	var vaultResp vaultKVResponse
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, apperror.NewInternal(fmt.Sprintf("failed to decode vault response: %v", err))
	}

	c.mu.Lock()
	c.cache[path] = vaultResp.Data.Data
	c.mu.Unlock()

	return vaultResp.Data.Data, nil
}

// GetString reads key
func (c *vaultClientImpl) GetString(ctx context.Context, path string, key string) (string, error) {
	data, err := c.GetSecret(ctx, path)
	if err != nil {
		return "", err
	}

	val, ok := data[key]
	if !ok {
		return "", apperror.NewNotFound(fmt.Sprintf("vault key %s", key))
	}

	strVal, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val), nil
	}

	return strVal, nil
}
