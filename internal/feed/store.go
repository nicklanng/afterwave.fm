package feed

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
	PK              string `dynamo:"pk"`
	SK              string `dynamo:"sk"`
	PostID          string `dynamo:"post_id"` // slug (unique per artist)
	ArtistHandle    string `dynamo:"artist_handle"`
	Title           string `dynamo:"title"`
	Body            string `dynamo:"body"`
	ImageURL        string `dynamo:"image_url,omitempty"`
	YouTubeURL      string `dynamo:"youtube_url,omitempty"`
	Explicit        bool   `dynamo:"explicit"`
	CreatedAt       string `dynamo:"created_at"`
	UpdatedAt       string `dynamo:"updated_at,omitempty"`
	CreatedByUserID string `dynamo:"created_by_user_id"`
}

type postByTimeRow struct {
	PK        string `dynamo:"pk" dynamodbav:"pk"`
	SK        string `dynamo:"sk" dynamodbav:"sk"`
	PostID    string `dynamo:"post_id" dynamodbav:"post_id"`
	CreatedAt string `dynamo:"created_at" dynamodbav:"created_at"`
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
		Put(s.tbl().Put(mainRow)).
		Put(s.tbl().Put(byTimeRow)).
		Run(ctx)
}

// Get returns the post by handle and post ID, or nil if not found.
func (s *Store) Get(ctx context.Context, handle, postID string) (*postRow, error) {
	handle = normalizeHandle(handle)
	var row postRow
	err := s.tbl().Get("pk", artistPK(handle)).Range("sk", dynamo.Equal, postSK(postID)).One(ctx, &row)
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// listByTimeCursor is the opaque cursor format: base64(json({p: pk, s: sk})).
func encodeListByTimeCursor(pk, sk string) string {
	b, _ := json.Marshal(struct {
		P string `json:"p"`
		S string `json:"s"`
	}{pk, sk})
	return base64.StdEncoding.EncodeToString(b)
}

func decodeListByTimeCursor(cursor string) (pk, sk string, err error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", "", err
	}
	var v struct {
		P string `json:"p"`
		S string `json:"s"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return "", "", err
	}
	return v.P, v.S, nil
}

// ListByTimePage returns post refs from the BYTIME index, newest first, using cursor-based pagination.
// cursor is optional (empty for first page). nextCursor is non-empty when more results exist.
func (s *Store) ListByTimePage(ctx context.Context, handle string, limit int, cursor string) ([]postByTimeRow, string, error) {
	handle = normalizeHandle(handle)
	pk := artistPK(handle)
	var startKey map[string]types.AttributeValue
	if cursor != "" {
		decodedPk, decodedSk, err := decodeListByTimeCursor(cursor)
		if err != nil {
			return nil, "", err
		}
		startKey = map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: decodedPk},
			"sk": &types.AttributeValueMemberS{Value: decodedSk},
		}
	}
	req := &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: pk},
			":prefix": &types.AttributeValueMemberS{Value: postByTimePrefix},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(int32(limit + 1)),
		ExclusiveStartKey: startKey,
	}
	out, err := s.db.Client().Query(ctx, req)
	if err != nil {
		return nil, "", err
	}
	rows := make([]postByTimeRow, 0, len(out.Items))
	for _, item := range out.Items {
		var row postByTimeRow
		if err := attributevalue.UnmarshalMap(item, &row); err != nil {
			return nil, "", err
		}
		rows = append(rows, row)
	}
	var next string
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
		last := rows[limit-1]
		next = encodeListByTimeCursor(last.PK, last.SK)
	}
	return rows, next, nil
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
