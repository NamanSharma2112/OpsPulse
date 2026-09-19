package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuthConfig describes the GitHub OAuth app OpsPulse signs users in with.
//
// The base URLs are configurable so the flow can be pointed at a stub in
// tests, and at a GitHub Enterprise host in a self-hosted install.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	// AuthorizeBase is where the user is sent, e.g. https://github.com.
	AuthorizeBase string
	// APIBase is the REST host, e.g. https://api.github.com.
	APIBase string
	Scopes  []string
}

// Configured reports whether sign-in with GitHub is available.
func (c OAuthConfig) Configured() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// NewState returns an unguessable value for the OAuth state parameter, which
// is what stops a third party from completing a login into someone's session.
func NewState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// AuthorizeURL is where the browser is sent to begin the flow.
func (c OAuthConfig) AuthorizeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", c.RedirectURL)
	q.Set("scope", strings.Join(c.Scopes, " "))
	q.Set("state", state)
	// Ask every time rather than silently reusing a prior grant, so the user
	// can see which scopes they are handing over.
	q.Set("allow_signup", "true")
	return strings.TrimRight(c.AuthorizeBase, "/") + "/login/oauth/authorize?" + q.Encode()
}

// Token is what GitHub returns from the code exchange.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// Exchange trades the callback's code for an access token.
func (c OAuthConfig) Exchange(ctx context.Context, code string) (*Token, error) {
	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", c.RedirectURL)

	endpoint := strings.TrimRight(c.AuthorizeBase, "/") + "/login/oauth/access_token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange code: github returned %d", res.StatusCode)
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if token.Error != "" {
		return nil, fmt.Errorf("github rejected the code: %s", firstNonEmpty(token.ErrorDesc, token.Error))
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("github returned no access token")
	}
	return &token, nil
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
