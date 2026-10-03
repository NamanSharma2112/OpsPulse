package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
)

// These exercise the four branches of UpsertGitHub against a real database,
// because the behaviour under test is which unique index a write collides
// with — something no fake can tell you.
func testDB(t *testing.T) *DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	db, err := Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)
	if _, err := db.Pool.Exec(context.Background(), `DELETE FROM users`); err != nil {
		t.Fatalf("clean users: %v", err)
	}
	return db
}

func githubUser(id int64, email, login string) *domain.User {
	return &domain.User{
		Email: email, Name: login, GitHubID: &id, GitHubLogin: login,
		GitHubAccessToken: "sealed-" + login, GitHubTokenScopes: "repo",
	}
}

func TestUpsertGitHubCreatesAccount(t *testing.T) {
	db := testDB(t)
	u := githubUser(1, "new@example.com", "newbie")
	if err := db.Users.UpsertGitHub(context.Background(), u); err != nil {
		t.Fatalf("UpsertGitHub() error: %v", err)
	}
	if u.ID == "" {
		t.Fatal("no id returned")
	}
}

func TestUpsertGitHubRefreshesReturningUser(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)

	first := githubUser(2, "back@example.com", "oldname")
	if err := db.Users.UpsertGitHub(ctx, first); err != nil {
		t.Fatalf("first sign-in: %v", err)
	}

	again := githubUser(2, "back@example.com", "newname")
	if err := db.Users.UpsertGitHub(ctx, again); err != nil {
		t.Fatalf("second sign-in: %v", err)
	}
	if again.ID != first.ID {
		t.Fatalf("a second row was created: %s then %s", first.ID, again.ID)
	}

	stored, err := db.Users.GetByID(ctx, first.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.GitHubLogin != "newname" {
		t.Fatalf("login = %q, want the refreshed one", stored.GitHubLogin)
	}
}

// The case that broke sign-in: a password account already holds the address
// GitHub reports.
func TestUpsertGitHubAdoptsPasswordAccount(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)

	existing := &domain.User{
		Email: "both@example.com", Name: "Both", PasswordHash: "pbkdf2_sha256$1$a$b",
	}
	if err := db.Users.Create(ctx, existing); err != nil {
		t.Fatalf("create password account: %v", err)
	}

	linked := githubUser(3, "both@example.com", "bothcat")
	if err := db.Users.UpsertGitHub(ctx, linked); err != nil {
		t.Fatalf("UpsertGitHub() on an existing address: %v", err)
	}
	if linked.ID != existing.ID {
		t.Fatalf("made a second account (%s) instead of linking to %s", linked.ID, existing.ID)
	}

	stored, err := db.Users.GetByID(ctx, existing.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.GitHubID == nil || *stored.GitHubID != 3 {
		t.Fatal("the GitHub identity was not attached")
	}
	if stored.PasswordHash == "" {
		t.Fatal("the password was dropped; the user can no longer sign in with it")
	}
}

// Linking here would hand one person another's data.
func TestUpsertGitHubRefusesAddressOwnedByAnotherAccount(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)

	if err := db.Users.UpsertGitHub(ctx, githubUser(4, "taken@example.com", "first")); err != nil {
		t.Fatalf("first account: %v", err)
	}

	err := db.Users.UpsertGitHub(ctx, githubUser(5, "taken@example.com", "second"))
	if err == nil {
		t.Fatal("a second GitHub account claimed an address it does not own")
	}
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v, want domain.ErrConflict so the caller can explain it", err)
	}
}
