package follows

import (
	"context"
	"errors"

	"github.com/sopatech/afterwave.fm/internal/artists"
)

var ErrArtistNotFound = errors.New("artist not found")

// ArtistResolver is used to verify artist exists when following.
type ArtistResolver interface {
	GetByHandle(ctx context.Context, handle string) (*artists.Artist, error)
}

type Service interface {
	Follow(ctx context.Context, userID string, handle string) error
	Unfollow(ctx context.Context, userID string, handle string) error
	ListFollowing(ctx context.Context, userID string) ([]string, error)
	IsFollowing(ctx context.Context, userID string, handle string) (bool, error)
}

type service struct {
	store  *Store
	artist ArtistResolver
}

func NewService(store *Store, artist ArtistResolver) Service {
	return &service{store: store, artist: artist}
}

func (s *service) Follow(ctx context.Context, userID string, handle string) error {
	handle = normalizeHandle(handle)
	artist, err := s.artist.GetByHandle(ctx, handle)
	if err != nil || artist == nil {
		return ErrArtistNotFound
	}
	_ = artist
	_, err = s.store.Follow(ctx, userID, handle)
	return err
}

func (s *service) Unfollow(ctx context.Context, userID string, handle string) error {
	handle = normalizeHandle(handle)
	_, err := s.store.Unfollow(ctx, userID, handle)
	return err
}

func (s *service) ListFollowing(ctx context.Context, userID string) ([]string, error) {
	return s.store.ListFollowing(ctx, userID)
}

func (s *service) IsFollowing(ctx context.Context, userID string, handle string) (bool, error) {
	handle = normalizeHandle(handle)
	return s.store.IsFollowing(ctx, userID, handle)
}
