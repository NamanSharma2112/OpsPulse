package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateWebhook(t *testing.T) {
	var got hookPayload
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/opspulse/hooks" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer gho_test" {
			t.Errorf("Authorization = %q", auth)
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Hook{ID: 4242, Active: true})
	}))
	defer stub.Close()

	hook, err := NewClient("gho_test", stub.URL).CreateWebhook(
		context.Background(), "acme", "opspulse",
		WebhookRequest{URL: "https://ops.example.com/v1/webhooks/github", Secret: "s3cret", Events: DefaultWebhookEvents},
	)
	if err != nil {
		t.Fatalf("CreateWebhook() error: %v", err)
	}
	if hook.ID != 4242 {
		t.Fatalf("hook id = %d, want 4242", hook.ID)
	}
	if got.Config.Secret != "s3cret" {
		t.Errorf("secret was not sent to GitHub: %+v", got.Config)
	}
	if got.Config.ContentType != "json" {
		t.Errorf("content_type = %q, want json", got.Config.ContentType)
	}
	if got.Config.InsecureSSL != "0" {
		t.Errorf("insecure_ssl = %q, want 0", got.Config.InsecureSSL)
	}
	if !got.Active {
		t.Error("hook was not created active")
	}
}

func TestCreateWebhookSurfacesStatus(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Must have admin rights to Repository."})
	}))
	defer stub.Close()

	_, err := NewClient("gho_test", stub.URL).CreateWebhook(
		context.Background(), "acme", "opspulse", WebhookRequest{URL: "https://x", Secret: "y"})

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if apiErr.Status != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", apiErr.Status)
	}
}

func TestListRepositoriesKeepsOnlyAdministrable(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		repos := []map[string]any{
			{"id": 1, "name": "owned", "full_name": "acme/owned", "permissions": map[string]bool{"admin": true}},
			{"id": 2, "name": "readonly", "full_name": "acme/readonly", "permissions": map[string]bool{"admin": false}},
		}
		_ = json.NewEncoder(w).Encode(repos)
	}))
	defer stub.Close()

	repos, err := NewClient("gho_test", stub.URL).ListRepositories(context.Background())
	if err != nil {
		t.Fatalf("ListRepositories() error: %v", err)
	}
	if len(repos) != 1 || repos[0].FullName != "acme/owned" {
		t.Fatalf("repos = %+v; want only the administrable one", repos)
	}
}

func TestCurrentUserFallsBackToPrimaryEmail(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			// GitHub hides a private address here.
			_ = json.NewEncoder(w).Encode(User{ID: 7, Login: "octocat", Name: "Octo"})
		case "/user/emails":
			_ = json.NewEncoder(w).Encode([]emailEntry{
				{Email: "unverified@example.com", Primary: false, Verified: false},
				{Email: "octo@example.com", Primary: true, Verified: true},
			})
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer stub.Close()

	user, err := NewClient("gho_test", stub.URL).CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("CurrentUser() error: %v", err)
	}
	if user.Email != "octo@example.com" {
		t.Fatalf("email = %q, want the verified primary address", user.Email)
	}
}
