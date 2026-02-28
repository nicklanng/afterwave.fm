package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sopatech/afterwave.fm/internal/auth"
)

// PKCE test verifier (must be >= 43 chars for S256). Challenge computed via auth.ComputeCodeChallenge.
const testPKCEVerifier = "test-pkce-verifier-at-least-43-chars-long-for-s256"

// uniqueEmail returns an email unique per test and run (avoids conflicts when reusing the same DynamoDB table).
func uniqueEmail(t *testing.T) string {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "-")
	return fmt.Sprintf("%s-%d@test.example", name, time.Now().UnixNano())
}

// uniqueHandle returns a valid artist handle (lowercase alphanumeric, 4–64 chars) unique per call (avoids collisions across tests).
func uniqueHandle(t *testing.T, prefix string) string {
	t.Helper()
	if len(prefix) < 4 {
		prefix = "band"
	}
	return fmt.Sprintf("%s%x", prefix, time.Now().UnixNano()%0xffffff)
}

// setClientID sets X-Client-ID (public clients, no secret).
func setClientID(req *http.Request, clientID string) {
	req.Header.Set("X-Client-ID", clientID)
}

func postJSON(client *http.Client, baseURL, path, body string, authToken string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return client.Do(req)
}

// signupWithPKCE performs signup with PKCE and token exchange; returns session and refresh tokens.
func signupWithPKCE(client *http.Client, baseURL, email, password, clientID string) (sessionToken, refreshToken string, err error) {
	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	signupBody := fmt.Sprintf(`{"email":"%s","password":"%s","client_id":"%s","code_challenge":"%s"}`,
		email, password, clientID, codeChallenge)
	resp, err := postJSON(client, baseURL, "/auth/signup", signupBody, "")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("signup: %d %s", resp.StatusCode, string(b))
	}
	authCode := parseAuthCode(b)
	if authCode == "" {
		return "", "", fmt.Errorf("no authorization_code in response")
	}
	tokenBody := fmt.Sprintf(`{"grant_type":"authorization_code","client_id":"%s","code":"%s","code_verifier":"%s"}`,
		clientID, authCode, testPKCEVerifier)
	tokenResp, err := postJSON(client, baseURL, "/auth/token", tokenBody, "")
	if err != nil {
		return "", "", err
	}
	defer tokenResp.Body.Close()
	b2, _ := io.ReadAll(tokenResp.Body)
	if tokenResp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("token: %d %s", tokenResp.StatusCode, string(b2))
	}
	session, refresh := parseTokenPair(b2)
	return session, refresh, nil
}

// signupWithPKCEAndMe performs signup with PKCE, token exchange, and GET /users/me; returns session, user ID, and error.
func signupWithPKCEAndMe(client *http.Client, baseURL, email, password, clientID string) (sessionToken, userID string, err error) {
	session, _, err := signupWithPKCE(client, baseURL, email, password, clientID)
	if err != nil {
		return "", "", err
	}
	resp, err := get(client, baseURL, "/users/me", session)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("me: %d %s", resp.StatusCode, string(b))
	}
	var me map[string]any
	if err := json.Unmarshal(b, &me); err != nil {
		return "", "", err
	}
	id, _ := me["id"].(string)
	return session, id, nil
}

// signupWithPKCEAndExpires is like signupWithPKCE but also returns expires_in from the token response.
func signupWithPKCEAndExpires(client *http.Client, baseURL, email, password, clientID string) (sessionToken, refreshToken string, expiresIn int, err error) {
	session, refresh, tokenBody, _, err := signupAndTokenResponse(client, baseURL, email, password, clientID)
	if err != nil {
		return "", "", 0, err
	}
	return session, refresh, parseExpiresIn(tokenBody), nil
}

