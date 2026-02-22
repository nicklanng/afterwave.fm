package feed

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sopatech/afterwave.fm/internal/artists"
	"github.com/sopatech/afterwave.fm/internal/search"
)

var (
	ErrArtistNotFound = errors.New("artist not found")
	ErrPostNotFound   = errors.New("post not found")
	ErrForbidden      = errors.New("not the owner of this artist page")
	ErrSlugConflict   = errors.New("a post with this title already exists for this artist")
)

// ArtistResolver is used to resolve artist by handle (for ownership check). Implemented by artists.Service.
type ArtistResolver interface {
	GetByHandle(ctx context.Context, handle string) (*artists.Artist, error)
}

// FollowingLister returns the list of artist handles a user follows. Implemented by follows.Service.
type FollowingLister interface {
	ListFollowing(ctx context.Context, userID string) ([]string, error)
}

type Service interface {
	CreatePost(ctx context.Context, handle string, title, body, imageURL, youtubeURL string, explicit bool, actorUserID string) (*Post, error)
	ListPosts(ctx context.Context, handle string, limit, from int) ([]Post, bool, error)
	GetPost(ctx context.Context, handle, postID string) (*Post, error)
	UpdatePost(ctx context.Context, handle, postID string, body, imageURL, youtubeURL *string, explicit *bool, actorUserID string) (*Post, error)
	DeletePost(ctx context.Context, handle, postID string, actorUserID string) error
	MyFeed(ctx context.Context, userID string, limit, from int) ([]Post, bool, error)
}

