package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/dynamo/v2"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	ErrEmailTaken   = errors.New("email already registered")
	ErrInvalidCreds = errors.New("invalid email or password")
	ErrUserNotFound = errors.New("user not found")
)


type Service interface {
	Signup(ctx context.Context, email, password string) (userID string, err error)
	Login(ctx context.Context, email, password string) (userID string, err error)
	DeleteAccount(ctx context.Context, userID string) error
	GetByID(ctx context.Context, userID string) (*User, error)
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type service struct {
	store *Store
}

func NewService(store *Store) Service {
	return &service{store: store}
}

func (s *service) Signup(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)
	if email == "" || len(password) < 8 {
		return "", fmt.Errorf("email and password (min 8 chars) required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	existing, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", ErrEmailTaken
	}

	userID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.store.PutUser(ctx, userID, email, string(hash), now); err != nil {
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

	row, err := s.store.GetByEmail(ctx, email)
	if err != nil || row == nil {
		return "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCreds
	}
	return row.ID, nil
}

func (s *service) DeleteAccount(ctx context.Context, userID string) error {
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
