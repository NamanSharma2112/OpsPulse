package domain

import "time"

// Roles a user can hold inside an organization.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Organization groups projects and the people who can see them. It optionally
// mirrors a GitHub organization.
type Organization struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	GitHubLogin string    `json:"github_login,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Membership ties a user to an organization with a role.
type Membership struct {
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
}
