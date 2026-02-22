package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFeed_CreatePost_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	// Unique handle so we don't collide with TestArtists_Create_Success which uses "myband"
	handle := "createpostband"
	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"`+handle+`","display_name":"My Band","bio":""}`, session)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createArtistResp.StatusCode)
	createArtistResp.Body.Close()

	resp, err := postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"Hello world","body":"Hello world!","explicit":false}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body: %s", string(body))

	var post map[string]any
	require.NoError(t, json.Unmarshal(body, &post))
	require.NotEmpty(t, post["post_id"])
	require.Equal(t, handle, post["artist_handle"])
	require.Equal(t, "Hello world", post["title"])
	require.Equal(t, "Hello world!", post["body"])
	require.Equal(t, false, post["explicit"])
	require.NotEmpty(t, post["created_at"])
	require.NotEmpty(t, post["created_by_user_id"])
}

func TestFeed_CreatePost_WithImageAndYouTube(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"band2","display_name":"Band Two","bio":""}`, session)
	require.NoError(t, err)

	resp, err := postJSON(client, base, "/artists/band2/posts", `{"title":"Check this out","body":"Check this out","image_url":"https://example.com/img.png","youtube_url":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","explicit":false}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body: %s", string(body))

	var post map[string]any
	require.NoError(t, json.Unmarshal(body, &post))
	require.Equal(t, "https://example.com/img.png", post["image_url"])
	require.Equal(t, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", post["youtube_url"])
}

func TestFeed_CreatePost_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"myband","display_name":"My Band"}`, session)
	require.NoError(t, err)

	resp, err := postJSON(client, base, "/artists/myband/posts", `{"title":"Hello","body":"Hello"}`, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestFeed_CreatePost_ArtistNotFound(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	resp, err := postJSON(client, base, "/artists/nonexistent/posts", `{"title":"Hello","body":"Hello"}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestFeed_CreatePost_SlugConflict(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"slugband","display_name":"Slug Band","bio":""}`, session)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createArtistResp.StatusCode)
	createArtistResp.Body.Close()

	resp1, err := postJSON(client, base, "/artists/slugband/posts", `{"title":"Same Title","body":"First"}`, session)
	require.NoError(t, err)
	resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	resp2, err := postJSON(client, base, "/artists/slugband/posts", `{"title":"Same Title","body":"Second"}`, session)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusConflict, resp2.StatusCode)
}

func TestFeed_CreatePost_Forbidden(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	ownerSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"notyours","display_name":"Mine"}`, ownerSession)
	require.NoError(t, err)

	otherSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	resp, err := postJSON(client, base, "/artists/notyours/posts", `{"body":"Hacked"}`, otherSession)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestFeed_ListPosts_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	// Unique handle so we don't collide with other tests that use "feedband"
	handle := "listpostsband"
	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"`+handle+`","display_name":"Feed Band","bio":""}`, session)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createArtistResp.StatusCode)
	createArtistResp.Body.Close()

	// Empty list first
	resp, err := get(client, base, "/artists/"+handle+"/posts", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list map[string]any
	require.NoError(t, json.Unmarshal(body, &list))
	require.NotNil(t, list["posts"])
	require.Len(t, list["posts"], 0)
	require.Equal(t, false, list["has_more"])

	// Create two posts (sleep 1s so created_at differs and list order is newest-first)
	_, err = postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"First post","body":"First post"}`, session)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	_, err = postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"Second post","body":"Second post"}`, session)
	require.NoError(t, err)

	resp2, err := get(client, base, "/artists/"+handle+"/posts", "")
	require.NoError(t, err)
	defer resp2.Body.Close()
	body2, _ := readBody(resp2)
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	require.NoError(t, json.Unmarshal(body2, &list))
	posts, _ := list["posts"].([]any)
	require.Len(t, posts, 2)
	require.Equal(t, false, list["has_more"], "2 posts and default limit 10 => has_more false")
	// Newest first
	require.Equal(t, "Second post", posts[0].(map[string]any)["body"])
	require.Equal(t, "First post", posts[1].(map[string]any)["body"])
	require.Equal(t, handle, posts[0].(map[string]any)["artist_handle"])
}