func signupAndTokenResponse(client *http.Client, baseURL, email, password, clientID string) (session, refresh string, tokenBody []byte, cookies map[string]string, err error) {
	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	signupBody := fmt.Sprintf(`{"email":"%s","password":"%s","client_id":"%s","code_challenge":"%s"}`,
		email, password, clientID, codeChallenge)
	resp, err := postJSON(client, baseURL, "/auth/signup", signupBody, "")
	if err != nil {
		return "", "", nil, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", nil, nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", "", nil, nil, fmt.Errorf("signup: %d %s", resp.StatusCode, string(b))
	}
	authCode := parseAuthCode(b)
	if authCode == "" {
		return "", "", nil, nil, fmt.Errorf("no authorization_code in response")
	}
	tokenBodyStr := fmt.Sprintf(`{"grant_type":"authorization_code","client_id":"%s","code":"%s","code_verifier":"%s"}`,
		clientID, authCode, testPKCEVerifier)
	tokenResp, err := postJSON(client, baseURL, "/auth/token", tokenBodyStr, "")
	if err != nil {
		return "", "", nil, nil, err
	}
	defer tokenResp.Body.Close()
	tokenBody, err = io.ReadAll(tokenResp.Body)
	if err != nil {
		return "", "", nil, nil, err
	}
	if tokenResp.StatusCode != http.StatusOK {
		return "", "", nil, nil, fmt.Errorf("token: %d %s", tokenResp.StatusCode, string(tokenBody))
	}
	session, refresh = parseTokenPair(tokenBody)
	return session, refresh, tokenBody, cookiesFromResponse(tokenResp), nil
}

// loginWithPKCEAndExpires is like loginWithPKCE but also returns expires_in from the token response.
func loginWithPKCEAndExpires(client *http.Client, baseURL, email, password, clientID string) (sessionToken, refreshToken string, expiresIn int, err error) {
	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	loginBody := fmt.Sprintf(`{"email":"%s","password":"%s","client_id":"%s","code_challenge":"%s"}`,
		email, password, clientID, codeChallenge)
	resp, err := postJSON(client, baseURL, "/auth/login", loginBody, "")
	if err != nil {
		return "", "", 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("login: %d %s", resp.StatusCode, string(b))
	}
	authCode := parseAuthCode(b)
	if authCode == "" {
		return "", "", 0, fmt.Errorf("no authorization_code in response")
	}
	tokenBodyStr := fmt.Sprintf(`{"grant_type":"authorization_code","client_id":"%s","code":"%s","code_verifier":"%s"}`,
		clientID, authCode, testPKCEVerifier)
	tokenResp, err := postJSON(client, baseURL, "/auth/token", tokenBodyStr, "")
	if err != nil {
		return "", "", 0, err
	}
	defer tokenResp.Body.Close()
	tokenBody, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return "", "", 0, err
	}
	if tokenResp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("token: %d %s", tokenResp.StatusCode, string(tokenBody))
	}
	session, refresh := parseTokenPair(tokenBody)
	return session, refresh, parseExpiresIn(tokenBody), nil
}

// loginWithPKCEAndCookies performs login+token exchange and returns tokens and cookies from the token response.
func loginWithPKCEAndCookies(client *http.Client, baseURL, email, password, clientID string) (session, refresh string, cookies map[string]string, err error) {
	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	loginBody := fmt.Sprintf(`{"email":"%s","password":"%s","client_id":"%s","code_challenge":"%s"}`,
		email, password, clientID, codeChallenge)
	resp, err := postJSON(client, baseURL, "/auth/login", loginBody, "")
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", nil, fmt.Errorf("login: %d %s", resp.StatusCode, string(b))
	}
	authCode := parseAuthCode(b)
	if authCode == "" {
		return "", "", nil, fmt.Errorf("no authorization_code in response")
	}
	tokenBodyStr := fmt.Sprintf(`{"grant_type":"authorization_code","client_id":"%s","code":"%s","code_verifier":"%s"}`,
		clientID, authCode, testPKCEVerifier)
	tokenResp, err := postJSON(client, baseURL, "/auth/token", tokenBodyStr, "")
	if err != nil {
		return "", "", nil, err
	}
	defer tokenResp.Body.Close()
	b2, _ := io.ReadAll(tokenResp.Body)
	if tokenResp.StatusCode != http.StatusOK {
		return "", "", nil, fmt.Errorf("token: %d %s", tokenResp.StatusCode, string(b2))
	}
	session, refresh = parseTokenPair(b2)
	return session, refresh, cookiesFromResponse(tokenResp), nil
}

