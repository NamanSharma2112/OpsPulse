package domain

import "time"

// Roles a user can hold inside an organisation.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Org groups projects and the people who can see them. It optionally mirrors
// a GitHub organisation.
type Org struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	GitHubLogin string    `json:"github_login,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Membership ties a user to an org with a role.
type Membership struct {
	OrgID     string    `json:"org_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
