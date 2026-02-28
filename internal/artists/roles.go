package artists

// Predefined role names. A user on an artist page can have multiple roles.
// RoleOwner is stored in the member table so ownership can be transferred in the future;
// it is not assignable via API (only set at artist create).
const (
	RoleOwner  = "owner"  // Page creator; stored for future transfer; not assignable via API
	RoleAdmin  = "admin"  // Can manage members, update artist settings; cannot delete page or billing
	RoleFeed   = "feed"   // Create, edit, delete feed posts
	RoleMusic  = "music"  // Upload and manage music
	RolePhotos = "photos" // Upload and manage photos
	RoleGigs   = "gigs"   // Add, edit, delete gigs
)

// AllPredefinedRoles returns roles that can be assigned via API (excludes owner).
func AllPredefinedRoles() []string {
	return []string{RoleAdmin, RoleFeed, RoleMusic, RolePhotos, RoleGigs}
}

// Permission constants for authorization checks.
const (
	PermArtistUpdate        = "artist:update"
	PermArtistDelete        = "artist:delete"
	PermArtistManageMembers = "artist:manage_members"
	PermArtistListMembers   = "artist:list_members" // Any page member can list; only owner/admin can add/update/remove
	PermFeedCreate          = "feed:create"
	PermFeedUpdate        = "feed:update"
	PermFeedDelete        = "feed:delete"
	PermMusicManage       = "music:manage"
	PermPhotosManage      = "photos:manage"
	PermGigsManage        = "gigs:manage"
)

// rolePermissions maps each role to the permissions it grants.
var rolePermissions = map[string][]string{
	RoleOwner: {
		PermArtistUpdate, PermArtistManageMembers, PermArtistListMembers,
		PermFeedCreate, PermFeedUpdate, PermFeedDelete,
		PermMusicManage, PermPhotosManage, PermGigsManage,
		// PermArtistDelete only for owner; not in map so only artist.OwnerUserID check grants it
	},
	RoleAdmin: {
		PermArtistUpdate, PermArtistManageMembers, PermArtistListMembers,
		PermFeedCreate, PermFeedUpdate, PermFeedDelete,
		PermMusicManage, PermPhotosManage, PermGigsManage,
	},
	RoleFeed:   {PermFeedCreate, PermFeedUpdate, PermFeedDelete, PermArtistListMembers},
	RoleMusic:  {PermMusicManage, PermArtistListMembers},
	RolePhotos: {PermPhotosManage, PermArtistListMembers},
	RoleGigs:   {PermGigsManage, PermArtistListMembers},
}

// RoleGrantsPermission returns true if the given role grants the given permission.
func RoleGrantsPermission(role, permission string) bool {
	for _, p := range rolePermissions[role] {
		if p == permission {
			return true
		}
	}
	return false
}

// RolesGrantPermission returns true if any of the given roles grants the given permission.
func RolesGrantPermission(roles []string, permission string) bool {
	for _, r := range roles {
		if RoleGrantsPermission(r, permission) {
			return true
		}
	}
	return false
}

// ValidRole returns true if role is a predefined role (including owner, for storage).
func ValidRole(role string) bool {
	if role == RoleOwner {
		return true
	}
	for _, r := range AllPredefinedRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// AssignableRole returns true if the role can be set via AddMember/UpdateMemberRoles (excludes owner).
func AssignableRole(role string) bool {
	for _, r := range AllPredefinedRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// ValidRoles returns true if every role in the slice is predefined. Duplicates are allowed in input (we dedupe when saving).
func ValidRoles(roles []string) bool {
	for _, r := range roles {
		if !ValidRole(r) {
			return false
		}
	}
	return true
}

// RolesAreAssignable returns true if all roles can be set via API (no owner).
func RolesAreAssignable(roles []string) bool {
	for _, r := range roles {
		if !AssignableRole(r) {
			return false
		}
	}
	return true
}
