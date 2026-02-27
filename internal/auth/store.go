package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// Auth domain: sessions, refresh tokens, auth codes, and registered clients live under AUTH#...
// Session: PK = AUTH#SESSION#<id>, SK = SESSION
// Refresh: PK = AUTH#REFRESH#<id>, SK = REFRESH
// User session index (no GSI): PK = AUTH#USER#<user_id>, SK = SESSION#<session_id> or REFRESH#<refresh_id> — for RevokeAllSessionsForUser
// Auth code: PK = AUTH#CODE#<code>, SK = CODE (one-time use, short-lived)
// Client:  PK = AUTH#CLIENT, SK = CLIENT#<client_id>

var ErrAuthCodeInvalid = errors.New("invalid or expired authorization code")

const (
	sessionPrefix     = "AUTH#SESSION#"
	refreshPrefix     = "AUTH#REFRESH#"
	userIndexPKPrefix = "AUTH#USER#"
	userIndexSession  = "SESSION#"
	userIndexRefresh  = "REFRESH#"
	codePrefix        = "AUTH#CODE#"
	clientPK          = "AUTH#CLIENT"
	clientSKPrefix   = "CLIENT#"
	sessionSK         = "SESSION"
	refreshSK         = "REFRESH"
	codeSK            = "CODE"
)

type sessionRow struct {
	PK        string `dynamodbav:"pk"`
	SK        string `dynamodbav:"sk"`
	UserID    string `dynamodbav:"user_id"`
	RefreshID string `dynamodbav:"refresh_id"`
	ExpiresAt string `dynamodbav:"expires_at"`
}

type refreshRow struct {
	PK        string `dynamodbav:"pk"`
	SK        string `dynamodbav:"sk"`
	UserID    string `dynamodbav:"user_id"`
	SessionID string `dynamodbav:"session_id"`
	ExpiresAt string `dynamodbav:"expires_at"`
}

type clientRow struct {
	PK                 string `dynamodbav:"pk"`
	SK                 string `dynamodbav:"sk"`
	ClientID           string `dynamodbav:"client_id"`
	SessionTTLSeconds  int    `dynamodbav:"session_ttl_seconds"`
	RefreshTTLSeconds  int    `dynamodbav:"refresh_ttl_seconds"`
}

// userSessionIndexRow is a minimal row for the user→session index (PK = AUTH#USER#<userID>, SK = SESSION#<id> or REFRESH#<id>).
type userSessionIndexRow struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
}

type authCodeRow struct {
	PK                  string `dynamodbav:"pk"`
	SK                  string `dynamodbav:"sk"`
	CodeChallenge       string `dynamodbav:"code_challenge"`
	CodeChallengeMethod string `dynamodbav:"code_challenge_method"`
	UserID              string `dynamodbav:"user_id"`
	ClientID            string `dynamodbav:"client_id"`
	ExpiresAt           string `dynamodbav:"expires_at"`
	ConsumedAt          string `dynamodbav:"consumed_at,omitempty"`
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

// CreateSession creates a session and linked refresh token, returns sessionID, refreshID, expiresAt.
func (s *Store) CreateSession(ctx context.Context, userID string, sessionTTL, refreshTTL time.Duration) (sessionID, refreshID string, sessionExpiresAt time.Time, err error) {
	sessionID = uuid.New().String()
	refreshID = uuid.New().String()
	sessionExpiresAt = time.Now().UTC().Add(sessionTTL)
	refreshExpiresAt := time.Now().UTC().Add(refreshTTL)

	sessRow := sessionRow{
		PK:        sessionPrefix + sessionID,
		SK:        sessionSK,
		UserID:    userID,
		RefreshID: refreshID,
		ExpiresAt: sessionExpiresAt.Format(time.RFC3339),
	}
	refRow := refreshRow{
		PK:        refreshPrefix + refreshID,
		SK:        refreshSK,
		UserID:    userID,
		SessionID: sessionID,
		ExpiresAt: refreshExpiresAt.Format(time.RFC3339),
	}
	userPK := userIndexPKPrefix + userID
	idxSess := userSessionIndexRow{PK: userPK, SK: userIndexSession + sessionID}
	idxRef := userSessionIndexRow{PK: userPK, SK: userIndexRefresh + refreshID}

	err = s.db.WriteTx().
		Put(s.tbl().Put(dynamo.AWSEncoding(sessRow)).If("attribute_not_exists(pk)")).
		Put(s.tbl().Put(dynamo.AWSEncoding(refRow)).If("attribute_not_exists(pk)")).
		Put(s.tbl().Put(dynamo.AWSEncoding(idxSess))).
		Put(s.tbl().Put(dynamo.AWSEncoding(idxRef))).
		Run(ctx)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return sessionID, refreshID, sessionExpiresAt, nil
}

// GetRefresh returns userID and sessionID for a refresh token, or empty if not found/expired.
func (s *Store) GetRefresh(ctx context.Context, refreshID string) (userID, sessionID string, err error) {
	var row refreshRow
	err = s.tbl().Get("pk", refreshPrefix+refreshID).Range("sk", dynamo.Equal, refreshSK).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return "", "", nil
		}
		return "", "", err
	}
	expiresAt, parseErr := time.Parse(time.RFC3339, row.ExpiresAt)
	if parseErr != nil || time.Now().UTC().After(expiresAt) {
		return "", "", nil
	}
	return row.UserID, row.SessionID, nil
}

