package artists

import (
	"context"
	"errors"

	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// Artist domain: two-row pattern (no GSI).
// Main row: PK = ARTISTS#<handle>, SK = ARTIST — full artist data.
// User index row: PK = ARTISTS#USER#<user_id>, SK = ARTIST#<handle> — for ListByOwner (denormalized display_name for list view).

const (
	artistPKPrefix   = "ARTISTS#"
	artistSK         = "ARTIST"
	userIndexPKPrefix = "ARTISTS#USER#"
	userIndexSKPrefix = "ARTIST#"
)

type artistRow struct {
	PK            string `dynamodbav:"pk"`
	SK            string `dynamodbav:"sk"`
	Handle        string `dynamodbav:"handle"`
	DisplayName   string `dynamodbav:"display_name"`
	Bio           string `dynamodbav:"bio"`
	OwnerUserID   string `dynamodbav:"owner_user_id"`
	CreatedAt     string `dynamodbav:"created_at"`
	FollowerCount int    `dynamodbav:"follower_count,omitempty"`
}

type userIndexRow struct {
	PK          string `dynamodbav:"pk"`
	SK          string `dynamodbav:"sk"`
	Handle      string `dynamodbav:"handle"`
	DisplayName string `dynamodbav:"display_name"`
	CreatedAt   string `dynamodbav:"created_at"`
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

func artistPK(handle string) string {
	return artistPKPrefix + handle
}

func userIndexPK(userID string) string {
	return userIndexPKPrefix + userID
}

func userIndexSK(handle string) string {
	return userIndexSKPrefix + handle
}

// GetByHandle returns the artist for the given handle, or nil if not found.
func (s *Store) GetByHandle(ctx context.Context, handle string) (*artistRow, error) {
	var row artistRow
	err := s.tbl().Get("pk", artistPK(handle)).Range("sk", dynamo.Equal, artistSK).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListByOwner returns all artists owned by the user (from user index, denormalized).
func (s *Store) ListByOwner(ctx context.Context, userID string) ([]artistRow, error) {
	pk := userIndexPK(userID)
	var out []artistRow
	iter := s.tbl().Get("pk", pk).Range("sk", dynamo.BeginsWith, userIndexSKPrefix).Iter()
	var idx userIndexRow
	for iter.Next(ctx, dynamo.AWSEncoding(&idx)) {
		out = append(out, artistRow{
			Handle:      idx.Handle,
			DisplayName: idx.DisplayName,
			OwnerUserID: userID,
			CreatedAt:   idx.CreatedAt,
		})
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates an artist (main row + user index row) in one transaction.
// Fails if handle already exists (conditional put on main row).
func (s *Store) Create(ctx context.Context, handle, displayName, bio, ownerUserID, createdAt string) error {
	mainRow := artistRow{
		PK:            artistPK(handle),
		SK:            artistSK,
		Handle:        handle,
		DisplayName:   displayName,
		Bio:           bio,
		OwnerUserID:   ownerUserID,
		CreatedAt:     createdAt,
		FollowerCount: 0,
	}
	idxRow := userIndexRow{
		PK:          userIndexPK(ownerUserID),
		SK:          userIndexSK(handle),
		Handle:      handle,
		DisplayName: displayName,
		CreatedAt:   createdAt,
	}
	return s.db.WriteTx().
		Put(s.tbl().Put(dynamo.AWSEncoding(mainRow)).If("attribute_not_exists(pk)")).
		Put(s.tbl().Put(dynamo.AWSEncoding(idxRow)).If("attribute_not_exists(pk)")).
		Run(ctx)
}

// Update updates display_name and bio on main row; display_name on user index row.
func (s *Store) Update(ctx context.Context, handle, displayName, bio string) error {
	main, err := s.GetByHandle(ctx, handle)
	if err != nil || main == nil {
		return err
	}
	// Update main row (display_name + bio)
	if err := s.tbl().Update("pk", artistPK(handle)).Range("sk", artistSK).
		Set("display_name", displayName).
		Set("bio", bio).
		Run(ctx); err != nil {
		return err
	}
	// Update user index row so list shows correct display name
	return s.tbl().Update("pk", userIndexPK(main.OwnerUserID)).Range("sk", userIndexSK(handle)).
		Set("display_name", displayName).
		Run(ctx)
}

// Delete deletes the artist (main row + user index row). Requires fetching main row to get owner_user_id.
func (s *Store) Delete(ctx context.Context, handle string) error {
	main, err := s.GetByHandle(ctx, handle)
	if err != nil {
		return err
	}
	if main == nil {
		return nil
	}
	return s.db.WriteTx().
		Delete(s.tbl().Delete("pk", artistPK(handle)).Range("sk", artistSK)).
		Delete(s.tbl().Delete("pk", userIndexPK(main.OwnerUserID)).Range("sk", userIndexSK(handle))).
		Run(ctx)
}
