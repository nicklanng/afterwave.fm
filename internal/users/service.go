package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/cognito"
)

var (
	ErrEmailTaken             = errors.New("email already registered")
	ErrInvalidCreds           = errors.New("invalid email or password")
	ErrUserNotFound           = errors.New("user not found")
	ErrAccountExistsWithPassword = errors.New("account already exists with email and password; use password login")
)


type Service interface {
	Signup(ctx context.Context, email, password string) (userID string, err error)
	Login(ctx context.Context, email, password string) (userID string, err error)
	DeleteAccount(ctx context.Context, userID string) error
	GetByID(ctx context.Context, userID string) (*User, error)
	EnsureUserForCognito(ctx context.Context, email, cognitoSub string) (userID string, err error)
	LinkCognitoSub(ctx context.Context, userID, cognitoSub string) error
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type service struct {
	store   *Store
	cognito cognito.Client
}

func NewService(store *Store, cognitoClient cognito.Client) Service {
	return &service{
		store:   store,
		cognito: cognitoClient,
	}
}

func (s *service) Signup(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)
	if email == "" || len(password) < 8 {
		return "", fmt.Errorf("email and password (min 8 chars) required")
	}

	existing, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", ErrEmailTaken
	}

	if s.cognito == nil {
		return "", fmt.Errorf("cognito client not configured")
	}

	cognitoSub, err := s.cognito.SignUp(ctx, email, password)
	if err != nil {
		return "", err
	}

	userID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.store.PutUser(ctx, userID, email, cognitoSub, now); err != nil {
		if dynamo.IsCondCheckFailed(err) {
			return "", ErrEmailTaken
		}
		return "", err
	}
	return userID, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)
	if email == "" {
		return "", ErrInvalidCreds
	}

	if s.cognito == nil {
		return "", fmt.Errorf("cognito client not configured")
	}

	cognitoSub, err := s.cognito.InitiateAuth(ctx, email, password)
	if err != nil {
		return "", ErrInvalidCreds
	}

	row, err := s.store.GetByCognitoSub(ctx, cognitoSub)
	if err != nil || row == nil {
		return "", ErrInvalidCreds
	}
	return row.ID, nil
}

func (s *service) DeleteAccount(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user id required")
	}
	row, err := s.store.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if row == nil {
		return nil
	}
	if s.cognito != nil {
		if err := s.cognito.AdminDeleteUser(ctx, row.Email); err != nil {
			return err
		}
	}
	return s.store.DeleteUser(ctx, userID)
}

func (s *service) GetByID(ctx context.Context, userID string) (*User, error) {
	row, err := s.store.GetByID(ctx, userID)
	if err != nil || row == nil {
		return nil, ErrUserNotFound
	}
	return &User{
		ID:        row.ID,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
	}, nil
}

// EnsureUserForCognito finds or creates a user for the given Cognito identity (email + sub).
// Used by federated login flows where Cognito has already authenticated the user.
// Returns ErrAccountExistsWithPassword if a user with this email already exists with a different
// Cognito identity (e.g. native signup), to prevent federated account takeover.
func (s *service) EnsureUserForCognito(ctx context.Context, email, cognitoSub string) (string, error) {
	email = normalizeEmail(email)
	if email == "" {
		return "", fmt.Errorf("email required")
	}

	row, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if row != nil {
		if row.CognitoSub != "" && row.CognitoSub != cognitoSub {
			// May be a linked IdP (e.g. user signed up with password then linked Google).
			linkedUser, err := s.store.GetByCognitoSub(ctx, cognitoSub)
			if err != nil {
				return "", err
			}
			if linkedUser != nil && linkedUser.ID == row.ID {
				return row.ID, nil
			}
			return "", ErrAccountExistsWithPassword
		}
		return row.ID, nil
	}

	userID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.store.PutUser(ctx, userID, email, cognitoSub, now); err != nil {
		if dynamo.IsCondCheckFailed(err) {
			row, err := s.store.GetByEmail(ctx, email)
			if err != nil {
				return "", err
			}
			if row != nil {
				return row.ID, nil
			}
		}
		return "", err
	}
	return userID, nil
}

// LinkCognitoSub links a Cognito sub (e.g. from Google/Apple IdP) to the given user. Used when an authenticated user adds a sign-in method.
func (s *service) LinkCognitoSub(ctx context.Context, userID, cognitoSub string) error {
	return s.store.AddLinkedCognitoSub(ctx, userID, cognitoSub)
}

func normalizeEmail(s string) string {
	b := []byte(s)
	start := 0
	for start < len(b) && (b[start] == ' ' || b[start] == '\t') {
		start++
	}
	end := len(b)
	for end > start && (b[end-1] == ' ' || b[end-1] == '\t') {
		end--
	}
	s = string(b[start:end])
	var out []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out = append(out, c)
	}
	return string(out)
}
