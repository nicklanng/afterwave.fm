package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

// FeedDoc is the document we store in the feed index (refs only; full content in DynamoDB).
type FeedDoc struct {
	PostID       string `json:"post_id"`
	ArtistHandle string `json:"artist_handle"`
	CreatedAt    string `json:"created_at"`
	BodyExcerpt  string `json:"body_excerpt,omitempty"`
	Explicit     bool   `json:"explicit"`
}

// FeedIndex indexes post refs for the collated "my feed" and search. Full content lives in DynamoDB.
type FeedIndex struct {
	os    *infra.OpenSearch
	index string
}

// NewFeedIndex returns a FeedIndex for the given OpenSearch client and index name.
func NewFeedIndex(os *infra.OpenSearch, indexName string) *FeedIndex {
	if indexName == "" {
		indexName = "afterwave-feed"
	}
	return &FeedIndex{os: os, index: indexName}
}

// IndexPost indexes or updates a post ref. Document ID = artist_handle#post_id for idempotent upsert.
func (f *FeedIndex) IndexPost(ctx context.Context, doc FeedDoc) error {
	if doc.PostID == "" || doc.ArtistHandle == "" {
		return fmt.Errorf("post_id and artist_handle required")
	}
	id := doc.ArtistHandle + "#" + doc.PostID
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	// refresh=wait_for so the document is searchable when the call returns (avoids flaky tests and eventual consistency)
	url := f.os.BaseURL + "/" + f.index + "/_doc/" + id + "?refresh=wait_for"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.os.DoWithRetry(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("opensearch index: %d %s", resp.StatusCode, string(b))
	}
	return nil
}

// DeletePost removes the post ref from the index.
func (f *FeedIndex) DeletePost(ctx context.Context, artistHandle, postID string) error {
	if artistHandle == "" || postID == "" {
		return fmt.Errorf("artist_handle and post_id required")
	}
	id := artistHandle + "#" + postID
	url := f.os.BaseURL + "/" + f.index + "/_doc/" + id
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := f.os.DoWithRetry(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 404 is ok (already deleted)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("opensearch delete: %d %s", resp.StatusCode, string(b))
	}
	return nil
}

// SearchFeedResult is a single hit from the feed search.
type SearchFeedResult struct {
	PostID       string `json:"post_id"`
	ArtistHandle string `json:"artist_handle"`
	CreatedAt    string `json:"created_at"`
	Explicit     bool   `json:"explicit"`
}

// SearchFeed returns post refs from the feed index: filter by artist_handle in handles, sort by created_at desc.
func (f *FeedIndex) SearchFeed(ctx context.Context, artistHandles []string, size, from int) ([]SearchFeedResult, error) {
	if size <= 0 {
		size = 20
	}
	if from < 0 {
		from = 0
	}
	var query map[string]any
	if len(artistHandles) == 0 {
		query = map[string]any{"match_none": struct{}{}}
	} else {
		query = map[string]any{
			"terms": map[string]any{"artist_handle": artistHandles},
		}
	}
	body := map[string]any{
		"query":            query,
		"sort":             []map[string]any{{"created_at": map[string]string{"order": "desc"}}},
		"size":             size,
		"from":             from,
		"track_total_hits": true,
		"_source":          []string{"post_id", "artist_handle", "created_at", "explicit"},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := f.os.BaseURL + "/" + f.index + "/_search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(bodyBytes)), nil }
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.os.DoWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("opensearch search: %d %s", resp.StatusCode, string(b))
	}
	var out struct {
		Hits struct {
			Hits []struct {
				Source SearchFeedResult `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	results := make([]SearchFeedResult, 0, len(out.Hits.Hits))
	for _, h := range out.Hits.Hits {
		results = append(results, h.Source)
	}
	return results, nil
}

// EnsureIndex creates the feed index with mapping if it does not exist. Safe to call at startup.
func (f *FeedIndex) EnsureIndex(ctx context.Context) error {
	url := f.os.BaseURL + "/" + f.index
	// Check exists
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return err
	}
	resp, err := f.os.DoWithRetry(ctx, req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	// Create with mapping
	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"post_id":        map[string]string{"type": "keyword"},
				"artist_handle":  map[string]string{"type": "keyword"},
				"created_at":     map[string]string{"type": "date"},
				"body_excerpt":   map[string]string{"type": "text"},
				"explicit":       map[string]string{"type": "boolean"},
			},
		},
	}
	bodyBytes, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	req2, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req2.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(bodyBytes)), nil }
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := f.os.DoWithRetry(ctx, req2)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK && resp2.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp2.Body)
		return fmt.Errorf("opensearch create index: %d %s", resp2.StatusCode, string(b))
	}
	return nil
}

// DeleteIndex removes the feed index. Used in tests to clear state.
func (f *FeedIndex) DeleteIndex(ctx context.Context) error {
	url := f.os.BaseURL + "/" + f.index
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := f.os.DoWithRetry(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 200 or 404 both ok
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("opensearch delete index: %d %s", resp.StatusCode, string(b))
	}
	return nil
}
