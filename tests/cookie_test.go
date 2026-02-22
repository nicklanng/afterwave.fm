package tests

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogin_SetsSessionAndRefreshCookies(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)

	_, _, cookies, err := loginWithPKCEAndCookies(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.Contains(t, cookies, "session_token", "token response should set session_token cookie")
	require.Contains(t, cookies, "refresh_token", "token response should set refresh_token cookie")
	require.NotEmpty(t, cookies["session_token"])
	require.NotEmpty(t, cookies["refresh_token"])
}

func TestMe_SuccessWithCookieOnly(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, cookies, err := signupWithPKCEAndCookies(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, cookies["session_token"], "signup+token should set session cookie")

	// GET /users/me with only Cookie header (no Authorization Bearer)
	meResp, err := getWithCookies(client, base, "/users/me", cookies)
	require.NoError(t, err)
	defer meResp.Body.Close()
	b, err := readBody(meResp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, meResp.StatusCode, "body: %s", string(b))
	id, gotEmail := parseUser(b)
	require.NotEmpty(t, id)
	require.True(t, strings.EqualFold(gotEmail, email))
}

func TestLogout_ClearsSessionCookies(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, cookies, err := signupWithPKCEAndCookies(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, cookies["session_token"])

	// Logout with cookie (no Bearer)
	logoutReq, err := http.NewRequest(http.MethodPost, base+"/auth/logout", nil)
	require.NoError(t, err)
	for name, value := range cookies {
		logoutReq.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	logoutResp, err := client.Do(logoutReq)
	require.NoError(t, err)
	logoutResp.Body.Close()
	require.Equal(t, http.StatusNoContent, logoutResp.StatusCode)

	// Response should clear cookies (Set-Cookie with empty value or Max-Age=0)
	cleared := 0
	for _, c := range logoutResp.Cookies() {
		if c.Name == "session_token" || c.Name == "refresh_token" {
			if c.Value == "" || c.MaxAge <= 0 {
				cleared++
			}
		}
	}
	require.GreaterOrEqual(t, cleared, 1, "logout should clear at least one auth cookie")
}

func TestRefresh_SuccessWithCookieOnly(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, cookies, err := signupWithPKCEAndCookies(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, cookies["refresh_token"], "signup+token should set refresh cookie")

	// Refresh with Cookie + X-Client-ID (no secret)
	refreshReq, err := http.NewRequest(http.MethodPost, base+"/auth/refresh", bytes.NewBufferString("{}"))
	require.NoError(t, err)
	refreshReq.Header.Set("Content-Type", "application/json")
	setClientID(refreshReq, "web")
	for name, value := range cookies {
		refreshReq.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	resp, err := client.Do(refreshReq)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, err := readBody(resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", string(b))
	newSession, newRefresh := parseTokenPair(b)
	require.NotEmpty(t, newSession)
	require.NotEmpty(t, newRefresh)
	require.NotEqual(t, cookies["refresh_token"], newRefresh, "rolling refresh: new refresh token")
}
