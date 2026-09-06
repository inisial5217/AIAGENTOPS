package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cifo-monitoring/backend/internal/config"
	"github.com/cifo-monitoring/backend/internal/model"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// RealmRoles keycloak realm roles
type RealmRoles struct {
	Roles []string `json:"roles"`
}

// AuthClaims parsed token claims
type AuthClaims struct {
	jwt.RegisteredClaims
	Email             string                 `json:"email"`
	Name              string                 `json:"name"`
	PreferredUsername string                 `json:"preferred_username"`
	RealmAccess       RealmRoles             `json:"realm_access"`
	ResourceAccess    map[string]RealmRoles `json:"resource_access"`
}

// AuthService handles authentication logic
type AuthService struct {
	cfg       *config.Config
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
	redis     *redis.Client
	jwks      *JWKSCache
	logger    *slog.Logger
	client    *http.Client
}

// NewAuthService creates auth service
func NewAuthService(
	cfg *config.Config,
	userRepo repository.UserRepository,
	auditRepo repository.AuditRepository,
	rdb *redis.Client,
	jwks *JWKSCache,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		cfg:       cfg,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		redis:     rdb,
		jwks:      jwks,
		logger:    logger,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// ValidateToken parses and validates token
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*AuthClaims, error) {
	// check development mock token fallback
	if s.cfg.Environment == "development" && strings.HasPrefix(tokenString, "dev-token-") {
		role := strings.TrimPrefix(tokenString, "dev-token-")
		var roles []string
		switch role {
		case "admin":
			roles = []string{"admin", "devops", "viewer"}
		case "devops":
			roles = []string{"devops", "viewer"}
		default:
			roles = []string{"viewer"}
		}
		return &AuthClaims{
			Email:             role + "@cifo.local",
			Name:              strings.ToUpper(string(role[0])) + role[1:] + " User",
			PreferredUsername: role,
			RealmAccess: RealmRoles{
				Roles: roles,
			},
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "dev-" + role + "-id",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    s.cfg.KeycloakIssuer,
			},
		}, nil
	}

	// check redis blacklist
	blacklisted, err := s.IsTokenBlacklisted(ctx, tokenString)
	if err != nil {
		s.logger.Warn("redis blacklist check failed", slog.String("err", err.Error()))
	} else if blacklisted {
		return nil, errors.New("token has been revoked")
	}

	// parse jwt with jwks key function
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("missing kid header in token")
		}

		return s.jwks.GetKey(kid)
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	return claims, nil
}

// ExtractRole extracts user role
func (s *AuthService) ExtractRole(claims *AuthClaims) string {
	for _, r := range claims.RealmAccess.Roles {
		switch strings.ToLower(r) {
		case "admin":
			return "admin"
		case "devops":
			return "devops"
		case "viewer":
			return "viewer"
		}
	}
	return "viewer"
}

// SyncUser syncs keycloak user
func (s *AuthService) SyncUser(ctx context.Context, claims *AuthClaims) (*model.User, error) {
	role := s.ExtractRole(claims)
	name := claims.Name
	if name == "" {
		name = claims.PreferredUsername
	}
	if name == "" {
		name = claims.Email
	}

	user := &model.User{
		Email:      claims.Email,
		Name:       name,
		Role:       role,
		KeycloakID: claims.Subject,
	}

	synced, err := s.userRepo.UpsertKeycloakUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("sync user in postgres: %w", err)
	}

	return synced, nil
}

// BlacklistToken adds token to redis
func (s *AuthService) BlacklistToken(ctx context.Context, tokenString string, exp time.Time) error {
	ttl := time.Until(exp)
	if ttl <= 0 {
		return nil
	}

	hash := sha256.Sum256([]byte(tokenString))
	key := fmt.Sprintf("jwt:blacklist:%s", hex.EncodeToString(hash[:]))

	return s.redis.Set(ctx, key, "revoked", ttl).Err()
}

// IsTokenBlacklisted checks token revocation
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	if s.redis == nil {
		return false, nil
	}
	hash := sha256.Sum256([]byte(tokenString))
	key := fmt.Sprintf("jwt:blacklist:%s", hex.EncodeToString(hash[:]))

	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// DirectLogin performs password grant
func (s *AuthService) DirectLogin(ctx context.Context, username, password string) (map[string]interface{}, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", s.cfg.KeycloakURL, s.cfg.KeycloakRealm)

	form := url.Values{}
	form.Set("client_id", s.cfg.KeycloakClientID)
	form.Set("grant_type", "password")
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		if s.cfg.Environment == "development" {
			if token := s.devLoginFallback(username, password); token != nil {
				return token, nil
			}
		}
		return nil, fmt.Errorf("execute login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read login response: %w", err)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("parse login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if s.cfg.Environment == "development" {
			if token := s.devLoginFallback(username, password); token != nil {
				return token, nil
			}
		}
		desc := "authentication failed"
		if errDesc, ok := res["error_description"].(string); ok {
			desc = errDesc
		}
		return nil, errors.New(desc)
	}

	return res, nil
}

// devLoginFallback provides resilient fallback in development mode
func (s *AuthService) devLoginFallback(username, password string) map[string]interface{} {
	validCredentials := map[string]struct {
		pass string
		role string
	}{
		"admin@cifo.local":  {pass: "admin123", role: "admin"},
		"devops@cifo.local": {pass: "devops123", role: "devops"},
		"viewer@cifo.local": {pass: "viewer123", role: "viewer"},
		"admin":             {pass: "admin123", role: "admin"},
		"devops":            {pass: "devops123", role: "devops"},
		"viewer":            {pass: "viewer123", role: "viewer"},
	}

	cred, ok := validCredentials[username]
	if !ok || cred.pass != password {
		return nil
	}

	return map[string]interface{}{
		"access_token":       "dev-token-" + cred.role,
		"token_type":         "Bearer",
		"expires_in":         86400,
		"refresh_expires_in": 86400,
		"scope":              "openid profile email",
	}
}

// LogAuditEvent records security audit log
func (s *AuthService) LogAuditEvent(ctx context.Context, log *model.AuditLog) error {
	if s.auditRepo == nil {
		return nil
	}
	return s.auditRepo.Create(ctx, log)
}
