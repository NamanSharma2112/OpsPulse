package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	body := []byte(`{"action":"opened"}`)
	const secret = "s3cret"

	tests := []struct {
		name   string
		header string
		want   error
	}{
		{"valid", sign(secret, body), nil},
		{"missing", "", ErrMissingSignature},
		{"wrong secret", sign("other", body), ErrBadSignature},
		{"no prefix", hex.EncodeToString([]byte("abc")), ErrBadSignature},
		{"not hex", "sha256=zzzz", ErrBadSignature},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := VerifySignature(secret, body, tc.header); !errors.Is(err, tc.want) {
				t.Fatalf("VerifySignature() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestVerifySignatureRejectsTamperedBody(t *testing.T) {
	const secret = "s3cret"
	header := sign(secret, []byte(`{"action":"opened"}`))
	if err := VerifySignature(secret, []byte(`{"action":"closed"}`), header); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("tampered body accepted: %v", err)
	}
}

func TestParsePayload(t *testing.T) {
	body := []byte(`{
		"action": "closed",
		"repository": {"name": "opspulse", "full_name": "acme/opspulse", "owner": {"login": "acme"}},
		"sender": {"login": "octocat"},
		"pull_request": {"number": 7, "title": "Add ingest", "state": "closed", "merged": true,
			"user": {"login": "octocat"}, "created_at": "2026-01-01T00:00:00Z",
			"merged_at": "2026-01-02T00:00:00Z"}
	}`)

	p, err := ParsePayload(body)
	if err != nil {
		t.Fatalf("ParsePayload() error: %v", err)
	}
	if p.Repository.Owner.Login != "acme" || p.Repository.Name != "opspulse" {
		t.Fatalf("repository = %q/%q", p.Repository.Owner.Login, p.Repository.Name)
	}
	if p.PullRequest == nil || p.PullRequest.Number != 7 || !p.PullRequest.Merged {
		t.Fatalf("pull request not parsed: %+v", p.PullRequest)
	}
	if p.PullRequest.MergedAt == nil {
		t.Fatal("merged_at not parsed")
	}
}
