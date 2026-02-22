package artists

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/guregu/dynamo/v2"
)

var (
	ErrHandleTaken   = errors.New("handle already in use")
	ErrArtistNotFound = errors.New("artist not found")
	ErrForbidden     = errors.New("not the owner of this artist page")
)

// Handle must be lowercase, alphanumeric only, 4–64 chars (min 4 so we can reserve 3-letter subdomains: www, tui, api, etc.).
var handleRegex = regexp.MustCompile(`^[a-z0-9]{4,64}$`)

type Service interface {
	Create(ctx context.Context, ownerUserID, handle, displayName, bio string) (*Artist, error)
	GetByHandle(ctx context.Context, handle string) (*Artist, error)
	ListByOwner(ctx context.Context, userID string) ([]Artist, error)
	Update(ctx context.Context, handle string, displayName, bio *string, actorUserID string) (*Artist, error)
	Delete(ctx context.Context, handle, actorUserID string) error
}

type Artist struct {
	Handle        string `json:"handle"`
	DisplayName   string `json:"display_name"`
	Bio           string `json:"bio"`
	OwnerUserID   string `json:"owner_user_id"`
	CreatedAt     string `json:"created_at"`
	FollowerCount int    `json:"follower_count"`
}

type service struct {
	store *Store
}

func NewService(store *Store) Service {
	return &service{store: store}
}

func normalizeHandle(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (s *service) Create(ctx context.Context, ownerUserID, handle, displayName, bio string) (*Artist, error) {
	handle = normalizeHandle(handle)
	displayName = strings.TrimSpace(displayName)
	bio = strings.TrimSpace(bio)
	if handle == "" {
		return nil, fmt.Errorf("handle required")
	}
	if !handleRegex.MatchString(handle) {
		return nil, fmt.Errorf("handle must be 4–64 lowercase letters or numbers, no spaces or special characters")
	}
	if displayName == "" {
		displayName = handle
	}

	existing, err := s.store.GetByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrHandleTaken
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	if err := s.store.Create(ctx, handle, displayName, bio, ownerUserID, createdAt); err != nil {
		if dynamo.IsCondCheckFailed(err) {
			return nil, ErrHandleTaken
		}
		return nil, err
	}
	return &Artist{
		Handle:        handle,
		DisplayName:   displayName,
		Bio:           bio,
		OwnerUserID:   ownerUserID,
		CreatedAt:     createdAt,
		FollowerCount: 0,
	}, nil
}

func (s *service) GetByHandle(ctx context.Context, handle string) (*Artist, error) {
	handle = normalizeHandle(handle)
	if handle == "" {
		return nil, ErrArtistNotFound
	}
	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return nil, ErrArtistNotFound
	}
	return rowToArtist(row), nil
}

func (s *service) ListByOwner(ctx context.Context, userID string) ([]Artist, error) {
	if userID == "" {
		return nil, nil
	}
	rows, err := s.store.ListByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Artist, len(rows))
	for i := range rows {
		out[i] = *rowToArtist(&rows[i])
	}
	return out, nil
}

func (s *service) Update(ctx context.Context, handle string, displayName, bio *string, actorUserID string) (*Artist, error) {
	handle = normalizeHandle(handle)
	if handle == "" {
		return nil, ErrArtistNotFound
	}

	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return nil, ErrArtistNotFound
	}
	if row.OwnerUserID != actorUserID {
		return nil, ErrForbidden
	}

	resolvedDisplayName := row.DisplayName
	if displayName != nil {
		resolvedDisplayName = strings.TrimSpace(*displayName)
		if resolvedDisplayName == "" {
			resolvedDisplayName = row.DisplayName
		}
	}
	resolvedBio := row.Bio
	if bio != nil {
		resolvedBio = strings.TrimSpace(*bio)
	}

	if err := s.store.Update(ctx, handle, resolvedDisplayName, resolvedBio); err != nil {
		return nil, err
	}
	return &Artist{
		Handle:      handle,
		DisplayName: resolvedDisplayName,
		Bio:         resolvedBio,
		OwnerUserID: row.OwnerUserID,
		CreatedAt:   row.CreatedAt,
	}, nil
}

func (s *service) Delete(ctx context.Context, handle, actorUserID string) error {
	handle = normalizeHandle(handle)
	if handle == "" {
		return ErrArtistNotFound
	}

	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return ErrArtistNotFound
	}
	if row.OwnerUserID != actorUserID {
		return ErrForbidden
	}
	return s.store.Delete(ctx, handle)
}

func rowToArtist(r *artistRow) *Artist {
	if r == nil {
		return nil
	}
	return &Artist{
		Handle:        r.Handle,
		DisplayName:   r.DisplayName,
		Bio:           r.Bio,
		OwnerUserID:   r.OwnerUserID,
		CreatedAt:     r.CreatedAt,
		FollowerCount: r.FollowerCount,
	}
}
