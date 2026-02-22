package feed

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/guregu/dynamo/v2"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// Slugify converts a title to a URL-safe slug (Hugo-style): lowercased, spaces and non-alphanumeric replaced with single hyphen, trimmed.
func Slugify(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	title = strings.ToLower(title)
	var b strings.Builder
	for _, r := range title {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-':
			if b.Len() > 0 && lastRune(b.String()) != '-' {
				b.WriteRune('-')
			}
		default:
			// skip other runes (special chars)
		}
	}
	return strings.Trim(b.String(), "-")
}

func lastRune(s string) rune {
	r, _ := utf8.DecodeLastRuneInString(s)
	return r
}

// Posts live under the same table as artists. Two-row pattern (no GSI):
// - Main row: PK = ARTISTS#<handle>, SK = POST#<post_id> — full post data.
// - Index row: PK = ARTISTS#<handle>, SK = POST#BYTIME#<created_at>#<post_id> — for List by time (desc).

const (
	artistPKPrefix   = "ARTISTS#"
	postSKPrefix     = "POST#"
	postByTimePrefix = "POST#BYTIME#"
)

func artistPK(handle string) string {
	return artistPKPrefix + normalizeHandle(handle)
}

func postSK(postID string) string {
	return postSKPrefix + postID
}

func postByTimeSK(createdAt, postID string) string {
	return postByTimePrefix + createdAt + "#" + postID
}

type postRow struct {
	PK               string `dynamodbav:"pk" dynamo:"pk"`
	SK               string `dynamodbav:"sk" dynamo:"sk"`
	PostID           string `dynamodbav:"post_id" dynamo:"post_id"` // slug (unique per artist)
	ArtistHandle     string `dynamodbav:"artist_handle" dynamo:"artist_handle"`
	Title            string `dynamodbav:"title" dynamo:"title"`
	Body             string `dynamodbav:"body" dynamo:"body"`
	ImageURL         string `dynamodbav:"image_url,omitempty" dynamo:"image_url,omitempty"`
	YouTubeURL       string `dynamodbav:"youtube_url,omitempty" dynamo:"youtube_url,omitempty"`
	Explicit         bool   `dynamodbav:"explicit" dynamo:"explicit"`
	CreatedAt        string `dynamodbav:"created_at" dynamo:"created_at"`
	UpdatedAt        string `dynamodbav:"updated_at,omitempty" dynamo:"updated_at,omitempty"`
	CreatedByUserID  string `dynamodbav:"created_by_user_id" dynamo:"created_by_user_id"`
}

type postByTimeRow struct {
	PK              string `dynamodbav:"pk"`
	SK              string `dynamodbav:"sk"`
	PostID          string `dynamodbav:"post_id"`
	CreatedAt       string `dynamodbav:"created_at"`
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

// Create writes the main post row and the BYTIME index row in one transaction.
func (s *Store) Create(ctx context.Context, handle string, row postRow) error {
	handle = normalizeHandle(handle)
	pk := artistPK(handle)
	mainRow := postRow{
		PK:              pk,
		SK:              postSK(row.PostID),
		PostID:          row.PostID,
		ArtistHandle:    handle,
		Title:           row.Title,
		Body:            row.Body,
		ImageURL:        row.ImageURL,
		YouTubeURL:      row.YouTubeURL,
		Explicit:        row.Explicit,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		CreatedByUserID: row.CreatedByUserID,
	}
	byTimeRow := postByTimeRow{
		PK:        pk,
		SK:        postByTimeSK(row.CreatedAt, row.PostID),
		PostID:    row.PostID,
		CreatedAt: row.CreatedAt,
	}
	return s.db.WriteTx().
		Put(s.tbl().Put(dynamo.AWSEncoding(mainRow))).
		Put(s.tbl().Put(dynamo.AWSEncoding(byTimeRow))).
		Run(ctx)
}

// Get returns the post by handle and post ID, or nil if not found.
func (s *Store) Get(ctx context.Context, handle, postID string) (*postRow, error) {
	handle = normalizeHandle(handle)
	var row postRow
	err := s.tbl().Get("pk", artistPK(handle)).Range("sk", dynamo.Equal, postSK(postID)).One(ctx, dynamo.AWSEncoding(&row))
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListByTime returns post refs (post_id, created_at) from the BYTIME index, newest first.
func (s *Store) ListByTime(ctx context.Context, handle string, limit int) ([]postByTimeRow, error) {
	handle = normalizeHandle(handle)
	pk := artistPK(handle)
	var out []postByTimeRow
	iter := s.tbl().Get("pk", pk).Range("sk", dynamo.BeginsWith, postByTimePrefix).Order(dynamo.Descending).Limit(limit).Iter()
	var row postByTimeRow
	for iter.Next(ctx, dynamo.AWSEncoding(&row)) {
		out = append(out, row)
	}
	return out, iter.Err()
}

// Update updates the main post row (body, image_url, youtube_url, explicit, updated_at).
func (s *Store) Update(ctx context.Context, handle, postID, body, imageURL, youtubeURL string, explicit bool, updatedAt string) error {
	handle = normalizeHandle(handle)
	return s.tbl().Update("pk", artistPK(handle)).Range("sk", postSK(postID)).
		Set("body", body).
		Set("image_url", imageURL).
		Set("youtube_url", youtubeURL).
		Set("explicit", explicit).
		Set("updated_at", updatedAt).
		Run(ctx)
}

// Delete removes the main post row and the BYTIME index row.
func (s *Store) Delete(ctx context.Context, handle, postID string) error {
	main, err := s.Get(ctx, handle, postID)
	if err != nil || main == nil {
		return err
	}
	return s.db.WriteTx().
		Delete(s.tbl().Delete("pk", artistPK(handle)).Range("sk", postSK(postID))).
		Delete(s.tbl().Delete("pk", artistPK(handle)).Range("sk", postByTimeSK(main.CreatedAt, postID))).
		Run(ctx)
}

func normalizeHandle(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// BatchGetPosts returns full post rows for the given post IDs (same handle), in the same order as postIDs.
// Uses DynamoDB BatchGetItem (one or more round-trips of up to 100 items each).
func (s *Store) BatchGetPosts(ctx context.Context, handle string, postIDs []string) ([]*postRow, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}
	handle = normalizeHandle(handle)
	pk := artistPK(handle)
	keys := make([]dynamo.Keyed, len(postIDs))
	for i, id := range postIDs {
		keys[i] = dynamo.Keys{pk, postSK(id)}
	}
	var rows []postRow
	if err := s.tbl().Batch("pk", "sk").Get(keys...).All(ctx, &rows); err != nil {
		return nil, err
	}
	// Preserve order of postIDs; point into rows slice (same order as returned by BatchGet is undefined)
	orderMap := make(map[string]*postRow, len(rows))
	for i := range rows {
		orderMap[rows[i].PostID] = &rows[i]
	}
	ordered := make([]*postRow, 0, len(postIDs))
	for _, id := range postIDs {
		if r := orderMap[id]; r != nil {
			ordered = append(ordered, r)
		}
	}
	return ordered, nil
}
