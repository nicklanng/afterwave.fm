package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- Refresh ---

func TestRefresh_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, refresh, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, refresh, "expected refresh token from signup")

	resp, err := postRefreshWithClientID(client, base, `{"refresh_token":"`+refresh+`"}`, "web")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, err := readBody(resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", string(b))
	newSession, newRefresh := parseTokenPair(b)
	require.NotEmpty(t, newSession, "expected new session token")
	require.NotEmpty(t, newRefresh, "expected new refresh token")
	require.NotEqual(t, refresh, newRefresh, "expected new refresh token (rolling refresh)")
}

func TestRefresh_InvalidToken(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := postRefreshWithClientID(client, base, `{"refresh_token":"invalid-refresh-id"}`, "web")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "body: %s", string(b))
}

func TestRefresh_BadRequest(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := postRefreshWithClientID(client, base, `{"refresh_token":""}`, "web")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "body: %s", string(b))
}

// --- Logout ---

func TestLogout_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	session, _, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, session, "expected session token from signup")

	req, err := http.NewRequest(http.MethodPost, base+"/auth/logout", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+session)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusNoContent, resp.StatusCode, "body: %s", string(b))
}

func TestLogout_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	req, err := http.NewRequest(http.MethodPost, base+"/auth/logout", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "body: %s", string(b))
}
