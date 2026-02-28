package artists

import (
	"context"
	"errors"

	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// Member store: two-row pattern (no GSI).
// Artist side: PK = ARTISTS#<handle>, SK = MEMBER#<user_id> — roles list.
// User side: PK = ARTIST_MEMBERS#USER#<user_id>, SK = ARTIST#<handle> — for ListByUser (denormalized).

const (
	memberSKPrefix       = "MEMBER#"
	artistMembersPKPrefix = "ARTIST_MEMBERS#USER#"
	artistMembersSKPrefix = "ARTIST#"
)

var ErrMemberNotFound = errors.New("member not found")

type memberRow struct {
	PK     string   `dynamo:"pk"`
	SK     string   `dynamo:"sk"`
	UserID string   `dynamo:"user_id"`
	Roles  []string `dynamo:"roles"`
}

type memberUserIndexRow struct {
	PK     string   `dynamo:"pk"`
	SK     string   `dynamo:"sk"`
	Handle string   `dynamo:"handle"`
	Roles  []string `dynamo:"roles"`
}

type MemberStore struct {
	db        *infra.Dynamo
	tableName string
}

func NewMemberStore(db *infra.Dynamo, tableName string) *MemberStore {
	return &MemberStore{db: db, tableName: tableName}
}

func (s *MemberStore) tbl() dynamo.Table {
	return s.db.Table(s.tableName)
}

func memberPK(handle string) string {
	return artistPKPrefix + handle
}

func memberSK(userID string) string {
	return memberSKPrefix + userID
}

func memberUserIndexPK(userID string) string {
	return artistMembersPKPrefix + userID
}

func memberUserIndexSK(handle string) string {
	return artistMembersSKPrefix + handle
}

// Get returns the member row for (handle, userID), or nil if not found.
func (s *MemberStore) Get(ctx context.Context, handle, userID string) (*memberRow, error) {
	var row memberRow
	err := s.tbl().Get("pk", memberPK(handle)).Range("sk", dynamo.Equal, memberSK(userID)).One(ctx, &row)
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListByArtist returns all members (user_id + roles) for the artist. Does not include the owner.
func (s *MemberStore) ListByArtist(ctx context.Context, handle string) ([]memberRow, error) {
	pk := memberPK(handle)
	var out []memberRow
	iter := s.tbl().Get("pk", pk).Range("sk", dynamo.BeginsWith, memberSKPrefix).Iter()
	var row memberRow
	for iter.Next(ctx, &row) {
		out = append(out, row)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListByUser returns all artist memberships (handle + roles) for the user.
func (s *MemberStore) ListByUser(ctx context.Context, userID string) ([]memberUserIndexRow, error) {
	pk := memberUserIndexPK(userID)
	var out []memberUserIndexRow
	iter := s.tbl().Get("pk", pk).Range("sk", dynamo.BeginsWith, artistMembersSKPrefix).Iter()
	var row memberUserIndexRow
	for iter.Next(ctx, &row) {
		out = append(out, row)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Put adds or overwrites membership (both rows) in a transaction.
func (s *MemberStore) Put(ctx context.Context, handle, userID string, roles []string) error {
	if len(roles) == 0 {
		return s.Delete(ctx, handle, userID)
	}
	mainRow := memberRow{
		PK:     memberPK(handle),
		SK:     memberSK(userID),
		UserID: userID,
		Roles:  roles,
	}
	idxRow := memberUserIndexRow{
		PK:     memberUserIndexPK(userID),
		SK:     memberUserIndexSK(handle),
		Handle: handle,
		Roles:  roles,
	}
	return s.db.WriteTx().
		Put(s.tbl().Put(mainRow)).
		Put(s.tbl().Put(idxRow)).
		Run(ctx)
}

// Delete removes the membership (both rows) in a transaction.
func (s *MemberStore) Delete(ctx context.Context, handle, userID string) error {
	return s.db.WriteTx().
		Delete(s.tbl().Delete("pk", memberPK(handle)).Range("sk", memberSK(userID))).
		Delete(s.tbl().Delete("pk", memberUserIndexPK(userID)).Range("sk", memberUserIndexSK(handle))).
		Run(ctx)
}