// signupWithPKCEAndCookies performs signup+token exchange and returns tokens and cookies from the token response.
func signupWithPKCEAndCookies(client *http.Client, baseURL, email, password, clientID string) (session, refresh string, cookies map[string]string, err error) {
	s, r, _, c, e := signupAndTokenResponse(client, baseURL, email, password, clientID)
	if e != nil {
		return "", "", nil, e
	}
	return s, r, c, nil
}

// loginWithPKCE performs login with PKCE and token exchange; returns session and refresh tokens.
func loginWithPKCE(client *http.Client, baseURL, email, password, clientID string) (sessionToken, refreshToken string, err error) {
	codeChallenge := auth.ComputeCodeChallenge(testPKCEVerifier)
	loginBody := fmt.Sprintf(`{"email":"%s","password":"%s","client_id":"%s","code_challenge":"%s"}`,
		email, password, clientID, codeChallenge)
	resp, err := postJSON(client, baseURL, "/auth/login", loginBody, "")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("login: %d %s", resp.StatusCode, string(b))
	}
	authCode := parseAuthCode(b)
	if authCode == "" {
		return "", "", fmt.Errorf("no authorization_code in response")
	}
	tokenBody := fmt.Sprintf(`{"grant_type":"authorization_code","client_id":"%s","code":"%s","code_verifier":"%s"}`,
		clientID, authCode, testPKCEVerifier)
	tokenResp, err := postJSON(client, baseURL, "/auth/token", tokenBody, "")
	if err != nil {
		return "", "", err
	}
	defer tokenResp.Body.Close()
	b2, _ := io.ReadAll(tokenResp.Body)
	if tokenResp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("token: %d %s", tokenResp.StatusCode, string(b2))
	}
	session, refresh := parseTokenPair(b2)
	return session, refresh, nil
}

func parseAuthCode(b []byte) string {
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	if s, ok := m["authorization_code"].(string); ok {
		return s
	}
	return ""
}

func get(client *http.Client, baseURL, path, authToken string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return client.Do(req)
}

func patchJSON(client *http.Client, baseURL, path, body string, authToken string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPatch, baseURL+path, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return client.Do(req)
}

func deleteReq(client *http.Client, baseURL, path, authToken string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return client.Do(req)
}

func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func parseTokenPair(b []byte) (sessionToken, refreshToken string) {
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		return "", ""
	}
	if s, ok := m["session_token"].(string); ok {
		sessionToken = s
	}
	if s, ok := m["refresh_token"].(string); ok {
		refreshToken = s
	}
	return sessionToken, refreshToken
}

func parseUser(b []byte) (id, email string) {
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		return "", ""
	}
	if s, ok := m["id"].(string); ok {
		id = s
	}
	if s, ok := m["email"].(string); ok {
		email = s
	}
	return id, email
}

// parseExpiresIn returns expires_in from a token-pair JSON body.
func parseExpiresIn(b []byte) int {
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		return 0
	}
	if n, ok := m["expires_in"].(float64); ok {
		return int(n)
	}
	return 0
}

// cookiesFromResponse parses Set-Cookie headers and returns a map of name -> value.
func cookiesFromResponse(resp *http.Response) map[string]string {
	out := make(map[string]string)
	for _, c := range resp.Cookies() {
		out[c.Name] = c.Value
	}
	return out
}

// getWithCookies performs GET with Cookie header (no Authorization).
func getWithCookies(client *http.Client, baseURL, path string, cookies map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	return client.Do(req)
}

// postWithCookies performs POST with Cookie header and optional body.
func postWithCookies(client *http.Client, baseURL, path, body string, cookies map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	return client.Do(req)
}

// postRefreshWithClientID POSTs to /refresh with X-Client-ID and optional body (e.g. refresh_token).
func postRefreshWithClientID(client *http.Client, baseURL, body, clientID string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, baseURL+"/auth/refresh", bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	setClientID(req, clientID)
	return client.Do(req)
}
