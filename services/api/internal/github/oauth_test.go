package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthorizeURL(t *testing.T) {
	cfg := OAuthConfig{
		ClientID:      "client-123",
		RedirectURL:   "https://ops.example.com/v1/auth/github/callback",
		AuthorizeBase: "https://github.com",
		Scopes:        []string{"read:user", "admin:repo_hook"},
	}

	raw := cfg.AuthorizeURL("state-abc")
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("AuthorizeURL() produced an unparseable URL: %v", err)
	}
	if parsed.Path != "/login/oauth/authorize" {
		t.Fatalf("path = %q, want /login/oauth/authorize", parsed.Path)
	}
	q := parsed.Query()
	if q.Get("client_id") != "client-123" {
		t.Errorf("client_id = %q", q.Get("client_id"))
	}
	if q.Get("state") != "state-abc" {
		t.Errorf("state = %q", q.Get("state"))
	}
	if q.Get("scope") != "read:user admin:repo_hook" {
		t.Errorf("scope = %q", q.Get("scope"))
	}
	if q.Get("redirect_uri") != cfg.RedirectURL {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
}

func TestNewStateIsUnguessable(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		s, err := NewState()
		if err != nil {
			t.Fatalf("NewState() error: %v", err)
		}
		if len(s) < 40 {
			t.Fatalf("state is too short to be unguessable: %q", s)
		}
		if seen[s] {
			t.Fatalf("NewState() repeated a value: %q", s)
		}
		seen[s] = true
	}
}

func TestExchange(t *testing.T) {
	var gotForm url.Values
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login/oauth/access_token" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}
		_ = r.ParseForm()
		gotForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Token{
			AccessToken: "gho_test", TokenType: "bearer", Scope: "repo,admin:repo_hook",
		})
	}))
	defer stub.Close()

	cfg := OAuthConfig{ClientID: "id", ClientSecret: "secret", AuthorizeBase: stub.URL}
	token, err := cfg.Exchange(context.Background(), "the-code")
	if err != nil {
		t.Fatalf("Exchange() error: %v", err)
	}
	if token.AccessToken != "gho_test" {
		t.Fatalf("AccessToken = %q", token.AccessToken)
	}
	if gotForm.Get("code") != "the-code" || gotForm.Get("client_secret") != "secret" {
		t.Fatalf("form = %v", gotForm)
	}
}

func TestExchangeSurfacesGitHubError(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Token{
			Error: "bad_verification_code", ErrorDesc: "The code passed is incorrect or expired.",
		})
	}))
	defer stub.Close()

	cfg := OAuthConfig{ClientID: "id", ClientSecret: "secret", AuthorizeBase: stub.URL}
	_, err := cfg.Exchange(context.Background(), "stale")
	if err == nil {
		t.Fatal("Exchange() accepted a rejected code")
	}
	if !strings.Contains(err.Error(), "incorrect or expired") {
		t.Fatalf("error did not carry GitHub's reason: %v", err)
	}
}

func TestConfigured(t *testing.T) {
	if (OAuthConfig{}).Configured() {
		t.Error("empty config reported as configured")
	}
	if (OAuthConfig{ClientID: "a"}).Configured() {
		t.Error("config without a secret reported as configured")
	}
	if !(OAuthConfig{ClientID: "a", ClientSecret: "b"}).Configured() {
		t.Error("complete config reported as unconfigured")
	}
}
