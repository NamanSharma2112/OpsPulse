// Command seed fills a local OpsPulse with plausible activity.
//
// Running the product without a GitHub OAuth app leaves you with an empty
// dashboard and no obvious way to see what it does: every panel reads a
// projection, and every projection is built from deliveries that only arrive
// once a real repository is wired to a public address. This walks the same
// path GitHub would — register, create an organization and a project, connect
// a repository, then sign and POST deliveries to /v1/webhooks/github — so the
// dashboard fills with data that went through signature verification,
// deduplication and the projection pipeline rather than being written behind
// them.
//
//	go run ./cmd/seed                 # against http://localhost:8080
//	go run ./cmd/seed -days 60        # a longer history
//
// It prints the session token and project id to put in .env.
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

type options struct {
	api      string
	email    string
	password string
	org      string
	project  string
	repo     string
	branch   string
	days     int
	seed     int64
}

func run() error {
	var o options
	flag.StringVar(&o.api, "api", envOr("OPSPULSE_API_URL", "http://localhost:8080"), "API base URL")
	flag.StringVar(&o.email, "email", "dev@opspulse.local", "account to create or reuse")
	flag.StringVar(&o.password, "password", "seed-password-please-change", "account password")
	flag.StringVar(&o.org, "org", "Acme", "organization name")
	flag.StringVar(&o.project, "project", "Checkout", "project name")
	flag.StringVar(&o.repo, "repo", "acme/checkout", "repository as owner/name")
	flag.StringVar(&o.branch, "branch", "main", "default branch")
	flag.IntVar(&o.days, "days", 30, "days of history to generate")
	flag.Int64Var(&o.seed, "seed", 1, "random seed; the same seed gives the same history")
	flag.Parse()

	if o.days < 1 {
		return fmt.Errorf("-days must be at least 1")
	}
	o.api = strings.TrimRight(o.api, "/")

	c := &client{base: o.api, http: &http.Client{Timeout: 30 * time.Second}}

	token, err := c.signIn(o.email, o.password)
	if err != nil {
		return err
	}
	c.token = token

	orgID, err := c.ensureOrganization(o.org)
	if err != nil {
		return err
	}
	projectID, err := c.ensureProject(orgID, o.project)
	if err != nil {
		return err
	}
	secret, err := c.connectRepository(projectID, o.repo, o.branch)
	if err != nil {
		return err
	}

	owner, name, ok := strings.Cut(o.repo, "/")
	if !ok {
		return fmt.Errorf("-repo must look like owner/name, got %q", o.repo)
	}

	gen := &generator{
		rnd:    rand.New(rand.NewSource(o.seed)),
		owner:  owner,
		name:   name,
		branch: o.branch,
		days:   o.days,
	}
	deliveries := gen.history()

	var accepted, duplicate int
	for _, d := range deliveries {
		dup, err := c.deliver(secret, d)
		if err != nil {
			return fmt.Errorf("deliver %s: %w", d.event, err)
		}
		if dup {
			duplicate++
			continue
		}
		accepted++
	}

	fmt.Printf("seeded %d deliveries over %d days (%d accepted, %d already present)\n",
		len(deliveries), o.days, accepted, duplicate)
	fmt.Printf("\nPut these in .env, then restart the dashboard:\n\n")
	fmt.Printf("  OPSPULSE_API_TOKEN=%s\n", token)
	fmt.Printf("  OPSPULSE_PROJECT_ID=%s\n\n", projectID)
	fmt.Printf("Or sign in at http://localhost:3000/signin as %s\n", o.email)
	return nil
}

// client talks to the API as one signed-in user.
type client struct {
	base  string
	token string
	http  *http.Client
}

