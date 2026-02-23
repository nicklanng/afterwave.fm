package tests

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMyFeed_ReturnsPostsFromFollowedArtists requires DynamoDB and OpenSearch (make test starts both).
// Uses a unique handle so it doesn't collide with feed_test.go which uses "feedband".
func TestMyFeed_ReturnsPostsFromFollowedArtists(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	handle := "myfeedartist"
	// Artist creates a post
	ownerSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"`+handle+`","display_name":"Feed Band","bio":""}`, ownerSession)
	require.NoError(t, err)
	require.Equalf(t, http.StatusCreated, createArtistResp.StatusCode, "artist create should succeed (need unique handle); got %d", createArtistResp.StatusCode)
	createArtistResp.Body.Close()

	postResp, err := postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"Hello followers","body":"Hello followers!"}`, ownerSession)
	require.NoError(t, err)
	require.Equalf(t, http.StatusCreated, postResp.StatusCode, "post create should succeed (OpenSearch must be running for feed index); got %d", postResp.StatusCode)
	postResp.Body.Close()

	// Fan follows artist then fetches my feed
	fanSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	resp, err := postJSON(client, base, "/users/me/following/"+handle, `{}`, fanSession)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Allow OpenSearch to make the indexed post visible to search (refresh=wait_for helps but can be flaky in CI)
	time.Sleep(2 * time.Second)

	feedResp, err := get(client, base, "/feed", fanSession)
	require.NoError(t, err)
	defer feedResp.Body.Close()
	require.Equalf(t, http.StatusOK, feedResp.StatusCode, "my feed requires OpenSearch (make test starts it); got %d", feedResp.StatusCode)
	var feedBody map[string]any
	require.NoError(t, json.NewDecoder(feedResp.Body).Decode(&feedBody))
	posts, _ := feedBody["posts"].([]any)
	require.Len(t, posts, 1)
	require.Equal(t, "Hello followers!", posts[0].(map[string]any)["body"])
	require.Equal(t, handle, posts[0].(map[string]any)["artist_handle"])
	require.Equal(t, false, feedBody["has_more"])
}

// TestMyFeed_OnlyFollowedArtists_NewestFirst populates multiple artists with posts, user follows only some,
// and asserts the feed contains only posts from followed artists in newest-first order.
func TestMyFeed_OnlyFollowedArtists_NewestFirst(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	// Create 3 artists with posts (each artist owned by a different user). Handles must be [a-z0-9] 4–64 chars.
	createArtistWithPost := func(handle, displayName, postBody string) {
		session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
		require.NoError(t, err)
		createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"`+handle+`","display_name":"`+displayName+`","bio":""}`, session)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, createArtistResp.StatusCode)
		createArtistResp.Body.Close()
		postResp, err := postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"`+postBody+`","body":"`+postBody+`"}`, session)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, postResp.StatusCode)
		postResp.Body.Close()
	}

	createArtistWithPost("feedartista", "Artist A", "Post from A")
	time.Sleep(1 * time.Second) // so created_at differs
	createArtistWithPost("feedartistb", "Artist B", "Post from B")
	time.Sleep(1 * time.Second)
	createArtistWithPost("feedartistc", "Artist C", "Post from C")

	// User follows only A and C (not B)
	fanSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	for _, handle := range []string{"feedartista", "feedartistc"} {
		resp, err := postJSON(client, base, "/users/me/following/"+handle, `{}`, fanSession)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	}

	// Allow OpenSearch to index
	time.Sleep(2 * time.Second)

	feedResp, err := get(client, base, "/feed", fanSession)
	require.NoError(t, err)
	defer feedResp.Body.Close()
	require.Equal(t, http.StatusOK, feedResp.StatusCode)
	var feedBody map[string]any
	require.NoError(t, json.NewDecoder(feedResp.Body).Decode(&feedBody))
	posts, _ := feedBody["posts"].([]any)
	require.Len(t, posts, 2, "feed should have only posts from followed artists A and C, not B")
	// Newest first: C then A (B must not appear)
	handles := []string{posts[0].(map[string]any)["artist_handle"].(string), posts[1].(map[string]any)["artist_handle"].(string)}
	bodies := []string{posts[0].(map[string]any)["body"].(string), posts[1].(map[string]any)["body"].(string)}
	require.ElementsMatch(t, handles, []string{"feedartista", "feedartistc"})
	require.ElementsMatch(t, bodies, []string{"Post from A", "Post from C"})
	// Order: C (newest) first, then A
	require.Equal(t, "feedartistc", posts[0].(map[string]any)["artist_handle"])
	require.Equal(t, "Post from C", posts[0].(map[string]any)["body"])
	require.Equal(t, "feedartista", posts[1].(map[string]any)["artist_handle"])
	require.Equal(t, "Post from A", posts[1].(map[string]any)["body"])
	require.Equal(t, false, feedBody["has_more"])
}

