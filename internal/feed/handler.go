package feed

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sopatech/afterwave.fm/internal/auth"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// CreatePost creates a post on the artist's feed (owner only).
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
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
		Title      string `json:"title"`
		Body       string `json:"body"`
		ImageURL   string `json:"image_url"`
		YouTubeURL string `json:"youtube_url"`
		Explicit   bool   `json:"explicit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	post, err := h.svc.CreatePost(r.Context(), handle, body.Title, body.Body, body.ImageURL, body.YouTubeURL, body.Explicit, userID)
	if err != nil {
		switch {
		case err == ErrArtistNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		case err == ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		case err == ErrSlugConflict:
			http.Error(w, "a post with this title already exists for this artist", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// ListPosts returns posts for the artist (public), newest first. Cursor-based pagination: limit (default 10), cursor (from previous next_cursor), has_more, next_cursor.
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	limit := 10
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursor := r.URL.Query().Get("cursor")

	list, nextCursor, err := h.svc.ListPosts(r.Context(), handle, limit, cursor)
	if err != nil {
		if err == ErrArtistNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []Post{}
	}

	out := map[string]any{"posts": list, "has_more": nextCursor != ""}
	if nextCursor != "" {
		out["next_cursor"] = nextCursor
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// GetPost returns a single post (public).
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	postID := r.PathValue("postId")
	if handle == "" || postID == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	post, err := h.svc.GetPost(r.Context(), handle, postID)
	if err != nil || post == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// UpdatePost updates a post (owner only).
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	handle := r.PathValue("handle")
	postID := r.PathValue("postId")
	if handle == "" || postID == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body struct {
		Body       *string `json:"body"`
		ImageURL   *string `json:"image_url"`
		YouTubeURL *string `json:"youtube_url"`
		Explicit   *bool   `json:"explicit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	post, err := h.svc.UpdatePost(r.Context(), handle, postID, body.Body, body.ImageURL, body.YouTubeURL, body.Explicit, userID)
	if err != nil {
		switch {
		case err == ErrArtistNotFound || err == ErrPostNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		case err == ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// MyFeed returns the collated feed for the current user (posts from artists they follow). Cursor-based pagination.
func (h *Handler) MyFeed(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	limit := 10
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursor := r.URL.Query().Get("cursor")
	list, nextCursor, err := h.svc.MyFeed(r.Context(), userID, limit, cursor)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []Post{}
	}
	out := map[string]any{"posts": list, "has_more": nextCursor != ""}
	if nextCursor != "" {
		out["next_cursor"] = nextCursor
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// DeletePost deletes a post (owner only).
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	handle := r.PathValue("handle")
	postID := r.PathValue("postId")
	if handle == "" || postID == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	err := h.svc.DeletePost(r.Context(), handle, postID, userID)
	if err != nil {
		switch {
		case err == ErrArtistNotFound || err == ErrPostNotFound:
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
