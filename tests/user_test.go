package tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sopatech/afterwave.fm/internal/auth"
	"github.com/stretchr/testify/require"
)

// --- Signup ---

func TestSignup_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	session, refresh, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, session, "expected session_token from token exchange")
	require.NotEmpty(t, refresh, "expected refresh_token from token exchange")
}

func TestSignup_DuplicateEmail(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)

	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	body := fmt.Sprintf(`{"email":"%s","password":"password123","client_id":"web","code_challenge":"%s"}`, email, codeChallenge)
	resp, err := postJSON(client, base, "/auth/signup", body, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "body: %s", string(b))
	require.NotContains(t, string(b), "email already registered", "should not enumerate emails")
}

func TestSignup_BadRequest(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	tests := []struct {
		name string
		body string
	}{
		{"missing body", ""},
		{"invalid JSON", "{"},
		{"short password", fmt.Sprintf(`{"email":"a@b.co","password":"short","client_id":"web","code_challenge":"%s"}`, codeChallenge)},
		{"empty email", fmt.Sprintf(`{"email":"","password":"password123","client_id":"web","code_challenge":"%s"}`, codeChallenge)},
		{"missing code_challenge", `{"email":"a@b.co","password":"password123","client_id":"web"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := postJSON(client, base, "/auth/signup", tt.body, "")
			require.NoError(t, err)
			defer resp.Body.Close()
			b, _ := readBody(resp)
			require.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, resp.StatusCode, "body: %s", string(b))
		})
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)

	session, refresh, err := loginWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, session, "expected session_token")
	require.NotEmpty(t, refresh, "expected refresh_token")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	body := fmt.Sprintf(`{"email":"%s","password":"wrongpassword","client_id":"web","code_challenge":"%s"}`, email, codeChallenge)
	resp, err := postJSON(client, base, "/auth/login", body, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "body: %s", string(b))
}

func TestLogin_BadRequest(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := postJSON(client, base, "/auth/login", "invalid", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "body: %s", string(b))
}

// --- Users/me ---

func TestMe_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	session, _, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, session, "expected session token from signup")

	resp, err := get(client, base, "/users/me", session)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, err := readBody(resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", string(b))
	id, gotEmail := parseUser(b)
	require.NotEmpty(t, id, "expected user id in response")
	require.NotEmpty(t, gotEmail, "expected email in response")
	require.True(t, strings.EqualFold(gotEmail, email), "expected email %q (case-insensitive), got %q", email, gotEmail)
}

func TestMe_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := get(client, base, "/users/me", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "body: %s", string(b))

	resp2, err := get(client, base, "/users/me", "invalid-token")
	require.NoError(t, err)
	defer resp2.Body.Close()
	b2, _ := readBody(resp2)
	require.Equal(t, http.StatusUnauthorized, resp2.StatusCode, "invalid token: body: %s", string(b2))
}

// --- Delete account ---

func TestDeleteAccount_Success(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	session, _, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, session, "expected session token from signup")

	resp, err := deleteReq(client, base, "/account", session)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusNoContent, resp.StatusCode, "body: %s", string(b))

	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	loginBody := fmt.Sprintf(`{"email":"%s","password":"password123","client_id":"web","code_challenge":"%s"}`, email, codeChallenge)
	loginResp, err := postJSON(client, base, "/auth/login", loginBody, "")
	require.NoError(t, err)
	loginResp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, loginResp.StatusCode, "after delete, login should fail")
}

func TestDeleteAccount_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	resp, err := deleteReq(client, base, "/account", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "body: %s", string(b))
}