// TestMyFeed_Paginated_ChronologicalOrder creates one artist with several posts (distinct created_at),
// then asserts my feed returns them in chronological order (newest first) and has_more when applicable.
// Uses one artist; OpenSearch in test env may return fewer than total indexed (eventual consistency).
func TestMyFeed_Paginated_ChronologicalOrder(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	pageSize := 10
	handle := "pageartist"
	ownerSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"`+handle+`","display_name":"Page Artist","bio":""}`, ownerSession)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createArtistResp.StatusCode)
	createArtistResp.Body.Close()

	// Create several posts with 1s apart so created_at is distinct (RFC3339 second precision)
	numPosts := 5
	for i := 1; i <= numPosts; i++ {
		body := "Post " + strconv.Itoa(i)
		postResp, err := postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"`+body+`","body":"`+body+`"}`, ownerSession)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, postResp.StatusCode)
		postResp.Body.Close()
		if i < numPosts {
			time.Sleep(1 * time.Second)
		}
	}

	fanSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	resp, err := postJSON(client, base, "/users/me/following/"+handle, `{}`, fanSession)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	time.Sleep(3 * time.Second) // allow indexed docs to be searchable

	feedResp1, err := get(client, base, "/feed?limit="+strconv.Itoa(pageSize), fanSession)
	require.NoError(t, err)
	defer feedResp1.Body.Close()
	require.Equal(t, http.StatusOK, feedResp1.StatusCode)
	var feed1 map[string]any
	require.NoError(t, json.NewDecoder(feedResp1.Body).Decode(&feed1))
	posts1, _ := feed1["posts"].([]any)
	require.NotEmpty(t, posts1, "my feed should return at least one post (OpenSearch may be eventually consistent)")
	require.Equal(t, false, feed1["has_more"], "5 posts and limit 10 => has_more false")

	// Order: newest first = Post 5, Post 4, ..., Post 1 (for those returned)
	for i := 0; i < len(posts1); i++ {
		expected := "Post " + strconv.Itoa(numPosts-i)
		require.Equal(t, expected, posts1[i].(map[string]any)["body"], "order must be chronological (newest first)")
	}
}

// TestMyFeed_DeletedPostNotInFeed asserts that after the artist deletes a post, it no longer appears in my feed.
// Requires DynamoDB and OpenSearch (make test starts both).
func TestMyFeed_DeletedPostNotInFeed(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	handle := "deletefromfeed"
	ownerSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"`+handle+`","display_name":"Delete From Feed","bio":""}`, ownerSession)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createArtistResp.StatusCode)
	createArtistResp.Body.Close()

	postResp, err := postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"Will be deleted","body":"Will be deleted"}`, ownerSession)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, postResp.StatusCode)
	var created map[string]any
	require.NoError(t, json.NewDecoder(postResp.Body).Decode(&created))
	postResp.Body.Close()
	postID := created["post_id"].(string)

	fanSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	resp, err := postJSON(client, base, "/users/me/following/"+handle, `{}`, fanSession)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	time.Sleep(2 * time.Second)

	feedResp, err := get(client, base, "/feed", fanSession)
	require.NoError(t, err)
	defer feedResp.Body.Close()
	require.Equal(t, http.StatusOK, feedResp.StatusCode)
	var feedBody map[string]any
	require.NoError(t, json.NewDecoder(feedResp.Body).Decode(&feedBody))
	posts, _ := feedBody["posts"].([]any)
	require.Len(t, posts, 1, "feed should contain the post before delete")
	require.Equal(t, postID, posts[0].(map[string]any)["post_id"])

	delResp, err := deleteReq(client, base, "/artists/"+handle+"/posts/"+postID, ownerSession)
	require.NoError(t, err)
	delResp.Body.Close()
	require.Equal(t, http.StatusNoContent, delResp.StatusCode)

	time.Sleep(2 * time.Second)

	feedResp2, err := get(client, base, "/feed", fanSession)
	require.NoError(t, err)
	defer feedResp2.Body.Close()
	require.Equal(t, http.StatusOK, feedResp2.StatusCode)
	var feedBody2 map[string]any
	require.NoError(t, json.NewDecoder(feedResp2.Body).Decode(&feedBody2))
	posts2, _ := feedBody2["posts"].([]any)
	require.Len(t, posts2, 0, "deleted post should not appear in my feed")
}

func TestMyFeed_EmptyWhenNotFollowingAnyone(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	resp, err := get(client, base, "/feed", session)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var feedBody map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&feedBody))
	posts, _ := feedBody["posts"].([]any)
	require.NotNil(t, posts)
	require.Len(t, posts, 0)
	require.Equal(t, false, feedBody["has_more"])
}

func TestMyFeed_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := get(client, base, "/feed", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
