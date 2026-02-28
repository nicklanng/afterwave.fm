package users

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// Users domain: two-row pattern (no GSI).
// Main row: PK = USERS#user,<first_char>, SK = USER#<id> — full user data.
// Email lookup row: PK = USERS#email#<shard>, SK = <email> — sharded by hash of email to avoid hot partition.
// Cognito sub lookup row: PK = USERS#cognito_sub#<first2>, SK = <sub> — for login-by-sub (sharded by first 2 chars of sub).
// Linked sub row (for cleanup on delete): PK = userPK(userID), SK = LINKED_SUB#<sub> — one per linked IdP.

const (
	usersPrefix        = "USERS#user,"
	userSKPrefix       = "USER#"
	emailPKPrefix      = "USERS#email#"
	emailShardLen      = 2 // first N hex chars of sha256; 2 => 256 partitions
	cognitoSubPKPrefix = "USERS#cognito_sub#"
	cognitoSubShardLen = 2   // first 2 chars of sub (UUID) for partition spread
	linkedSubSKPrefix  = "LINKED_SUB#"
)

// ErrSubLinkedToOtherAccount is returned by AddLinkedCognitoSub when the Cognito sub is already linked to a different user.
var ErrSubLinkedToOtherAccount = errors.New("this identity is already linked to another account")

type userRow struct {
	PK         string `dynamo:"pk"`
	SK         string `dynamo:"sk"`
	ID         string `dynamo:"id"`
	Email      string `dynamo:"email"`
	CognitoSub string `dynamo:"cognito_sub,omitempty"`
	CreatedAt  string `dynamo:"created_at"`
}

// emailLookupRow is the second row for email→user_id lookup (no GSI).
type emailLookupRow struct {
	PK     string `dynamo:"pk"`
	SK     string `dynamo:"sk"`
	UserID string `dynamo:"user_id"`
}

// cognitoSubLookupRow is the row for cognito_sub→user_id lookup (sharded by first 2 chars of sub).
type cognitoSubLookupRow struct {
	PK     string `dynamo:"pk"`
	SK     string `dynamo:"sk"`
	UserID string `dynamo:"user_id"`
}

// linkedSubRow is stored under the user so we can delete all linked subs when the user is deleted.
type linkedSubRow struct {
	PK string `dynamo:"pk"`
	SK string `dynamo:"sk"`
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

func userPK(userID string) string {
	if userID == "" {
		return ""
	}
	return usersPrefix + string(userID[0])
}

func userSK(userID string) string {
	return userSKPrefix + userID
}

// emailShard returns the first emailShardLen hex chars of sha256(email) to partition email lookups.
func emailShard(email string) string {
	h := sha256.Sum256([]byte(email))
	return hex.EncodeToString(h[:])[:emailShardLen]
}

func emailPK(email string) string {
	return emailPKPrefix + emailShard(email)
}

func cognitoSubPK(sub string) string {
	if len(sub) < cognitoSubShardLen {
		return cognitoSubPKPrefix + sub
	}
	return cognitoSubPKPrefix + sub[:cognitoSubShardLen]
}

// GetByEmail returns the user row for the given email, or nil if not found.
// Email must be normalized (lowercase). Uses sharded email lookup row then main row (2 GetItems, no GSI).
func (s *Store) GetByEmail(ctx context.Context, email string) (*userRow, error) {
	var emailRow emailLookupRow
	err := s.tbl().Get("pk", emailPK(email)).Range("sk", dynamo.Equal, email).One(ctx, &emailRow)
	if err != nil {
		if !errors.Is(err, dynamo.ErrNotFound) {
			return nil, err
		}
		return nil, nil
	}
	if emailRow.UserID == "" {
		return nil, nil
	}
	return s.GetByID(ctx, emailRow.UserID)
}

// GetByCognitoSub returns the user row for the given Cognito sub, or nil if not found.
func (s *Store) GetByCognitoSub(ctx context.Context, cognitoSub string) (*userRow, error) {
	if cognitoSub == "" {
		return nil, nil
	}
	var lookup cognitoSubLookupRow
	err := s.tbl().Get("pk", cognitoSubPK(cognitoSub)).Range("sk", dynamo.Equal, cognitoSub).One(ctx, &lookup)
	if err != nil {
		if !errors.Is(err, dynamo.ErrNotFound) {
			return nil, err
		}
		return nil, nil
	}
	if lookup.UserID == "" {
		return nil, nil
	}
	return s.GetByID(ctx, lookup.UserID)
}

// AddLinkedCognitoSub links a Cognito sub (e.g. from Google/Apple) to an existing user so they can sign in with that IdP.
// Idempotent: if the sub is already linked to this user, returns nil. Returns error if sub is linked to another user.
func (s *Store) AddLinkedCognitoSub(ctx context.Context, userID, cognitoSub string) error {
	if userID == "" || cognitoSub == "" {
		return fmt.Errorf("user id and cognito sub required")
	}
	row, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("user not found")
	}
	existing, err := s.GetByCognitoSub(ctx, cognitoSub)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != userID {
		return ErrSubLinkedToOtherAccount
	}
	if existing != nil && existing.ID == userID {
		return nil // already linked
	}
	lookupRow := cognitoSubLookupRow{
		PK:     cognitoSubPK(cognitoSub),
		SK:     cognitoSub,
		UserID: userID,
	}
	linkRow := linkedSubRow{
		PK: userPK(userID),
		SK: linkedSubSKPrefix + cognitoSub,
	}
	return s.db.WriteTx().
		Put(s.tbl().Put(lookupRow)).
		Put(s.tbl().Put(linkRow)).
		Run(ctx)
}

