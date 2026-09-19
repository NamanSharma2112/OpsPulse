package domain

import "time"

// User is a person who signs in to OpsPulse, either with a password or
// through GitHub OAuth.
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"-"`
	GitHubID     *int64 `json:"github_id,omitempty"`
	GitHubLogin  string `json:"github_login,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	// GitHubAccessToken is stored sealed and never serialised.
	GitHubAccessToken string     `json:"-"`
	GitHubTokenScopes string     `json:"github_token_scopes,omitempty"`
	GitHubConnectedAt *time.Time `json:"github_connected_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// HasGitHubToken reports whether OpsPulse can call GitHub as this user.
func (u User) HasGitHubToken() bool { return u.GitHubAccessToken != "" }
