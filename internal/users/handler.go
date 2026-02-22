package users

import (
	"encoding/json"
	"net/http"

	"github.com/sopatech/afterwave.fm/internal/auth"
)

type Handler struct {
	svc     Service
	authSvc *auth.Service
	cookie  auth.CookieConfig
}

func NewHandler(svc Service, authSvc *auth.Service, cookie auth.CookieConfig) *Handler {
	return &Handler{svc: svc, authSvc: authSvc, cookie: cookie}
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
			http.Error(w, "email already registered", http.StatusConflict)
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
