package domain

import "time"

// Project is a unit of work inside an organization. It watches one or more
// repositories; the repositories themselves are separate records.
type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
