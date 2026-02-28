package follows

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// Two-row pattern (no GSI): user index + artist index.
// - User index: PK = FOLLOWS#USER#<user_id>, SK = <handle>, followed_at — for "list who I follow" and to get followed_at on unfollow.
// - Artist index: PK = ARTISTS#<handle>, SK = FOLLOWED#<followed_at>#<user_id> — for "list followers of artist" ordered by recent.

const (
	followsUserPKPrefix = "FOLLOWS#USER#"
	artistPKPrefix      = "ARTISTS#"
	artistMainSK        = "ARTIST" // main artist row SK (same table)
	followedSKPrefix    = "FOLLOWED#"
)

func userIndexPK(userID string) string {
	return followsUserPKPrefix + userID
}

func artistPK(handle string) string {
	return artistPKPrefix + normalizeHandle(handle)
}

// followedSK builds SK for the artist-side row so we can query by prefix and order by time (SK desc).
func followedSK(followedAt, userID string) string {
	return followedSKPrefix + followedAt + "#" + userID
}

type Store struct {
	db        *infra.Dynamo
	tableName string
}

func NewStore(db *infra.Dynamo, tableName string) *Store {
	return &Store{db: db, tableName: tableName}
}

func (s *Store) tbl() dynamo.Table {
	return s.db.Table(s.tableName)
}

func normalizeHandle(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Follow adds the follow relationship (both rows) in a transaction. Idempotent (no-op if already following).
// Returns inserted=true only when a new follow was written, so the caller can increment follower count once.
func (s *Store) Follow(ctx context.Context, userID string, handle string) (inserted bool, err error) {
	handle = normalizeHandle(handle)
	if userID == "" || handle == "" {
		return false, errors.New("user_id and handle required")
	}
	existing, _ := s.getUserFollowRow(ctx, userID, handle)
	if existing != nil {
		return false, nil
	}
	followedAt := time.Now().UTC().Format(time.RFC3339)
	userRow := struct {
		PK         string `dynamo:"pk"`
		SK         string `dynamo:"sk"`
		Handle     string `dynamo:"handle"`
		FollowedAt string `dynamo:"followed_at"`
	}{userIndexPK(userID), handle, handle, followedAt}
	artistFollowerRow := struct {
		PK         string `dynamo:"pk"`
		SK         string `dynamo:"sk"`
		UserID     string `dynamo:"user_id"`
		FollowedAt string `dynamo:"followed_at"`
	}{artistPK(handle), followedSK(followedAt, userID), userID, followedAt}
	// Single transaction: both follow rows + increment artist follower_count
	err = s.db.WriteTx().
		Put(s.tbl().Put(userRow)).
		Put(s.tbl().Put(artistFollowerRow)).
		Update(s.tbl().Update("pk", artistPK(handle)).Range("sk", artistMainSK).
			SetExpr("follower_count = if_not_exists(follower_count, ?) + ?", 0, 1)).
		Run(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

type userFollowRow struct {
	PK         string `dynamo:"pk"`
	SK         string `dynamo:"sk"`
	FollowedAt string `dynamo:"followed_at"`
}

func (s *Store) getUserFollowRow(ctx context.Context, userID string, handle string) (*userFollowRow, error) {
	var row userFollowRow
	err := s.tbl().Get("pk", userIndexPK(userID)).Range("sk", dynamo.Equal, handle).One(ctx, &row)
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// Unfollow removes both rows. Idempotent. Returns true if a follow was actually removed (so caller can decrement count).
func (s *Store) Unfollow(ctx context.Context, userID string, handle string) (removed bool, err error) {
	handle = normalizeHandle(handle)
	if userID == "" || handle == "" {
		return false, errors.New("user_id and handle required")
	}
	userRow, err := s.getUserFollowRow(ctx, userID, handle)
	if err != nil || userRow == nil {
		return false, err
	}
	sk := followedSK(userRow.FollowedAt, userID)
	// Single transaction: delete both follow rows + decrement artist follower_count (condition: count >= 1)
	err = s.db.WriteTx().
		Delete(s.tbl().Delete("pk", userIndexPK(userID)).Range("sk", handle)).
		Delete(s.tbl().Delete("pk", artistPK(handle)).Range("sk", sk)).
		Update(s.tbl().Update("pk", artistPK(handle)).Range("sk", artistMainSK).
			SetExpr("follower_count = follower_count + ?", -1).
			If("follower_count >= ?", 1)).
		Run(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ListFollowing returns all artist handles the user follows.
func (s *Store) ListFollowing(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, nil
	}
	pk := userIndexPK(userID)
	var out []string
	iter := s.tbl().Get("pk", pk).Iter()
	var row struct {
		PK     string `dynamo:"pk"`
		SK     string `dynamo:"sk"`
		Handle string `dynamo:"handle"`
	}
	for iter.Next(ctx, &row) {
		out = append(out, row.SK)
	}
	return out, iter.Err()
}

// ListFollowers returns follower user IDs for the artist, most recent first.
func (s *Store) ListFollowers(ctx context.Context, handle string, limit int) ([]string, error) {
	handle = normalizeHandle(handle)
	if handle == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	pk := artistPK(handle)
	var out []string
	iter := s.tbl().Get("pk", pk).Range("sk", dynamo.BeginsWith, followedSKPrefix).Order(dynamo.Descending).Limit(limit).Iter()
	var row struct {
		PK     string `dynamo:"pk"`
		SK     string `dynamo:"sk"`
		UserID string `dynamo:"user_id"`
	}
	for iter.Next(ctx, &row) {
		out = append(out, row.UserID)
	}
	return out, iter.Err()
}

// IsFollowing returns whether the user follows the artist.
func (s *Store) IsFollowing(ctx context.Context, userID string, handle string) (bool, error) {
	handle = normalizeHandle(handle)
	if userID == "" || handle == "" {
		return false, nil
	}
	row, err := s.getUserFollowRow(ctx, userID, handle)
	if err != nil || row == nil {
		return false, err
	}
	return true, nil
}
