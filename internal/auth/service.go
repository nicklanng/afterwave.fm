package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")

// TokenPair is session token (JWT) + refresh token (opaque) returned from login/signup/refresh.
type TokenPair struct {
	SessionToken     string    `json:"session_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	ExpiresIn        int       `json:"expires_in"`         // seconds until session expiry
	RefreshExpiresIn int       `json:"refresh_expires_in"` // seconds until refresh expiry (for cookie max-age)
}

type Service struct {
	store      *Store
	privateKey *rsa.PrivateKey
}

func NewService(store *Store, privateKey *rsa.PrivateKey) *Service {
	return &Service{store: store, privateKey: privateKey}
}

// AuthCodeTTL is how long an authorization code is valid.
const AuthCodeTTL = 5 * time.Minute

// GetClientTTLs returns TTLs for a public client by client_id. Returns nil if client not found.
func (s *Service) GetClientTTLs(ctx context.Context, clientID string) (*ClientTTLs, error) {
	ttls, err := s.store.GetClientTTLs(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if ttls.SessionTTL == 0 && ttls.RefreshTTL == 0 {
		return nil, nil
	}
	return &ttls, nil
}

// CreateAuthCode creates a one-time auth code for the user/client and PKCE challenge. Returns code and expires_in seconds.
func (s *Service) CreateAuthCode(ctx context.Context, userID, clientID, codeChallenge, codeChallengeMethod string) (code string, expiresIn int, err error) {
	code = uuid.New().String()
	if err := s.store.CreateAuthCode(ctx, code, codeChallenge, codeChallengeMethod, userID, clientID, AuthCodeTTL); err != nil {
		return "", 0, err
	}
	return code, int(AuthCodeTTL.Seconds()), nil
}

// ExchangeCode exchanges an authorization code + code_verifier for tokens (PKCE). Validates client_id matches code.
func (s *Service) ExchangeCode(ctx context.Context, code, codeVerifier, clientID string) (*TokenPair, error) {
	data, err := s.store.GetAuthCodeAndDelete(ctx, code)
	if err != nil {
		if errors.Is(err, ErrAuthCodeInvalid) {
			return nil, err
		}
		return nil, err
	}
	if data.ClientID != clientID {
		return nil, ErrAuthCodeInvalid
	}
	if !VerifyCodeVerifier(codeVerifier, data.CodeChallenge, data.CodeChallengeMethod) {
		return nil, ErrAuthCodeInvalid
	}
	ttls, err := s.GetClientTTLs(ctx, clientID)
	if err != nil || ttls == nil {
		return nil, ErrAuthCodeInvalid
	}
	return s.NewSession(ctx, data.UserID, ttls)
}

// NewSession creates a session and refresh token for the user with the given TTLs (from auth client row).
func (s *Service) NewSession(ctx context.Context, userID string, ttls *ClientTTLs) (*TokenPair, error) {
	sessionID, refreshID, expiresAt, err := s.store.CreateSession(ctx, userID, ttls.SessionTTL, ttls.RefreshTTL)
	if err != nil {
		return nil, err
	}
	sessionToken, err := s.signSession(userID, sessionID, ttls.SessionTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		SessionToken:     sessionToken,
		RefreshToken:     refreshID,
		ExpiresAt:        expiresAt,
		ExpiresIn:        int(ttls.SessionTTL.Seconds()),
		RefreshExpiresIn: int(ttls.RefreshTTL.Seconds()),
	}, nil
}

// Refresh consumes the refresh token, revokes it, creates new session+refresh with the given TTLs, returns new TokenPair.
func (s *Service) Refresh(ctx context.Context, refreshID string, ttls *ClientTTLs) (*TokenPair, error) {
	userID, _, err := s.store.GetRefresh(ctx, refreshID)
	if err != nil || userID == "" {
		return nil, ErrInvalidRefreshToken
	}
	if err := s.store.RevokeRefresh(ctx, refreshID); err != nil {
		return nil, err
	}
	return s.NewSession(ctx, userID, ttls)
}

// Logout revokes the session (and its linked refresh token).
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.store.RevokeSession(ctx, sessionID)
}

// RevokeAllSessionsForUser revokes every session and refresh token for the user. Call before deleting the user account.
func (s *Service) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	return s.store.RevokeAllSessionsForUser(ctx, userID)
}

func (s *Service) signSession(userID, sessionID string, sessionTTL time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ID:        sessionID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(sessionTTL)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tok.SignedString(s.privateKey)
}