// signIn registers the account, falling back to logging in when it already
// exists, so the command is safe to run repeatedly.
func (c *client) signIn(email, password string) (string, error) {
	var session struct {
		Token string `json:"token"`
	}
	name, _, _ := strings.Cut(email, "@")
	err := c.do("POST", "/v1/auth/register",
		map[string]string{"email": email, "name": name, "password": password}, &session)
	if err == nil {
		return session.Token, nil
	}
	var status statusError
	if !asStatus(err, &status) || status.code != http.StatusConflict {
		return "", fmt.Errorf("register: %w", err)
	}
	if err := c.do("POST", "/v1/auth/login",
		map[string]string{"email": email, "password": password}, &session); err != nil {
		return "", fmt.Errorf("log in as the existing %s: %w", email, err)
	}
	return session.Token, nil
}

func (c *client) ensureOrganization(name string) (string, error) {
	var existing struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
	}
	if err := c.do("GET", "/v1/organizations", nil, &existing); err != nil {
		return "", fmt.Errorf("list organizations: %w", err)
	}
	for _, o := range existing.Organizations {
		if o.Name == name {
			return o.ID, nil
		}
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := c.do("POST", "/v1/organizations", map[string]string{"name": name}, &created); err != nil {
		return "", fmt.Errorf("create organization: %w", err)
	}
	return created.ID, nil
}

func (c *client) ensureProject(orgID, name string) (string, error) {
	var existing struct {
		Projects []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := c.do("GET", "/v1/projects", nil, &existing); err != nil {
		return "", fmt.Errorf("list projects: %w", err)
	}
	for _, p := range existing.Projects {
		if p.Name == name {
			return p.ID, nil
		}
	}
	var created struct {
		ID string `json:"id"`
	}
	body := map[string]string{"organization_id": orgID, "name": name}
	if err := c.do("POST", "/v1/projects", body, &created); err != nil {
		return "", fmt.Errorf("create project: %w", err)
	}
	return created.ID, nil
}

// connectRepository returns the repository's webhook secret, which the API
// serves exactly once — on the call that creates it. A repository connected by
// an earlier run cannot be signed for, so that is reported rather than
// producing deliveries the API would reject.
func (c *client) connectRepository(projectID, repo, branch string) (string, error) {
	var result struct {
		WebhookSecret string `json:"webhook_secret"`
	}
	body := map[string]string{"repo": repo, "default_branch": branch}
	err := c.do("POST", "/v1/projects/"+projectID+"/repositories", body, &result)
	if err != nil {
		var status statusError
		if asStatus(err, &status) && status.code == http.StatusConflict {
			return "", fmt.Errorf("%s is already connected, and its webhook secret is "+
				"only ever served once. Seed a different repository with -repo, or "+
				"start from an empty database (make reset)", repo)
		}
		return "", fmt.Errorf("connect %s: %w", repo, err)
	}
	if result.WebhookSecret == "" {
		return "", fmt.Errorf("the API connected %s but returned no webhook secret", repo)
	}
	return result.WebhookSecret, nil
}

// deliver signs one payload the way GitHub does and posts it.
func (c *client) deliver(secret string, d delivery) (bool, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(d.body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest("POST", c.base+"/v1/webhooks/github", bytes.NewReader(d.body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", d.event)
	req.Header.Set("X-GitHub-Delivery", d.id)
	req.Header.Set("X-Hub-Signature-256", signature)

	res, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	payload, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return false, statusError{code: res.StatusCode, body: strings.TrimSpace(string(payload))}
	}
	var result struct {
		Duplicate bool `json:"duplicate"`
	}
	_ = json.Unmarshal(payload, &result)
	return result.Duplicate, nil
}

func (c *client) do(method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w (is the API running?)", method, path, err)
	}
	defer res.Body.Close()
	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		return statusError{code: res.StatusCode, body: strings.TrimSpace(string(payload))}
	}
	if out == nil || len(payload) == 0 {
		return nil
	}
	return json.Unmarshal(payload, out)
}

// statusError carries the status code so callers can branch on "already
// exists" without matching on message text.
type statusError struct {
	code int
	body string
}

func (e statusError) Error() string {
	return fmt.Sprintf("status %d: %s", e.code, e.body)
}

func asStatus(err error, target *statusError) bool {
	s, ok := err.(statusError)
	if ok {
		*target = s
	}
	return ok
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
