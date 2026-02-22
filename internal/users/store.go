package users

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// Users domain: two-row pattern (no GSI).
// Main row: PK = USERS#user,<first_char>, SK = USER#<id> — full user data.
// Email lookup row: PK = USERS#email#<shard>, SK = <email> — sharded by hash of email to avoid hot partition.

const (
	usersPrefix   = "USERS#user,"
	userSKPrefix  = "USER#"
	emailPKPrefix = "USERS#email#"
	emailShardLen = 2 // first N hex chars of sha256; 2 => 256 partitions
)

type userRow struct {
	PK           string `dynamodbav:"pk"`
	SK           string `dynamodbav:"sk"`
	ID           string `dynamodbav:"id"`
	Email        string `dynamodbav:"email"`
	PasswordHash string `dynamodbav:"password_hash"`
	CreatedAt    string `dynamodbav:"created_at"`
}

// emailLookupRow is the second row for email→user_id lookup (no GSI).
type emailLookupRow struct {
	PK     string `dynamodbav:"pk"`
	SK     string `dynamodbav:"sk"`
	UserID string `dynamodbav:"user_id"`
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

// GetByEmail returns the user row for the given email, or nil if not found.
// Email must be normalized (lowercase). Uses sharded email lookup row then main row (2 GetItems, no GSI).
func (s *Store) GetByEmail(ctx context.Context, email string) (*userRow, error) {
	var emailRow emailLookupRow
	err := s.tbl().Get("pk", emailPK(email)).Range("sk", dynamo.Equal, email).One(ctx, dynamo.AWSEncoding(&emailRow))
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

// GetByID returns the user row for the given user ID, or nil if not found.
func (s *Store) GetByID(ctx context.Context, userID string) (*userRow, error) {
	var row userRow
	err := s.tbl().Get("pk", userPK(userID)).Range("sk", dynamo.Equal, userSK(userID)).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// PutUser creates or overwrites a user (main row + email lookup row) in one transaction.
// Email must be normalized (lowercase).
func (s *Store) PutUser(ctx context.Context, userID, email, passwordHash, createdAt string) error {
	mainRow := userRow{
		PK:           userPK(userID),
		SK:           userSK(userID),
		ID:           userID,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
	}
	emailRow := emailLookupRow{
		PK:     emailPK(email),
		SK:     email,
		UserID: userID,
	}
	return s.db.WriteTx().
		Put(s.tbl().Put(dynamo.AWSEncoding(mainRow)).If("attribute_not_exists(pk)")).
		Put(s.tbl().Put(dynamo.AWSEncoding(emailRow)).If("attribute_not_exists(pk)")).
		Run(ctx)
}

// DeleteUser deletes the user by ID (main row + email lookup row). Requires fetching user first to get email.
func (s *Store) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user id required")
	}
	row, err := s.GetByID(ctx, userID)
	if err != nil || row == nil {
		return err
	}
	return s.db.WriteTx().
		Delete(s.tbl().Delete("pk", userPK(userID)).Range("sk", userSK(userID))).
		Delete(s.tbl().Delete("pk", emailPK(row.Email)).Range("sk", row.Email)).
		Run(ctx)
}
