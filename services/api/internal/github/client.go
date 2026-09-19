package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client calls the GitHub REST API as a signed-in user.
type Client struct {
	token   string
	apiBase string
}

// NewClient returns a client authenticated with a user access token.
func NewClient(token, apiBase string) *Client {
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	return &Client{token: token, apiBase: strings.TrimRight(apiBase, "/")}
}

// User is the subset of GitHub's user object OpsPulse stores.
type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// Repo is the subset of a repository OpsPulse shows when picking one.
type Repo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
	HTMLURL       string `json:"html_url"`
	Permissions   struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
	} `json:"permissions"`
}

// CurrentUser returns the account the token belongs to.
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	var user User
	if err := c.do(ctx, http.MethodGet, "/user", nil, &user); err != nil {
		return nil, err
	}
	// A private primary address is hidden from /user, so ask separately. It
	// is not fatal if this fails — the caller can synthesise an address.
	if user.Email == "" {
		user.Email = c.primaryEmail(ctx)
	}
	return &user, nil
}

type emailEntry struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (c *Client) primaryEmail(ctx context.Context) string {
	var entries []emailEntry
	if err := c.do(ctx, http.MethodGet, "/user/emails", nil, &entries); err != nil {
		return ""
	}
	for _, e := range entries {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	return ""
}

// ListRepositories returns the repositories the user can administer, which is
// the set OpsPulse can install a webhook on.
func (c *Client) ListRepositories(ctx context.Context) ([]Repo, error) {
	out := []Repo{}
	// Two pages is plenty for picking a repository; the list is a chooser,
	// not an inventory.
	for page := 1; page <= 2; page++ {
		var batch []Repo
		path := fmt.Sprintf("/user/repos?per_page=100&sort=pushed&affiliation=owner,collaborator,organization_member&page=%d", page)
		if err := c.do(ctx, http.MethodGet, path, nil, &batch); err != nil {
			return nil, err
		}
		for _, repo := range batch {
			if repo.Permissions.Admin {
				out = append(out, repo)
			}
		}
		if len(batch) < 100 {
			break
		}
	}
	return out, nil
}

// WebhookRequest is the hook OpsPulse installs on a repository.
type WebhookRequest struct {
	URL    string
	Secret string
	Events []string
}

type hookPayload struct {
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Events []string `json:"events"`
	Config struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		Secret      string `json:"secret"`
		InsecureSSL string `json:"insecure_ssl"`
	} `json:"config"`
}

// Hook is GitHub's response to installing a webhook.
type Hook struct {
	ID     int64  `json:"id"`
	Active bool   `json:"active"`
	URL    string `json:"url"`
}

// DefaultWebhookEvents are the deliveries OpsPulse knows how to project.
var DefaultWebhookEvents = []string{
	"push", "pull_request", "deployment", "deployment_status", "workflow_run", "issues",
}

// CreateWebhook installs a webhook on a repository and returns GitHub's id
// for it.
func (c *Client) CreateWebhook(ctx context.Context, owner, repo string, in WebhookRequest) (*Hook, error) {
	var body hookPayload
	body.Name = "web"
	body.Active = true
	body.Events = in.Events
	body.Config.URL = in.URL
	body.Config.ContentType = "json"
	body.Config.Secret = in.Secret
	body.Config.InsecureSSL = "0"

	var hook Hook
	path := fmt.Sprintf("/repos/%s/%s/hooks", owner, repo)
	if err := c.do(ctx, http.MethodPost, path, body, &hook); err != nil {
		return nil, err
	}
	return &hook, nil
}

// DeleteWebhook removes a webhook OpsPulse installed.
func (c *Client) DeleteWebhook(ctx context.Context, owner, repo string, hookID int64) error {
	path := fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, hookID)
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// APIError carries GitHub's status and message so callers can tell a missing
// permission apart from a genuine failure.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("github api: %d %s", e.Status, e.Message)
}

func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		encoded, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiBase+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("call github: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read github response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var wrapped struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(raw, &wrapped)
		return &APIError{Status: res.StatusCode, Message: firstNonEmpty(wrapped.Message, http.StatusText(res.StatusCode))}
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}
	return nil
}