// GetSession returns userID and refreshID for a session, or empty if not found.
func (s *Store) GetSession(ctx context.Context, sessionID string) (userID, refreshID string, err error) {
	var row sessionRow
	err = s.tbl().Get("pk", sessionPrefix+sessionID).Range("sk", dynamo.Equal, sessionSK).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return "", "", nil
		}
		return "", "", err
	}
	return row.UserID, row.RefreshID, nil
}

// RevokeRefresh deletes the refresh token and its linked session and their user index rows.
func (s *Store) RevokeRefresh(ctx context.Context, refreshID string) error {
	var row refreshRow
	err := s.tbl().Get("pk", refreshPrefix+refreshID).Range("sk", dynamo.Equal, refreshSK).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil
		}
		return err
	}
	userPK := userIndexPKPrefix + row.UserID
	return s.db.WriteTx().
		Delete(s.tbl().Delete("pk", sessionPrefix+row.SessionID).Range("sk", sessionSK)).
		Delete(s.tbl().Delete("pk", refreshPrefix+refreshID).Range("sk", refreshSK)).
		Delete(s.tbl().Delete("pk", userPK).Range("sk", userIndexSession+row.SessionID)).
		Delete(s.tbl().Delete("pk", userPK).Range("sk", userIndexRefresh+refreshID)).
		Run(ctx)
}

// RevokeSession deletes the session and its linked refresh token and their user index rows.
func (s *Store) RevokeSession(ctx context.Context, sessionID string) error {
	var row sessionRow
	err := s.tbl().Get("pk", sessionPrefix+sessionID).Range("sk", dynamo.Equal, sessionSK).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil
		}
		return err
	}
	userPK := userIndexPKPrefix + row.UserID
	return s.db.WriteTx().
		Delete(s.tbl().Delete("pk", sessionPrefix+sessionID).Range("sk", sessionSK)).
		Delete(s.tbl().Delete("pk", refreshPrefix+row.RefreshID).Range("sk", refreshSK)).
		Delete(s.tbl().Delete("pk", userPK).Range("sk", userIndexSession+sessionID)).
		Delete(s.tbl().Delete("pk", userPK).Range("sk", userIndexRefresh+row.RefreshID)).
		Run(ctx)
}

// RevokeAllSessionsForUser revokes every session (and linked refresh token) for the user. Used on account deletion.
func (s *Store) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	pk := userIndexPKPrefix + userID
	var idxRow userSessionIndexRow
	iter := s.tbl().Get("pk", pk).Range("sk", dynamo.BeginsWith, userIndexSession).Iter()
	for iter.Next(ctx, dynamo.AWSEncoding(&idxRow)) {
		sessionID := strings.TrimPrefix(idxRow.SK, userIndexSession)
		if sessionID == "" {
			continue
		}
		if err := s.RevokeSession(ctx, sessionID); err != nil {
			return err
		}
	}
	return iter.Err()
}

// ClientTTLs is the session and refresh TTL for an auth client (from DynamoDB).
type ClientTTLs struct {
	SessionTTL time.Duration
	RefreshTTL time.Duration
}

// GetClientTTLs returns TTLs for the client (public clients, no secret). Returns zero TTLs and nil error if not found.
func (s *Store) GetClientTTLs(ctx context.Context, clientID string) (ClientTTLs, error) {
	var row clientRow
	err := s.tbl().Get("pk", clientPK).Range("sk", dynamo.Equal, clientSKPrefix+clientID).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return ClientTTLs{}, nil
		}
		return ClientTTLs{}, err
	}
	return ClientTTLs{
		SessionTTL: time.Duration(row.SessionTTLSeconds) * time.Second,
		RefreshTTL: time.Duration(row.RefreshTTLSeconds) * time.Second,
	}, nil
}

