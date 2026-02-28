package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArtists_Create_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, session)

	resp, err := postJSON(client, base, "/artists", `{"handle":"myband","display_name":"My Band","bio":"We are a band."}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body: %s", string(body))

	var artist map[string]any
	require.NoError(t, json.Unmarshal(body, &artist))
	require.Equal(t, "myband", artist["handle"])
	require.Equal(t, "My Band", artist["display_name"])
	require.Equal(t, "We are a band.", artist["bio"])
	require.NotEmpty(t, artist["owner_user_id"])
	require.NotEmpty(t, artist["created_at"])
}

func TestArtists_Create_DuplicateHandle_Conflict(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	createBody := `{"handle":"taken","display_name":"First"}`
	resp1, err := postJSON(client, base, "/artists", createBody, session)
	require.NoError(t, err)
	resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	// Second user tries same handle
	session2, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	resp2, err := postJSON(client, base, "/artists", createBody, session2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusConflict, resp2.StatusCode)
}

func TestArtists_Create_InvalidHandle_BadRequest(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	resp, err := postJSON(client, base, "/artists", `{"handle":"Bad Handle!","display_name":"X"}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestArtists_Create_HandleTooShort_BadRequest(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	// Min 4 chars so we can reserve 3-letter subdomains (www, tui, api)
	resp, err := postJSON(client, base, "/artists", `{"handle":"abc","display_name":"Too Short"}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestArtists_Create_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := postJSON(client, base, "/artists", `{"handle":"myband","display_name":"My Band"}`, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestArtists_ListMine_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	// Empty list first
	resp, err := get(client, base, "/artists/me", session)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list map[string]any
	require.NoError(t, json.Unmarshal(body, &list))
	require.NotNil(t, list["artists"])
	require.Len(t, list["artists"], 0)

	// Create one artist (unique handle so we don't collide with TestArtists_Create_Success which uses "myband")
	_, err = postJSON(client, base, "/artists", `{"handle":"listmineband","display_name":"My Band","bio":""}`, session)
	require.NoError(t, err)

	resp2, err := get(client, base, "/artists/me", session)
	require.NoError(t, err)
	defer resp2.Body.Close()
	body2, _ := readBody(resp2)
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	require.NoError(t, json.Unmarshal(body2, &list))
	artists, _ := list["artists"].([]any)
	require.Len(t, artists, 1)
	require.Equal(t, "listmineband", artists[0].(map[string]any)["handle"])
}

func TestArtists_ListMine_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := get(client, base, "/artists/me", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestArtists_GetByHandle_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"publicband","display_name":"Public Band","bio":"Our story."}`, session)
	require.NoError(t, err)

	// Public: no auth required
	resp, err := get(client, base, "/artists/publicband", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var artist map[string]any
	require.NoError(t, json.Unmarshal(body, &artist))
	require.Equal(t, "publicband", artist["handle"])
	require.Equal(t, "Public Band", artist["display_name"])
	require.Equal(t, "Our story.", artist["bio"])
}

func TestArtists_GetByHandle_NotFound(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := get(client, base, "/artists/nonexistent", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestArtists_Update_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"updateme","display_name":"Old Name","bio":"Old bio"}`, session)
	require.NoError(t, err)

	resp, err := patchJSON(client, base, "/artists/updateme", `{"display_name":"New Name","bio":"New bio"}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var artist map[string]any
	require.NoError(t, json.Unmarshal(body, &artist))
	require.Equal(t, "New Name", artist["display_name"])
	require.Equal(t, "New bio", artist["bio"])

	// Public get reflects update
	resp2, err := get(client, base, "/artists/updateme", "")
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&artist))
	require.Equal(t, "New Name", artist["display_name"])
	require.Equal(t, "New bio", artist["bio"])
}

