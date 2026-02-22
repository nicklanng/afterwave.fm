package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sopatech/afterwave.fm/internal/auth"
	"github.com/stretchr/testify/require"
)

// Web: 15 min session = 900 seconds. Native (ios/android/desktop): 30 days = 2592000 seconds.
const (
	expiresInWebSeconds    = 900     // 15 * 60
	expiresInNativeSeconds = 2592000 // 30 * 24 * 3600
)

func TestSignup_ClientWeb_ShortSessionTTL(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	session, refresh, expiresIn, err := signupWithPKCEAndExpires(client, base, email, "password123", "web")
	require.NoError(t, err)
	require.NotEmpty(t, session)
	require.NotEmpty(t, refresh)
	require.Equal(t, expiresInWebSeconds, expiresIn, "web client should get 15 min session (900s), got %d", expiresIn)
}

func TestSignup_ClientIOS_LongSessionTTL(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, expiresIn, err := signupWithPKCEAndExpires(client, base, email, "password123", "ios")
	require.NoError(t, err)
	require.Equal(t, expiresInNativeSeconds, expiresIn, "ios client should get 30-day session (2592000s), got %d", expiresIn)
}

func TestLogin_ClientAndroid_LongSessionTTL(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)

	_, _, expiresIn, err := loginWithPKCEAndExpires(client, base, email, "password123", "android")
	require.NoError(t, err)
	require.Equal(t, expiresInNativeSeconds, expiresIn, "android client should get 30-day session, got %d", expiresIn)
}

func TestLogin_ClientDesktop_LongSessionTTL(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	_, _, err := signupWithPKCE(client, base, email, "password123", "web")
	require.NoError(t, err)

	_, _, expiresIn, err := loginWithPKCEAndExpires(client, base, email, "password123", "desktop")
	require.NoError(t, err)
	require.Equal(t, expiresInNativeSeconds, expiresIn, "desktop client should get 30-day session, got %d", expiresIn)
}

func TestToken_InvalidCodeVerifier_Unauthorized(t *testing.T) {
	server, base := newTestServer(t)
	defer server.Close()
	client := server.Client()

	email := uniqueEmail(t)
	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	signupBody := fmt.Sprintf(`{"email":"%s","password":"password123","client_id":"web","code_challenge":"%s"}`, email, codeChallenge)
	resp, err := postJSON(client, base, "/auth/signup", signupBody, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := readBody(resp)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body: %s", string(b))
	authCode := parseAuthCode(b)
	require.NotEmpty(t, authCode)

	tokenBody := fmt.Sprintf(`{"grant_type":"authorization_code","client_id":"web","code":"%s","code_verifier":"wrong-verifier"}`, authCode)
	tokenResp, err := postJSON(client, base, "/auth/token", tokenBody, "")
	require.NoError(t, err)
	defer tokenResp.Body.Close()
	b2, _ := readBody(tokenResp)
	require.Equal(t, http.StatusUnauthorized, tokenResp.StatusCode, "body: %s", string(b2))
	require.Contains(t, string(b2), "invalid or expired authorization code")
}
