package artists

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/guregu/dynamo/v2"
)

var (
	ErrHandleTaken     = errors.New("handle already in use")
	ErrArtistNotFound  = errors.New("artist not found")
	ErrForbidden       = errors.New("forbidden")
	ErrCannotRemoveOwner = errors.New("cannot remove the owner")
	ErrInvalidRoles    = errors.New("invalid roles")
)

// Handle must be lowercase, alphanumeric only, 4–64 chars (min 4 so we can reserve 3-letter subdomains: www, tui, api, etc.).
var handleRegex = regexp.MustCompile(`^[a-z0-9]{4,64}$`)

type Service interface {
	Create(ctx context.Context, ownerUserID, handle, displayName, bio string) (*Artist, error)
	GetByHandle(ctx context.Context, handle string) (*Artist, error)
	ListByOwner(ctx context.Context, userID string) ([]Artist, error)
	ListForUser(ctx context.Context, userID string) ([]ArtistWithRole, error)
	Update(ctx context.Context, handle string, displayName, bio *string, actorUserID string) (*Artist, error)
	Delete(ctx context.Context, handle, actorUserID string) error
	HasPermission(ctx context.Context, handle, userID, permission string) (bool, error)
	AddMember(ctx context.Context, handle, userID string, roles []string, actorUserID string) error
	RemoveMember(ctx context.Context, handle, userID string, actorUserID string) error
	UpdateMemberRoles(ctx context.Context, handle, userID string, roles []string, actorUserID string) error
	ListMembers(ctx context.Context, handle, actorUserID string) ([]Member, error)
}

// ArtistWithRole is an artist plus the current user's role(s). Used for GET /artists/me.
type ArtistWithRole struct {
	Artist
	Role  string   `json:"role"`  // "owner" or "member"
	Roles []string `json:"roles,omitempty"` // when role is "member", the list of roles
}

// Member is a user with roles on an artist page (excludes the owner).
type Member struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
}

type Artist struct {
	Handle        string `json:"handle"`
	DisplayName   string `json:"display_name"`
	Bio           string `json:"bio"`
	OwnerUserID   string `json:"owner_user_id"`
	CreatedAt     string `json:"created_at"`
	FollowerCount int    `json:"follower_count"`
}

type service struct {
	store       *Store
	memberStore *MemberStore
}

func NewService(store *Store, memberStore *MemberStore) Service {
	return &service{store: store, memberStore: memberStore}
}

func normalizeHandle(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (s *service) Create(ctx context.Context, ownerUserID, handle, displayName, bio string) (*Artist, error) {
	handle = normalizeHandle(handle)
	displayName = strings.TrimSpace(displayName)
	bio = strings.TrimSpace(bio)
	if handle == "" {
		return nil, fmt.Errorf("handle required")
	}
	if !handleRegex.MatchString(handle) {
		return nil, fmt.Errorf("handle must be 4–64 lowercase letters or numbers, no spaces or special characters")
	}
	if displayName == "" {
		displayName = handle
	}

	existing, err := s.store.GetByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrHandleTaken
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	if err := s.store.Create(ctx, handle, displayName, bio, ownerUserID, createdAt); err != nil {
		if dynamo.IsCondCheckFailed(err) {
			return nil, ErrHandleTaken
		}
		return nil, err
	}
	// Store owner in member table so ownership can be transferred in the future (no API to change it yet).
	if err := s.memberStore.Put(ctx, handle, ownerUserID, []string{RoleOwner}); err != nil {
		return nil, err
	}
	return &Artist{
		Handle:        handle,
		DisplayName:   displayName,
		Bio:           bio,
		OwnerUserID:   ownerUserID,
		CreatedAt:     createdAt,
		FollowerCount: 0,
	}, nil
}

func (s *service) GetByHandle(ctx context.Context, handle string) (*Artist, error) {
	handle = normalizeHandle(handle)
	if handle == "" {
		return nil, ErrArtistNotFound
	}
	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return nil, ErrArtistNotFound
	}
	return rowToArtist(row), nil
}

func (s *service) ListByOwner(ctx context.Context, userID string) ([]Artist, error) {
	if userID == "" {
		return nil, nil
	}
	rows, err := s.store.ListByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Artist, len(rows))
	for i := range rows {
		out[i] = *rowToArtist(&rows[i])
	}
	return out, nil
}

func (s *service) Update(ctx context.Context, handle string, displayName, bio *string, actorUserID string) (*Artist, error) {
	handle = normalizeHandle(handle)
	if handle == "" {
		return nil, ErrArtistNotFound
	}

	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return nil, ErrArtistNotFound
	}
	ok, err := s.HasPermission(ctx, handle, actorUserID, PermArtistUpdate)
	if err != nil || !ok {
		return nil, ErrForbidden
	}

	resolvedDisplayName := row.DisplayName
	if displayName != nil {
		resolvedDisplayName = strings.TrimSpace(*displayName)
		if resolvedDisplayName == "" {
			resolvedDisplayName = row.DisplayName
		}
	}
	resolvedBio := row.Bio
	if bio != nil {
		resolvedBio = strings.TrimSpace(*bio)
	}

	if err := s.store.Update(ctx, handle, resolvedDisplayName, resolvedBio); err != nil {
		return nil, err
	}
	return &Artist{
		Handle:      handle,
		DisplayName: resolvedDisplayName,
		Bio:         resolvedBio,
		OwnerUserID: row.OwnerUserID,
		CreatedAt:   row.CreatedAt,
	}, nil
}

