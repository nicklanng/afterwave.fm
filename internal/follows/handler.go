package follows

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

// Follow adds the artist to the current user's following list.
func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
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
	if err := h.svc.Follow(r.Context(), userID, handle); err != nil {
		if err == ErrArtistNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unfollow removes the artist from the current user's following list.
func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
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
	if err := h.svc.Unfollow(r.Context(), userID, handle); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListFollowing returns the list of artist handles the current user follows.
func (h *Handler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	handles, err := h.svc.ListFollowing(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if handles == nil {
		handles = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"handles": handles})
}