func TestArtists_Update_BioOnly(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"bioonly","display_name":"Name","bio":"Original"}`, session)
	require.NoError(t, err)

	// PATCH only bio (display_name omitted) — should keep display_name, update bio
	resp, err := patchJSON(client, base, "/artists/bioonly", `{"bio":"Updated bio only"}`, session)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := readBody(resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var artist map[string]any
	require.NoError(t, json.Unmarshal(body, &artist))
	require.Equal(t, "Name", artist["display_name"])
	require.Equal(t, "Updated bio only", artist["bio"])
}

func TestArtists_Update_Forbidden(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	ownerSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"notyours","display_name":"Mine"}`, ownerSession)
	require.NoError(t, err)

	otherSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	resp, err := patchJSON(client, base, "/artists/notyours", `{"display_name":"Hacked"}`, otherSession)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestArtists_Delete_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	session, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"deleteme","display_name":"Delete Me","bio":""}`, session)
	require.NoError(t, err)

	resp, err := deleteReq(client, base, "/artists/deleteme", session)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp2, err := get(client, base, "/artists/deleteme", "")
	require.NoError(t, err)
	resp2.Body.Close()
	require.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestArtists_Delete_Forbidden(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	ownerSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"notyours","display_name":"Mine"}`, ownerSession)
	require.NoError(t, err)

	otherSession, _, err := signupWithPKCE(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)

	resp, err := deleteReq(client, base, "/artists/notyours", otherSession)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestArtists_Members_AddListUpdateRemove(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	ownerSession, ownerID, err := signupWithPKCEAndMe(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	_, err = postJSON(client, base, "/artists", `{"handle":"memberband","display_name":"Member Band","bio":""}`, ownerSession)
	require.NoError(t, err)

	// Add member (second user)
	memberSession, memberID, err := signupWithPKCEAndMe(client, base, uniqueEmail(t), "password123", "web")
	require.NoError(t, err)
	require.NotEqual(t, ownerID, memberID)

	resp, err := postJSON(client, base, "/artists/memberband/members", `{"user_id":"`+memberID+`","roles":["feed","music"]}`, ownerSession)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// List members (includes owner + invited member)
	resp2, err := get(client, base, "/artists/memberband/members", ownerSession)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	var list map[string]any
	body2, _ := readBody(resp2)
	require.NoError(t, json.Unmarshal(body2, &list))
	members, _ := list["members"].([]any)
	require.Len(t, members, 2)
	var foundOwner, foundMember bool
	for _, m := range members {
		mid := m.(map[string]any)["user_id"].(string)
		roles, _ := m.(map[string]any)["roles"].([]any)
		if mid == ownerID {
			foundOwner = true
			require.Len(t, roles, 1)
			require.Equal(t, "owner", roles[0])
		}
		if mid == memberID {
			foundMember = true
			require.Len(t, roles, 2)
		}
	}
	require.True(t, foundOwner)
	require.True(t, foundMember)

	// Non-admin member (feed role) can list members but cannot add/update/remove
	respListAsMember, err := get(client, base, "/artists/memberband/members", memberSession)
	require.NoError(t, err)
	defer respListAsMember.Body.Close()
	require.Equal(t, http.StatusOK, respListAsMember.StatusCode, "any page member can list members")
	respAddAsMember, err := postJSON(client, base, "/artists/memberband/members", `{"user_id":"`+memberID+`","roles":["feed"]}`, memberSession)
	require.NoError(t, err)
	respAddAsMember.Body.Close()
	require.Equal(t, http.StatusForbidden, respAddAsMember.StatusCode, "only owner/admin can add member")
	respPatchAsMember, err := patchJSON(client, base, "/artists/memberband/members/"+memberID, `{"roles":["feed","music"]}`, memberSession)
	require.NoError(t, err)
	respPatchAsMember.Body.Close()
	require.Equal(t, http.StatusForbidden, respPatchAsMember.StatusCode, "only owner/admin can update member")
	respDelAsMember, err := deleteReq(client, base, "/artists/memberband/members/"+ownerID, memberSession)
	require.NoError(t, err)
	respDelAsMember.Body.Close()
	require.Equal(t, http.StatusForbidden, respDelAsMember.StatusCode, "only owner/admin can remove member")

	// Member with feed role can create post
	_, err = postJSON(client, base, "/artists/memberband/posts", `{"title":"From member","body":"Post by member"}`, memberSession)
	require.NoError(t, err)

	// Update member roles (remove music)
	resp3, err := patchJSON(client, base, "/artists/memberband/members/"+memberID, `{"roles":["feed"]}`, ownerSession)
	require.NoError(t, err)
	resp3.Body.Close()
	require.Equal(t, http.StatusNoContent, resp3.StatusCode)

	// Remove member
	resp4, err := deleteReq(client, base, "/artists/memberband/members/"+memberID, ownerSession)
	require.NoError(t, err)
	resp4.Body.Close()
	require.Equal(t, http.StatusNoContent, resp4.StatusCode)

	// Member can no longer create post
	resp5, err := postJSON(client, base, "/artists/memberband/posts", `{"title":"After remove","body":"Should fail"}`, memberSession)
	require.NoError(t, err)
	resp5.Body.Close()
	require.Equal(t, http.StatusForbidden, resp5.StatusCode)
}
