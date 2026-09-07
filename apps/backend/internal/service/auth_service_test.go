package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/cifo-monitoring/backend/internal/config"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// mockUserRepo mocks UserRepository
type mockUserRepo struct {
	upserted *model.User
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByKeycloakID(ctx context.Context, keycloakID string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpsertKeycloakUser(ctx context.Context, user *model.User) (*model.User, error) {
	m.upserted = user
	user.ID = "11111111-1111-1111-1111-111111111111"
	user.IsActive = true
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return user, nil
}
func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) UpdateRole(ctx context.Context, id string, role string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) SetActive(ctx context.Context, id string, isActive bool) (*model.User, error) {
	return nil, nil
}

func TestExtractRole(t *testing.T) {
	svc := &AuthService{}

	tests := []struct {
		name     string
		roles    []string
		expected string
	}{
		{"admin role", []string{"default-roles-cifo", "admin", "offline_access"}, "admin"},
		{"devops role", []string{"default-roles-cifo", "devops"}, "devops"},
		{"viewer role", []string{"default-roles-cifo", "viewer"}, "viewer"},
		{"unknown role defaults to viewer", []string{"some-other-role"}, "viewer"},
		{"empty roles defaults to viewer", []string{}, "viewer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &AuthClaims{
				RealmAccess: RealmRoles{Roles: tt.roles},
			}
			assert.Equal(t, tt.expected, svc.ExtractRole(claims))
		})
	}
}

func TestSyncUser(t *testing.T) {
	mockRepo := &mockUserRepo{}
	cfg := &config.Config{}
	svc := NewAuthService(cfg, mockRepo, nil, nil, nil, nil)

	claims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "kc-user-uuid-12345",
		},
		Email: "admin@cifo.local",
		Name:  "Admin CIFO",
		RealmAccess: RealmRoles{
			Roles: []string{"admin"},
		},
	}

	user, err := svc.SyncUser(context.Background(), claims)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin@cifo.local", user.Email)
	assert.Equal(t, "Admin CIFO", user.Name)
	assert.Equal(t, "admin", user.Role)
	assert.Equal(t, "kc-user-uuid-12345", user.KeycloakID)
	assert.True(t, user.IsActive)
}

func TestJWKSCache(t *testing.T) {
	cache := NewJWKSCache("http://localhost:8180/realms/cifo/protocol/openid-connect/certs", 1*time.Hour)
	assert.NotNil(t, cache)

	// inject a mock RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	cache.Lock()
	cache.keys["test-kid"] = &privateKey.PublicKey
	cache.expiresAt = time.Now().Add(1 * time.Hour)
	cache.Unlock()

	pubKey, err := cache.GetKey("test-kid")
	assert.NoError(t, err)
	assert.Equal(t, &privateKey.PublicKey, pubKey)
}

func TestValidateToken_DevTokens(t *testing.T) {
	cfg := &config.Config{Environment: "development", KeycloakIssuer: "http://test-issuer"}
	svc := NewAuthService(cfg, nil, nil, nil, nil, nil)

	claimsAdmin, err := svc.ValidateToken(context.Background(), "dev-token-admin")
	assert.NoError(t, err)
	assert.Equal(t, "admin@cifo.local", claimsAdmin.Email)
	assert.Equal(t, "admin", svc.ExtractRole(claimsAdmin))

	claimsDevops, err := svc.ValidateToken(context.Background(), "dev-token-devops")
	assert.NoError(t, err)
	assert.Equal(t, "devops", svc.ExtractRole(claimsDevops))

	claimsViewer, err := svc.ValidateToken(context.Background(), "dev-token-viewer")
	assert.NoError(t, err)
	assert.Equal(t, "viewer", svc.ExtractRole(claimsViewer))
}

func TestValidateToken_RSA(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	cache := NewJWKSCache("http://localhost:8180/realms/cifo/protocol/openid-connect/certs", 1*time.Hour)
	cache.Lock()
	cache.keys["key-1"] = &privateKey.PublicKey
	cache.expiresAt = time.Now().Add(1 * time.Hour)
	cache.Unlock()

	cfg := &config.Config{Environment: "production"}
	svc := NewAuthService(cfg, nil, nil, nil, cache, nil)

	// create token signed with private key
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &AuthClaims{
		Email: "prod@cifo.local",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "prod-sub-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})
	token.Header["kid"] = "key-1"
	tokenStr, err := token.SignedString(privateKey)
	assert.NoError(t, err)

	claims, err := svc.ValidateToken(context.Background(), tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, "prod@cifo.local", claims.Email)

	// test missing kid header
	badToken := jwt.NewWithClaims(jwt.SigningMethodRS256, &AuthClaims{})
	badTokenStr, err := badToken.SignedString(privateKey)
	assert.NoError(t, err)
	_, err = svc.ValidateToken(context.Background(), badTokenStr)
	assert.Error(t, err)
}

func TestSyncUser_FallbackName(t *testing.T) {
	mockRepo := &mockUserRepo{}
	cfg := &config.Config{}
	svc := NewAuthService(cfg, mockRepo, nil, nil, nil, nil)

	claims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "kc-user-2"},
		Email:            "dev@cifo.local",
		PreferredUsername: "devuser",
		RealmAccess:      RealmRoles{Roles: []string{"devops"}},
	}

	user, err := svc.SyncUser(context.Background(), claims)
	assert.NoError(t, err)
	assert.Equal(t, "devuser", user.Name)

	// when PreferredUsername is also empty, fallback to Email
	claims.PreferredUsername = ""
	user, err = svc.SyncUser(context.Background(), claims)
	assert.NoError(t, err)
	assert.Equal(t, "dev@cifo.local", user.Name)
}

func TestBlacklistToken_Expired(t *testing.T) {
	svc := &AuthService{}
	err := svc.BlacklistToken(context.Background(), "token-1", time.Now().Add(-1*time.Hour))
	assert.NoError(t, err)
}
