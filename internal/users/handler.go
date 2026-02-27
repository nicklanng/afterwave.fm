package users

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sopatech/afterwave.fm/internal/auth"
	"github.com/sopatech/afterwave.fm/internal/cognito"
)

const oauthStateCookieName = "oauth_state"
const oauthPKCECookieName = "oauth_pkce"
const oauthLinkUserCookieName = "oauth_link_user"
const oauthStateCookieMaxAge = 600 // 10 minutes

type Handler struct {
	svc                 Service
	authSvc             *auth.Service
	cookie              auth.CookieConfig
	cognitoDomain       string
	cognitoRegion       string
	cognitoUserPoolID   string
	cognitoClientID     string
	cognitoClientSecret string
	callbackURL         string
	frontendRedirectURI  string
	oauthStateSecret    string
}

func NewHandler(svc Service, authSvc *auth.Service, cookie auth.CookieConfig, cognitoDomain, cognitoRegion, cognitoUserPoolID, cognitoClientID, cognitoClientSecret, callbackURL, frontendRedirectURI, oauthStateSecret string) *Handler {
	return &Handler{
		svc:                 svc,
		authSvc:             authSvc,
		cookie:              cookie,
		cognitoDomain:       cognitoDomain,
		cognitoRegion:       cognitoRegion,
		cognitoUserPoolID:   cognitoUserPoolID,
		cognitoClientID:     cognitoClientID,
		cognitoClientSecret: cognitoClientSecret,
		callbackURL:         callbackURL,
		frontendRedirectURI: frontendRedirectURI,
		oauthStateSecret:    oauthStateSecret,
	}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email               string `json:"email"`
		Password            string `json:"password"`
		ClientID            string `json:"client_id"`
		CodeChallenge       string `json:"code_challenge"`
		CodeChallengeMethod string `json:"code_challenge_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.ClientID == "" || body.CodeChallenge == "" {
		http.Error(w, "client_id and code_challenge required", http.StatusBadRequest)
		return
	}

	userID, err := h.svc.Signup(r.Context(), body.Email, body.Password)
	if err != nil {
		switch {
		case err == ErrEmailTaken:
			http.Error(w, "signup failed", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	codeChallengeMethod := body.CodeChallengeMethod
	if codeChallengeMethod == "" {
		codeChallengeMethod = auth.CodeChallengeMethodS256
	}
	code, expiresIn, err := h.authSvc.CreateAuthCode(r.Context(), userID, body.ClientID, body.CodeChallenge, codeChallengeMethod)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"authorization_code": code,
		"expires_in":         expiresIn,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email               string `json:"email"`
		Password            string `json:"password"`
		ClientID            string `json:"client_id"`
		CodeChallenge       string `json:"code_challenge"`
		CodeChallengeMethod string `json:"code_challenge_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.ClientID == "" || body.CodeChallenge == "" {
		http.Error(w, "client_id and code_challenge required", http.StatusBadRequest)
		return
	}

	userID, err := h.svc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		if err == ErrInvalidCreds {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	codeChallengeMethod := body.CodeChallengeMethod
	if codeChallengeMethod == "" {
		codeChallengeMethod = auth.CodeChallengeMethodS256
	}
	code, expiresIn, err := h.authSvc.CreateAuthCode(r.Context(), userID, body.ClientID, body.CodeChallenge, codeChallengeMethod)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"authorization_code": code,
		"expires_in":         expiresIn,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		if err == ErrUserNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.authSvc.RevokeAllSessionsForUser(r.Context(), userID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := h.svc.DeleteAccount(r.Context(), userID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GoogleAuthRedirect starts the Cognito Hosted UI flow for Google.
func (h *Handler) GoogleAuthRedirect(w http.ResponseWriter, r *http.Request) {
	h.redirectToIdP(w, r, "Google")
}

// AppleAuthRedirect starts the Cognito Hosted UI flow for Apple.
func (h *Handler) AppleAuthRedirect(w http.ResponseWriter, r *http.Request) {
	h.redirectToIdP(w, r, "Apple")
}

// LinkGoogleRedirect starts the "link Google account" flow. Requires auth. Redirects to Cognito Hosted UI; on return, the Google identity is linked to the current user.
func (h *Handler) LinkGoogleRedirect(w http.ResponseWriter, r *http.Request) {
	h.redirectToLinkIdP(w, r, "Google")
}

// LinkAppleRedirect starts the "link Apple account" flow. Requires auth. Redirects to Cognito Hosted UI; on return, the Apple identity is linked to the current user.
func (h *Handler) LinkAppleRedirect(w http.ResponseWriter, r *http.Request) {
	h.redirectToLinkIdP(w, r, "Apple")
}

func (h *Handler) redirectToLinkIdP(w http.ResponseWriter, r *http.Request, provider string) {
	if h.cognitoDomain == "" || h.cognitoClientID == "" || h.callbackURL == "" {
		http.Error(w, "federated login not configured", http.StatusNotImplemented)
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if h.oauthStateSecret == "" {
		http.Error(w, "account linking requires OAUTH_STATE_SECRET", http.StatusInternalServerError)
		return
	}
	state := r.URL.Query().Get("state")
	stateB := make([]byte, 32)
	if _, err := rand.Read(stateB); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	state = base64.URLEncoding.EncodeToString(stateB)
	mac := hmac.New(sha256.New, []byte(h.oauthStateSecret))
	mac.Write([]byte(state))
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state + "." + hex.EncodeToString(mac.Sum(nil)),
		Path:     "/",
		MaxAge:   oauthStateCookieMaxAge,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	linkVal := base64.URLEncoding.EncodeToString([]byte(userID))
	mac2 := hmac.New(sha256.New, []byte(h.oauthStateSecret))
	mac2.Write([]byte(linkVal))
	http.SetCookie(w, &http.Cookie{
		Name:     oauthLinkUserCookieName,
		Value:    linkVal + "." + hex.EncodeToString(mac2.Sum(nil)),
		Path:     "/",
		MaxAge:   oauthStateCookieMaxAge,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	q := url.Values{}
	q.Set("client_id", h.cognitoClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid email")
	q.Set("redirect_uri", h.callbackURL)
	q.Set("identity_provider", provider)
	q.Set("state", state)
	redirectURL := h.cognitoDomain + "/oauth2/authorize?" + q.Encode()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) redirectToIdP(w http.ResponseWriter, r *http.Request, provider string) {
	if h.cognitoDomain == "" || h.cognitoClientID == "" || h.callbackURL == "" {
		http.Error(w, "federated login not configured", http.StatusNotImplemented)
		return
	}
	clientID := strings.TrimSpace(r.URL.Query().Get("client_id"))
	codeChallenge := strings.TrimSpace(r.URL.Query().Get("code_challenge"))
	codeMethod := strings.TrimSpace(r.URL.Query().Get("code_challenge_method"))
	if clientID == "" || codeChallenge == "" {
		http.Error(w, "client_id and code_challenge required for federated login", http.StatusBadRequest)
		return
	}
	if codeMethod == "" {
		codeMethod = auth.CodeChallengeMethodS256
	}

	state := r.URL.Query().Get("state")
	if h.oauthStateSecret != "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		state = base64.URLEncoding.EncodeToString(b)
		mac := hmac.New(sha256.New, []byte(h.oauthStateSecret))
		mac.Write([]byte(state))
		sig := hex.EncodeToString(mac.Sum(nil))
		http.SetCookie(w, &http.Cookie{
			Name:     oauthStateCookieName,
			Value:    state + "." + sig,
			Path:     "/",
			MaxAge:   oauthStateCookieMaxAge,
			HttpOnly: true,
			Secure:   h.cookie.Secure,
			SameSite: http.SameSiteLaxMode,
		})
	}
	// Store PKCE params for callback (Cognito only echoes code and state).
	pkceVal := base64.URLEncoding.EncodeToString([]byte(clientID + "\n" + codeChallenge + "\n" + codeMethod))
	http.SetCookie(w, &http.Cookie{
		Name:     oauthPKCECookieName,
		Value:    pkceVal,
		Path:     "/",
		MaxAge:   oauthStateCookieMaxAge,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	q := url.Values{}
	q.Set("client_id", h.cognitoClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid email")
	q.Set("redirect_uri", h.callbackURL)
	q.Set("identity_provider", provider)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL := h.cognitoDomain + "/oauth2/authorize?" + q.Encode()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// pkceFromCallbackCookie reads client_id, code_challenge, and code_challenge_method from the oauth_pkce cookie,
// clears the cookie, and returns the values. Returns empty strings if missing or invalid.
func (h *Handler) pkceFromCallbackCookie(r *http.Request, w http.ResponseWriter) (clientID, codeChallenge, codeMethod string) {
	cookie, err := r.Cookie(oauthPKCECookieName)
	if err != nil || cookie.Value == "" {
		return "", "", ""
	}
	dec, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", "", ""
	}
	parts := strings.SplitN(string(dec), "\n", 3)
	if len(parts) < 2 {
		return "", "", ""
	}
	clientID = strings.TrimSpace(parts[0])
	codeChallenge = strings.TrimSpace(parts[1])
	if len(parts) > 2 {
		codeMethod = strings.TrimSpace(parts[2])
	}
	// Clear cookie so it cannot be reused
	http.SetCookie(w, &http.Cookie{
		Name:     oauthPKCECookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	return clientID, codeChallenge, codeMethod
}

// linkUserIDFromCookie reads and verifies the oauth_link_user cookie. Returns the user ID if valid, or "".
// Call clearLinkCookie to clear it after handling the link flow.
func (h *Handler) linkUserIDFromCookie(r *http.Request) string {
	if h.oauthStateSecret == "" {
		return ""
	}
	cookie, err := r.Cookie(oauthLinkUserCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(h.oauthStateSecret))
	mac.Write([]byte(parts[0]))
	if hex.EncodeToString(mac.Sum(nil)) != parts[1] {
		return ""
	}
	dec, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return ""
	}
	return string(dec)
}

func (h *Handler) clearLinkCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthLinkUserCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// redirectToFrontendWithQuery redirects to frontendRedirectURI with the given query key and value.
// If frontendRedirectURI is empty, writes a JSON object with the key instead.
func (h *Handler) redirectToFrontendWithQuery(w http.ResponseWriter, r *http.Request, key, value string) {
	if h.frontendRedirectURI == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{key: value})
		return
	}
	u, err := url.Parse(h.frontendRedirectURI)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

// FederatedCallback handles the redirect back from Cognito Hosted UI (Google/Apple).
func (h *Handler) FederatedCallback(w http.ResponseWriter, r *http.Request) {
	if h.cognitoDomain == "" || h.cognitoClientID == "" || h.callbackURL == "" {
		http.Error(w, "federated login not configured", http.StatusNotImplemented)
		return
	}
	queryState := r.URL.Query().Get("state")
	if h.oauthStateSecret != "" {
		cookie, err := r.Cookie(oauthStateCookieName)
		if err != nil || cookie.Value == "" {
			http.Error(w, "missing or invalid state", http.StatusBadRequest)
			return
		}
		parts := strings.SplitN(cookie.Value, ".", 2)
		if len(parts) != 2 {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		mac := hmac.New(sha256.New, []byte(h.oauthStateSecret))
		mac.Write([]byte(parts[0]))
		if hex.EncodeToString(mac.Sum(nil)) != parts[1] {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		if parts[0] != queryState {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		// Clear state cookie
		http.SetCookie(w, &http.Cookie{
			Name:     oauthStateCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   h.cookie.Secure,
			SameSite: http.SameSiteLaxMode,
		})
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", h.cognitoClientID)
	form.Set("redirect_uri", h.callbackURL)
	form.Set("code", code)
	if h.cognitoClientSecret != "" {
		form.Set("client_secret", h.cognitoClientSecret)
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.cognitoDomain+"/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = body
		http.Error(w, "federated token exchange failed", http.StatusUnauthorized)
		return
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tokenResp.IDToken == "" {
		http.Error(w, "missing id_token", http.StatusUnauthorized)
		return
	}

	sub, email, err := cognito.ValidateIDToken(r.Context(), tokenResp.IDToken, h.cognitoRegion, h.cognitoUserPoolID, h.cognitoClientID)
	if err != nil {
		http.Error(w, "invalid id_token", http.StatusUnauthorized)
		return
	}

	// Account-linking flow: user was authenticated and started "link Google/Apple"; link this IdP to their account.
	if linkUserID := h.linkUserIDFromCookie(r); linkUserID != "" {
		h.clearLinkCookie(w)
		if err := h.svc.LinkCognitoSub(r.Context(), linkUserID, sub); err != nil {
			if errors.Is(err, ErrSubLinkedToOtherAccount) {
				h.redirectToFrontendWithQuery(w, r, "error", "already_linked")
				return
			}
			http.Error(w, "failed to link account", http.StatusInternalServerError)
			return
		}
		h.redirectToFrontendWithQuery(w, r, "linked", "1")
		return
	}

	userID, err := h.svc.EnsureUserForCognito(r.Context(), email, sub)
	if err != nil {
		if err == ErrAccountExistsWithPassword {
			http.Error(w, "account already exists with email and password; use password login", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Issue our auth code for the client to exchange via /auth/token.
	// PKCE params were stored in cookie when starting the IdP flow (Cognito only sends code and state).
	clientID, codeChallenge, codeMethod := h.pkceFromCallbackCookie(r, w)
	if clientID == "" || codeChallenge == "" {
		http.Error(w, "client_id and code_challenge required (start federated login with query params)", http.StatusBadRequest)
		return
	}
	if codeMethod == "" {
		codeMethod = auth.CodeChallengeMethodS256
	}

	authCode, _, err := h.authSvc.CreateAuthCode(r.Context(), userID, clientID, codeChallenge, codeMethod)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Redirect back to frontend with the authorization code.
	if h.frontendRedirectURI == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"authorization_code": authCode})
		return
	}

	u, err := url.Parse(h.frontendRedirectURI)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	q := u.Query()
	q.Set("code", authCode)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}