// listLinkedCognitoSubs returns all linked cognito subs for the user (for cleanup on delete).
func (s *Store) listLinkedCognitoSubs(ctx context.Context, userID string) ([]string, error) {
	pk := userPK(userID)
	var subs []string
	var linkRow linkedSubRow
	iter := s.tbl().Get("pk", pk).Range("sk", dynamo.BeginsWith, linkedSubSKPrefix).Iter()
	for iter.Next(ctx, &linkRow) {
		sub := strings.TrimPrefix(linkRow.SK, linkedSubSKPrefix)
		if sub != "" {
			subs = append(subs, sub)
		}
	}
	return subs, iter.Err()
}

// GetByID returns the user row for the given user ID, or nil if not found.
func (s *Store) GetByID(ctx context.Context, userID string) (*userRow, error) {
	var row userRow
	err := s.tbl().Get("pk", userPK(userID)).Range("sk", dynamo.Equal, userSK(userID)).One(ctx, &row)
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// PutUser creates a user (main row + email lookup row + optional cognito_sub lookup) in one transaction.
// Email must be normalized (lowercase). Fails if a user with the same ID already exists.
func (s *Store) PutUser(ctx context.Context, userID, email, cognitoSub, createdAt string) error {
	mainRow := userRow{
		PK:         userPK(userID),
		SK:         userSK(userID),
		ID:         userID,
		Email:      email,
		CognitoSub: cognitoSub,
		CreatedAt:  createdAt,
	}
	emailRow := emailLookupRow{
		PK:     emailPK(email),
		SK:     email,
		UserID: userID,
	}
	tx := s.db.WriteTx().
		Put(s.tbl().Put(mainRow).If("attribute_not_exists(pk)")).
		Put(s.tbl().Put(emailRow).If("attribute_not_exists(pk)"))
	if cognitoSub != "" {
		subRow := cognitoSubLookupRow{
			PK:     cognitoSubPK(cognitoSub),
			SK:     cognitoSub,
			UserID: userID,
		}
		tx = tx.Put(s.tbl().Put(subRow).If("attribute_not_exists(pk)"))
	}
	return tx.Run(ctx)
}

// DeleteUser deletes the user by ID (main row + email lookup row + primary cognito_sub + all linked cognito_sub rows).
func (s *Store) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user id required")
	}
	row, err := s.GetByID(ctx, userID)
	if err != nil || row == nil {
		return err
	}
	linkedSubs, err := s.listLinkedCognitoSubs(ctx, userID)
	if err != nil {
		return err
	}
	tx := s.db.WriteTx().
		Delete(s.tbl().Delete("pk", userPK(userID)).Range("sk", userSK(userID))).
		Delete(s.tbl().Delete("pk", emailPK(row.Email)).Range("sk", row.Email))
	if row.CognitoSub != "" {
		tx = tx.Delete(s.tbl().Delete("pk", cognitoSubPK(row.CognitoSub)).Range("sk", row.CognitoSub))
	}
	for _, sub := range linkedSubs {
		tx = tx.
			Delete(s.tbl().Delete("pk", cognitoSubPK(sub)).Range("sk", sub)).
			Delete(s.tbl().Delete("pk", userPK(userID)).Range("sk", linkedSubSKPrefix+sub))
	}
	return tx.Run(ctx)
}
