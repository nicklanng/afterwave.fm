package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	svc    *Service
	cookie CookieConfig
}

func NewHandler(svc *Service, cookie CookieConfig) *Handler {
	return &Handler{svc: svc, cookie: cookie}
}

// Token exchanges authorization_code + code_verifier for tokens (PKCE). Sets httpOnly cookies.
func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GrantType    string `json:"grant_type"`
		ClientID     string `json:"client_id"`
		Code         string `json:"code"`
		CodeVerifier string `json:"code_verifier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.GrantType != "authorization_code" || body.ClientID == "" || body.Code == "" || body.CodeVerifier == "" {
		http.Error(w, "grant_type, client_id, code, and code_verifier required", http.StatusBadRequest)
		return
	}
	pair, err := h.svc.ExchangeCode(r.Context(), body.Code, body.CodeVerifier, body.ClientID)
	if err != nil {
		if err == ErrAuthCodeInvalid {
			http.Error(w, "invalid or expired authorization code", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	SetSessionCookies(w, h.cookie, pair.SessionToken, pair.RefreshToken, pair.ExpiresIn, pair.RefreshExpiresIn)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pair)
}

// Refresh expects refresh_token in Cookie or JSON body and client_id (X-Client-ID). Returns new tokens and sets httpOnly cookies.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	refreshToken := RefreshTokenFromRequest(r, body.RefreshToken)
	if refreshToken == "" {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}
	clientID := strings.TrimSpace(r.Header.Get("X-Client-ID"))
	if clientID == "" {
		http.Error(w, "X-Client-ID required", http.StatusBadRequest)
		return
	}
	ttls, err := h.svc.GetClientTTLs(r.Context(), clientID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if ttls == nil {
		http.Error(w, "unknown client", http.StatusUnauthorized)
		return
	}
	pair, err := h.svc.Refresh(r.Context(), refreshToken, ttls)
	if err != nil {
		if err == ErrInvalidRefreshToken {
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	SetSessionCookies(w, h.cookie, pair.SessionToken, pair.RefreshToken, pair.ExpiresIn, pair.RefreshExpiresIn)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pair)
}

// Logout revokes the current session (and its linked refresh token) and clears session cookies.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := SessionIDFromContext(r.Context())
	if sessionID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.Logout(r.Context(), sessionID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ClearSessionCookies(w, h.cookie)
	w.WriteHeader(http.StatusNoContent)
}
