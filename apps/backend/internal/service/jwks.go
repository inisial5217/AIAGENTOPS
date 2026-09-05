package service

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// JSONWebKey represents individual key in JWKS
type JSONWebKey struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	Alg string   `json:"alg"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// JSONWebKeySet represents JWKS response
type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}

// JWKSCache caches public keys
type JWKSCache struct {
	sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	ttl       time.Duration
	jwksURL   string
	client    *http.Client
}

// NewJWKSCache creates JWKS cache
func NewJWKSCache(jwksURL string, ttl time.Duration) *JWKSCache {
	if ttl <= 0 {
		ttl = 1 * time.Hour
	}
	return &JWKSCache{
		keys:    make(map[string]*rsa.PublicKey),
		ttl:     ttl,
		jwksURL: jwksURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// GetKey retrieves RSA public key
func (c *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	c.RLock()
	key, ok := c.keys[kid]
	valid := time.Now().Before(c.expiresAt)
	c.RUnlock()

	if ok && valid {
		return key, nil
	}

	// refresh keys from jwks
	if err := c.refresh(); err != nil {
		// if refresh fails but key exists, use it
		if ok {
			return key, nil
		}
		return nil, fmt.Errorf("refresh jwks: %w", err)
	}

	c.RLock()
	defer c.RUnlock()
	key, ok = c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("kid %q not found in JWKS", kid)
	}
	return key, nil
}

// refresh downloads and parses keys
func (c *JWKSCache) refresh() error {
	resp, err := c.client.Get(c.jwksURL)
	if err != nil {
		return fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS response status: %d", resp.StatusCode)
	}

	var jwks JSONWebKeySet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode JWKS: %w", err)
	}

	newKeys := make(map[string]*rsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pubKey, err := parseRSAPublicKey(k.N, k.E)
		if err == nil {
			newKeys[k.Kid] = pubKey
		}
	}

	c.Lock()
	c.keys = newKeys
	c.expiresAt = time.Now().Add(c.ttl)
	c.Unlock()

	return nil
}

// parseRSAPublicKey parses n and e into RSA key
func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}

	var eInt uint64
	for _, b := range eBytes {
		eInt = (eInt << 8) | uint64(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(eInt),
	}, nil
}