func TestFeed_ListPosts_Public(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"publicfeed","display_name":"Public","bio":""}`, session)
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists/publicfeed/posts", `{"title":"Public post","body":"Public post"}`, session)
	require.NoError(t, err)

	resp, err := get(client, base, "/artists/publicfeed/posts", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	posts, _ := list["posts"].([]any)
	require.Len(t, posts, 1)
	require.Equal(t, false, list["has_more"])
	require.Equal(t, "Public post", posts[0].(map[string]any)["body"])
}

func TestFeed_ListPosts_ArtistNotFound(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := get(client, base, "/artists/nonexistent/posts", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestFeed_GetPost_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"getband","display_name":"Get Band","bio":""}`, session)
	require.NoError(t, err)
	createResp, err := postJSON(client, base, "/artists/getband/posts", `{"title":"Single post","body":"Single post","explicit":true}`, session)
	require.NoError(t, err)
	defer createResp.Body.Close()
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	postID := created["post_id"].(string)
	require.NotEmpty(t, postID)

	resp, err := get(client, base, "/artists/getband/posts/"+postID, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var post map[string]any
	require.NoError(t, json.Unmarshal(body, &post))
	require.Equal(t, postID, post["post_id"])
	require.Equal(t, "Single post", post["body"])
	require.Equal(t, true, post["explicit"])
}

func TestFeed_GetPost_NotFound(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"getband","display_name":"Get Band"}`, session)
	require.NoError(t, err)

	resp, err := get(client, base, "/artists/getband/posts/no-such-slug", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestFeed_UpdatePost_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"updateband","display_name":"Update Band","bio":""}`, session)
	require.NoError(t, err)
	createResp, err := postJSON(client, base, "/artists/updateband/posts", `{"title":"Original","body":"Original"}`, session)
	require.NoError(t, err)
	defer createResp.Body.Close()
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	postID := created["post_id"].(string)

	resp, err := patchJSON(client, base, "/artists/updateband/posts/"+postID, `{"body":"Updated body"}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var post map[string]any
	require.NoError(t, json.Unmarshal(body, &post))
	require.Equal(t, "Updated body", post["body"])
	require.NotEmpty(t, post["updated_at"])

	resp2, err := get(client, base, "/artists/updateband/posts/"+postID, "")
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&post))
	require.Equal(t, "Updated body", post["body"])
}

func TestFeed_UpdatePost_Forbidden(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	// Use unique handle so we own the artist (other tests use "notyours")
	handle := "updatepostforbidden"
	ownerSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"`+handle+`","display_name":"Mine"}`, ownerSession)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createArtistResp.StatusCode)
	createArtistResp.Body.Close()

	createResp, err := postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"Original","body":"Original"}`, ownerSession)
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	postID := created["post_id"].(string)

	otherSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	resp, err := patchJSON(client, base, "/artists/"+handle+"/posts/"+postID, `{"body":"Hacked"}`, otherSession)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestFeed_DeletePost_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"deleteband","display_name":"Delete Band","bio":""}`, session)
	require.NoError(t, err)
	createResp, err := postJSON(client, base, "/artists/deleteband/posts", `{"title":"To be deleted","body":"To be deleted"}`, session)
	require.NoError(t, err)
	defer createResp.Body.Close()
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	postID := created["post_id"].(string)

	resp, err := deleteReq(client, base, "/artists/deleteband/posts/"+postID, session)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp2, err := get(client, base, "/artists/deleteband/posts/"+postID, "")
	require.NoError(t, err)
	resp2.Body.Close()
	require.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestFeed_DeletePost_Forbidden(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	// Use unique handle so we own the artist (other tests use "notyours")
	handle := "deletepostforbidden"
	ownerSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	createArtistResp, err := postJSON(client, base, "/artists", `{"handle":"`+handle+`","display_name":"Mine"}`, ownerSession)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createArtistResp.StatusCode)
	createArtistResp.Body.Close()

	createResp, err := postJSON(client, base, "/artists/"+handle+"/posts", `{"title":"Original","body":"Original"}`, ownerSession)
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	postID := created["post_id"].(string)

	otherSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	resp, err := deleteReq(client, base, "/artists/"+handle+"/posts/"+postID, otherSession)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