type Post struct {
	PostID          string `json:"post_id"` // slug (unique per artist, derived from title)
	ArtistHandle    string `json:"artist_handle"`
	Title           string `json:"title"`
	Body            string `json:"body"`
	ImageURL        string `json:"image_url,omitempty"`
	YouTubeURL      string `json:"youtube_url,omitempty"`
	Explicit        bool   `json:"explicit"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	CreatedByUserID string `json:"created_by_user_id"`
}

type service struct {
	store      *Store
	artist     ArtistResolver
	indexer    FeedIndexer
	following  FollowingLister
	feedIndex  *search.FeedIndex
}

// FeedIndexer indexes post refs to OpenSearch (optional; when nil, indexing is skipped).
type FeedIndexer interface {
	IndexPost(ctx context.Context, doc search.FeedDoc) error
	DeletePost(ctx context.Context, artistHandle, postID string) error
}

func NewService(store *Store, artist ArtistResolver) Service {
	return &service{store: store, artist: artist}
}

// NewServiceWithSearch returns a Service that indexes to OpenSearch on create/update/delete and supports MyFeed.
func NewServiceWithSearch(store *Store, artist ArtistResolver, indexer FeedIndexer, following FollowingLister, feedIndex *search.FeedIndex) Service {
	return &service{store: store, artist: artist, indexer: indexer, following: following, feedIndex: feedIndex}
}

func (s *service) ensureOwner(ctx context.Context, handle string, actorUserID string) error {
	artist, err := s.artist.GetByHandle(ctx, handle)
	if err != nil || artist == nil {
		return ErrArtistNotFound
	}
	if artist.OwnerUserID != actorUserID {
		return ErrForbidden
	}
	return nil
}

func (s *service) CreatePost(ctx context.Context, handle string, title, body, imageURL, youtubeURL string, explicit bool, actorUserID string) (*Post, error) {
	handle = normalizeHandle(handle)
	if err := s.ensureOwner(ctx, handle, actorUserID); err != nil {
		return nil, err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	slug := Slugify(title)
	if slug == "" {
		return nil, errors.New("title must contain at least one letter or number")
	}
	existing, _ := s.store.Get(ctx, handle, slug)
	if existing != nil {
		return nil, ErrSlugConflict
	}
	body = strings.TrimSpace(body)
	createdAt := time.Now().UTC().Format(time.RFC3339)
	row := postRow{
		PostID:          slug,
		ArtistHandle:    handle,
		Title:           title,
		Body:            body,
		ImageURL:        strings.TrimSpace(imageURL),
		YouTubeURL:      strings.TrimSpace(youtubeURL),
		Explicit:        explicit,
		CreatedAt:       createdAt,
		CreatedByUserID: actorUserID,
	}
	if err := s.store.Create(ctx, handle, row); err != nil {
		return nil, err
	}
	if s.indexer != nil {
		if err := s.indexer.IndexPost(ctx, search.FeedDoc{
			PostID:       row.PostID,
			ArtistHandle: handle,
			CreatedAt:    row.CreatedAt,
			BodyExcerpt:  truncateBody(row.Body, 200),
			Explicit:     row.Explicit,
		}); err != nil {
			return nil, err
		}
	}
	return rowToPost(&row), nil
}

func (s *service) ListPosts(ctx context.Context, handle string, limit, from int) ([]Post, bool, error) {
	handle = normalizeHandle(handle)
	artist, err := s.artist.GetByHandle(ctx, handle)
	if err != nil || artist == nil {
		return nil, false, ErrArtistNotFound
	}
	_ = artist
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if from < 0 {
		from = 0
	}
	// Request from+limit+1 to detect if there are more results
	byTimeRows, err := s.store.ListByTime(ctx, handle, from+limit+1)
	if err != nil {
		return nil, false, err
	}
	if len(byTimeRows) == 0 {
		return nil, false, nil
	}
	hasMore := len(byTimeRows) > from+limit
	end := from + limit
	if end > len(byTimeRows) {
		end = len(byTimeRows)
	}
	pageRows := byTimeRows[from:end]
	postIDs := make([]string, len(pageRows))
	for i := range pageRows {
		postIDs[i] = pageRows[i].PostID
	}
	rows, err := s.store.BatchGetPosts(ctx, handle, postIDs)
	if err != nil {
		return nil, false, err
	}
	out := make([]Post, len(rows))
	for i := range rows {
		out[i] = *rowToPost(rows[i])
	}
	return out, hasMore, nil
}

func (s *service) GetPost(ctx context.Context, handle, postID string) (*Post, error) {
	handle = normalizeHandle(handle)
	row, err := s.store.Get(ctx, handle, postID)
	if err != nil || row == nil {
		return nil, ErrPostNotFound
	}
	return rowToPost(row), nil
}

func (s *service) UpdatePost(ctx context.Context, handle, postID string, body, imageURL, youtubeURL *string, explicit *bool, actorUserID string) (*Post, error) {
	handle = normalizeHandle(handle)
	if err := s.ensureOwner(ctx, handle, actorUserID); err != nil {
		return nil, err
	}
	row, err := s.store.Get(ctx, handle, postID)
	if err != nil || row == nil {
		return nil, ErrPostNotFound
	}
	resolvedBody := row.Body
	if body != nil {
		resolvedBody = strings.TrimSpace(*body)
	}
	resolvedImageURL := row.ImageURL
	if imageURL != nil {
		resolvedImageURL = strings.TrimSpace(*imageURL)
	}
	resolvedYouTubeURL := row.YouTubeURL
	if youtubeURL != nil {
		resolvedYouTubeURL = strings.TrimSpace(*youtubeURL)
	}
	resolvedExplicit := row.Explicit
	if explicit != nil {
		resolvedExplicit = *explicit
	}
	updatedAt := time.Now().UTC().Format(time.RFC3339)
	if err := s.store.Update(ctx, handle, postID, resolvedBody, resolvedImageURL, resolvedYouTubeURL, resolvedExplicit, updatedAt); err != nil {
		return nil, err
	}
	if s.indexer != nil {
		_ = s.indexer.IndexPost(ctx, search.FeedDoc{
			PostID:       postID,
			ArtistHandle: handle,
			CreatedAt:    row.CreatedAt,
			BodyExcerpt:  truncateBody(resolvedBody, 200),
			Explicit:     resolvedExplicit,
		})
	}
	return &Post{
		PostID:          postID,
		ArtistHandle:    handle,
		Title:           row.Title,
		Body:            resolvedBody,
		ImageURL:        resolvedImageURL,
		YouTubeURL:      resolvedYouTubeURL,
		Explicit:        resolvedExplicit,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       updatedAt,
		CreatedByUserID: row.CreatedByUserID,
	}, nil
}

func (s *service) DeletePost(ctx context.Context, handle, postID string, actorUserID string) error {
	handle = normalizeHandle(handle)
	if err := s.ensureOwner(ctx, handle, actorUserID); err != nil {
		return err
	}
	row, err := s.store.Get(ctx, handle, postID)
	if err != nil || row == nil {
		return ErrPostNotFound
	}
	if err := s.store.Delete(ctx, handle, postID); err != nil {
		return err
	}
	if s.indexer != nil {
		_ = s.indexer.DeletePost(ctx, handle, postID)
	}
	return nil
}

func truncateBody(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// MyFeed returns the collated feed for the user (posts from artists they follow), hydrated from DynamoDB.
// It returns posts, hasMore (true if there are more results after this page), and error.
func (s *service) MyFeed(ctx context.Context, userID string, limit, from int) ([]Post, bool, error) {
	if s.following == nil || s.feedIndex == nil {
		return nil, false, nil
	}
	handles, err := s.following.ListFollowing(ctx, userID)
	if err != nil || len(handles) == 0 {
		return nil, false, err
	}
	// Request limit+1 to detect if there are more results
	refs, err := s.feedIndex.SearchFeed(ctx, handles, limit+1, from)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(refs) > limit
	if hasMore {
		refs = refs[:limit]
	}
	if len(refs) == 0 {
		return nil, false, nil
	}
	// Group refs by artist_handle for batch fetch
	byHandle := make(map[string][]string)
	for _, r := range refs {
		byHandle[r.ArtistHandle] = append(byHandle[r.ArtistHandle], r.PostID)
	}
	// Fetch full posts from DynamoDB per artist
	postMap := make(map[string]map[string]*Post) // handle -> postID -> Post
	for handle, postIDs := range byHandle {
		rows, err := s.store.BatchGetPosts(ctx, handle, postIDs)
		if err != nil {
			return nil, false, err
		}
		if postMap[handle] == nil {
			postMap[handle] = make(map[string]*Post)
		}
		for _, row := range rows {
			p := rowToPost(row)
			postMap[handle][p.PostID] = p
		}
	}
	// Assemble in refs order
	out := make([]Post, 0, len(refs))
	for _, r := range refs {
		if p := postMap[r.ArtistHandle][r.PostID]; p != nil {
			out = append(out, *p)
		}
	}
	return out, hasMore, nil
}

func rowToPost(r *postRow) *Post {
	if r == nil {
		return nil
	}
	return &Post{
		PostID:          r.PostID,
		ArtistHandle:    r.ArtistHandle,
		Title:           r.Title,
		Body:            r.Body,
		ImageURL:        r.ImageURL,
		YouTubeURL:      r.YouTubeURL,
		Explicit:        r.Explicit,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		CreatedByUserID: r.CreatedByUserID,
	}
}
