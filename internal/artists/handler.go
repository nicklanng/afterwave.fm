package artists

import (
	"encoding/json"
	"net/http"

	"github.com/sopatech/afterwave.fm/internal/auth"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Create creates an artist page (owner = authenticated user).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Handle      string `json:"handle"`
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	artist, err := h.svc.Create(r.Context(), userID, body.Handle, body.DisplayName, body.Bio)
	if err != nil {
		switch {
		case err == ErrHandleTaken:
			http.Error(w, "handle already in use", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(artist)
}

// ListMine returns the authenticated user's artist pages.
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	list, err := h.svc.ListByOwner(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []Artist{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"artists": list})
}

// GetByHandle returns an artist by handle (public).
func (h *Handler) GetByHandle(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	artist, err := h.svc.GetByHandle(r.Context(), handle)
	if err != nil || artist == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artist)
}

// Update updates an artist (owner only).
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body struct {
		DisplayName *string `json:"display_name"`
		Bio         *string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	artist, err := h.svc.Update(r.Context(), handle, body.DisplayName, body.Bio, userID)
	if err != nil {
		switch {
		case err == ErrArtistNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		case err == ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artist)
}

// Delete deletes an artist page (owner only).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	err := h.svc.Delete(r.Context(), handle, userID)
	if err != nil {
		switch {
		case err == ErrArtistNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		case err == ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
