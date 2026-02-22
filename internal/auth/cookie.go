package auth

import (
	"net/http"
	"strings"
	"time"
)

const (
	SessionCookieName = "session_token"
	RefreshCookieName = "refresh_token"
)

// CookieConfig controls cookie attributes (e.g. Secure for HTTPS).
type CookieConfig struct {
	Secure bool // Secure=true means cookies only sent over HTTPS. Set false for local HTTP dev.
}

// DefaultCookieConfig returns Secure=true. For local HTTP dev, set Secure to false (e.g. from env).
func DefaultCookieConfig() CookieConfig {
	return CookieConfig{Secure: true}
}

// SetSessionCookies sets httpOnly cookies for session and refresh tokens on w.
// maxAgeSeconds is the session lifetime (used for session cookie); refresh cookie uses a longer maxAge (e.g. 90 days).
func SetSessionCookies(w http.ResponseWriter, cfg CookieConfig, sessionToken, refreshToken string, sessionMaxAgeSeconds, refreshMaxAgeSeconds int) {
	sessionCookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   sessionMaxAgeSeconds,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	refreshCookie := &http.Cookie{
		Name:     RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   refreshMaxAgeSeconds,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, sessionCookie)
	http.SetCookie(w, refreshCookie)
}

// ClearSessionCookies clears the session and refresh cookies (e.g. on logout).
func ClearSessionCookies(w http.ResponseWriter, cfg CookieConfig) {
	for _, name := range []string{SessionCookieName, RefreshCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   cfg.Secure,
			SameSite: http.SameSiteStrictMode,
		})
	}
}

// SessionTokenFromRequest returns the session token from Cookie or Authorization Bearer. Prefers cookie.
func SessionTokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(SessionCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(prefix):])
}

// RefreshTokenFromRequest returns the refresh token from Cookie or from the given bodyValue (for backward compat / mobile).
func RefreshTokenFromRequest(r *http.Request, bodyValue string) string {
	if c, err := r.Cookie(RefreshCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	return bodyValue
}