// CreateClient stores a public auth client (no secret) with TTLs. Idempotent: overwrites if client exists.
func (s *Store) CreateClient(ctx context.Context, clientID string, sessionTTLSeconds, refreshTTLSeconds int) error {
	row := clientRow{
		PK:                 clientPK,
		SK:                 clientSKPrefix + clientID,
		ClientID:           clientID,
		SessionTTLSeconds:  sessionTTLSeconds,
		RefreshTTLSeconds:  refreshTTLSeconds,
	}
	return s.tbl().Put(dynamo.AWSEncoding(row)).Run(ctx)
}

// ClientCredential is used to seed auth clients (e.g. on startup). Public clients only (no secret).
type ClientCredential struct {
	ID                 string
	SessionTTLSeconds  int
	RefreshTTLSeconds  int
}

// EnsureAuthClients creates or updates each client so stored TTLs match the given credentials. Idempotent.
func (s *Store) EnsureAuthClients(ctx context.Context, clients []ClientCredential) error {
	for _, c := range clients {
		if c.ID == "" {
			continue
		}
		ttls, err := s.GetClientTTLs(ctx, c.ID)
		if err != nil {
			return err
		}
		wantSession := time.Duration(c.SessionTTLSeconds) * time.Second
		wantRefresh := time.Duration(c.RefreshTTLSeconds) * time.Second
		missing := ttls.SessionTTL == 0 && ttls.RefreshTTL == 0
		differs := ttls.SessionTTL != wantSession || ttls.RefreshTTL != wantRefresh
		if missing || differs {
			if err := s.CreateClient(ctx, c.ID, c.SessionTTLSeconds, c.RefreshTTLSeconds); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateAuthCode stores a one-time auth code with PKCE challenge. TTL is how long the code is valid.
func (s *Store) CreateAuthCode(ctx context.Context, code, codeChallenge, codeChallengeMethod, userID, clientID string, ttl time.Duration) error {
	expiresAt := time.Now().UTC().Add(ttl)
	row := authCodeRow{
		PK:                  codePrefix + code,
		SK:                  codeSK,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		UserID:              userID,
		ClientID:            clientID,
		ExpiresAt:           expiresAt.Format(time.RFC3339),
	}
	return s.tbl().Put(dynamo.AWSEncoding(row)).If("attribute_not_exists(pk)").Run(ctx)
}

// AuthCodeData is the payload stored with an auth code (returned when consuming the code).
type AuthCodeData struct {
	CodeChallenge       string
	CodeChallengeMethod string
	UserID              string
	ClientID            string
}

// GetAuthCodeAndDelete retrieves the auth code and marks it consumed (one-time use). Returns ErrAuthCodeInvalid if not found, expired, or already consumed.
// Uses a conditional update so only one exchange can succeed under race.
func (s *Store) GetAuthCodeAndDelete(ctx context.Context, code string) (AuthCodeData, error) {
	tbl := s.tbl()
	pk := codePrefix + code
	var row authCodeRow
	err := tbl.Get("pk", pk).Range("sk", dynamo.Equal, codeSK).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return AuthCodeData{}, ErrAuthCodeInvalid
		}
		return AuthCodeData{}, err
	}
	expiresAt, _ := time.Parse(time.RFC3339, row.ExpiresAt)
	if time.Now().UTC().After(expiresAt) {
		_ = tbl.Delete("pk", pk).Range("sk", codeSK).Run(ctx)
		return AuthCodeData{}, ErrAuthCodeInvalid
	}
	if row.ConsumedAt != "" {
		return AuthCodeData{}, ErrAuthCodeInvalid
	}
	now := time.Now().UTC().Format(time.RFC3339)
	err = tbl.Update("pk", pk).Range("sk", codeSK).
		Set("consumed_at", now).
		If("attribute_not_exists(consumed_at)").
		Run(ctx)
	if err != nil {
		if dynamo.IsCondCheckFailed(err) {
			return AuthCodeData{}, ErrAuthCodeInvalid
		}
		return AuthCodeData{}, err
	}
	return AuthCodeData{
		CodeChallenge:       row.CodeChallenge,
		CodeChallengeMethod: row.CodeChallengeMethod,
		UserID:              row.UserID,
		ClientID:            row.ClientID,
	}, nil
}