func (s *service) Delete(ctx context.Context, handle, actorUserID string) error {
	handle = normalizeHandle(handle)
	if handle == "" {
		return ErrArtistNotFound
	}

	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return ErrArtistNotFound
	}
	ok, err := s.HasPermission(ctx, handle, actorUserID, PermArtistDelete)
	if err != nil || !ok {
		return ErrForbidden
	}
	// Remove owner's member row so artist delete is consistent.
	if err := s.memberStore.Delete(ctx, handle, row.OwnerUserID); err != nil {
		return err
	}
	return s.store.Delete(ctx, handle)
}

func rowToArtist(r *artistRow) *Artist {
	if r == nil {
		return nil
	}
	return &Artist{
		Handle:        r.Handle,
		DisplayName:   r.DisplayName,
		Bio:           r.Bio,
		OwnerUserID:   r.OwnerUserID,
		CreatedAt:     r.CreatedAt,
		FollowerCount: r.FollowerCount,
	}
}

func (s *service) HasPermission(ctx context.Context, handle, userID, permission string) (bool, error) {
	handle = normalizeHandle(handle)
	if handle == "" || userID == "" {
		return false, nil
	}
	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return false, err
	}
	if row.OwnerUserID == userID {
		return true, nil
	}
	mem, err := s.memberStore.Get(ctx, handle, userID)
	if err != nil || mem == nil {
		return false, err
	}
	return RolesGrantPermission(mem.Roles, permission), nil
}

func (s *service) ListForUser(ctx context.Context, userID string) ([]ArtistWithRole, error) {
	if userID == "" {
		return nil, nil
	}
	var out []ArtistWithRole
	owned, err := s.store.ListByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}
	ownedHandles := make(map[string]bool)
	for i := range owned {
		out = append(out, ArtistWithRole{Artist: *rowToArtist(&owned[i]), Role: "owner"})
		ownedHandles[owned[i].Handle] = true
	}
	memberships, err := s.memberStore.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, m := range memberships {
		if ownedHandles[m.Handle] {
			continue
		}
		artist, err := s.GetByHandle(ctx, m.Handle)
		if err != nil || artist == nil {
			continue
		}
		out = append(out, ArtistWithRole{Artist: *artist, Role: "member", Roles: m.Roles})
	}
	return out, nil
}

func dedupeRoles(roles []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, r := range roles {
		if !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}

func (s *service) AddMember(ctx context.Context, handle, userID string, roles []string, actorUserID string) error {
	handle = normalizeHandle(handle)
	if handle == "" {
		return ErrArtistNotFound
	}
	ok, err := s.HasPermission(ctx, handle, actorUserID, PermArtistManageMembers)
	if err != nil || !ok {
		return ErrForbidden
	}
	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return ErrArtistNotFound
	}
	if row.OwnerUserID == userID {
		return errors.New("user is already the owner")
	}
	roles = dedupeRoles(roles)
	if len(roles) == 0 {
		return ErrInvalidRoles
	}
	if !RolesAreAssignable(roles) {
		return ErrInvalidRoles
	}
	return s.memberStore.Put(ctx, handle, userID, roles)
}

func (s *service) RemoveMember(ctx context.Context, handle, userID string, actorUserID string) error {
	handle = normalizeHandle(handle)
	if handle == "" {
		return ErrArtistNotFound
	}
	ok, err := s.HasPermission(ctx, handle, actorUserID, PermArtistManageMembers)
	if err != nil || !ok {
		return ErrForbidden
	}
	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return ErrArtistNotFound
	}
	if row.OwnerUserID == userID {
		return ErrCannotRemoveOwner
	}
	return s.memberStore.Delete(ctx, handle, userID)
}

func (s *service) UpdateMemberRoles(ctx context.Context, handle, userID string, roles []string, actorUserID string) error {
	handle = normalizeHandle(handle)
	if handle == "" {
		return ErrArtistNotFound
	}
	ok, err := s.HasPermission(ctx, handle, actorUserID, PermArtistManageMembers)
	if err != nil || !ok {
		return ErrForbidden
	}
	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return ErrArtistNotFound
	}
	if row.OwnerUserID == userID {
		return errors.New("cannot change owner roles")
	}
	roles = dedupeRoles(roles)
	if len(roles) == 0 {
		return s.memberStore.Delete(ctx, handle, userID)
	}
	if !RolesAreAssignable(roles) {
		return ErrInvalidRoles
	}
	return s.memberStore.Put(ctx, handle, userID, roles)
}

func (s *service) ListMembers(ctx context.Context, handle, actorUserID string) ([]Member, error) {
	handle = normalizeHandle(handle)
	if handle == "" {
		return nil, ErrArtistNotFound
	}
	ok, err := s.HasPermission(ctx, handle, actorUserID, PermArtistListMembers)
	if err != nil || !ok {
		return nil, ErrForbidden
	}
	row, err := s.store.GetByHandle(ctx, handle)
	if err != nil || row == nil {
		return nil, ErrArtistNotFound
	}
	rows, err := s.memberStore.ListByArtist(ctx, handle)
	if err != nil {
		return nil, err
	}
	hasOwner := false
	for i := range rows {
		if rows[i].UserID == row.OwnerUserID {
			hasOwner = true
			break
		}
	}
	out := make([]Member, 0, len(rows)+1)
	if !hasOwner {
		out = append(out, Member{UserID: row.OwnerUserID, Roles: []string{RoleOwner}})
	}
	for i := range rows {
		out = append(out, Member{UserID: rows[i].UserID, Roles: rows[i].Roles})
	}
	return out, nil
}
