package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFollows_Follow_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"myband","display_name":"My Band","bio":""}`, session)
	require.NoError(t, err)

	// Second user follows the artist
	session2, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	resp, err := postJSON(client, base, "/users/me/following/myband", `{}`, session2)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp2, err := get(client, base, "/users/me/following", session2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	body, _ := readBody(resp2)
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	var list map[string]any
	require.NoError(t, json.Unmarshal(body, &list))
	handles, _ := list["handles"].([]any)
	require.Len(t, handles, 1)
	require.Equal(t, "myband", handles[0])

	// Artist follower_count is incremented
	resp3, err := get(client, base, "/artists/myband", session2)
	require.NoError(t, err)
	defer resp3.Body.Close()
	require.Equal(t, http.StatusOK, resp3.StatusCode)
	var artist map[string]any
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&artist))
	require.Equal(t, float64(1), artist["follower_count"], "artist should have follower_count 1 after one follow")
}

func TestFollows_Follow_ArtistNotFound(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	resp, err := postJSON(client, base, "/users/me/following/nonexistent", `{}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestFollows_Unfollow_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"unfollowband","display_name":"Band","bio":""}`, session)
	require.NoError(t, err)

	session2, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/users/me/following/unfollowband", `{}`, session2)
	require.NoError(t, err)

	resp, err := deleteReq(client, base, "/users/me/following/unfollowband", session2)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp2, err := get(client, base, "/users/me/following", session2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	var list map[string]any
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&list))
	handles, _ := list["handles"].([]any)
	require.Len(t, handles, 0)

	// Artist follower_count is decremented
	resp3, err := get(client, base, "/artists/unfollowband", session2)
	require.NoError(t, err)
	defer resp3.Body.Close()
	require.Equal(t, http.StatusOK, resp3.StatusCode)
	var artist map[string]any
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&artist))
	require.Equal(t, float64(0), artist["follower_count"], "artist should have follower_count 0 after unfollow")
}

func TestFollows_ListFollowing_Empty(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	resp, err := get(client, base, "/users/me/following", session)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	require.NotNil(t, list["handles"])
	require.Len(t, list["handles"], 0)
}

func TestFollows_ListFollowing_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := get(client, base, "/users/me/following", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
