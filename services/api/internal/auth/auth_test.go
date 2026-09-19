package auth

import (
	"strings"
	"testing"
	"time"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if !strings.HasPrefix(hash, "pbkdf2_sha256$") {
		t.Fatalf("unexpected hash format: %q", hash)
	}

	ok, err := VerifyPassword("correct horse battery staple", hash)
	if err != nil || !ok {
		t.Fatalf("VerifyPassword(correct) = %v, %v", ok, err)
	}
	ok, err = VerifyPassword("wrong password", hash)
	if err != nil || ok {
		t.Fatalf("VerifyPassword(wrong) = %v, %v", ok, err)
	}
}

func TestHashPasswordUsesFreshSalt(t *testing.T) {
	a, _ := HashPassword("same")
	b, _ := HashPassword("same")
	if a == b {
		t.Fatal("identical passwords produced identical hashes; salt is not random")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	for _, bad := range []string{"", "plaintext", "bcrypt$1$2$3", "pbkdf2_sha256$0$a$b"} {
		if _, err := VerifyPassword("x", bad); err == nil {
			t.Fatalf("VerifyPassword(%q) accepted a malformed hash", bad)
		}
	}
}

func TestTokenRoundTrip(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Hour)
	token, err := issuer.Issue("user-1", "dev@example.com")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}
	claims, err := issuer.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "dev@example.com" {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestVerifyRejectsForeignSignature(t *testing.T) {
	token, err := NewTokenIssuer("secret-a", time.Hour).Issue("user-1", "dev@example.com")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}
	if _, err := NewTokenIssuer("secret-b", time.Hour).Verify(token); err != ErrBadSignature {
		t.Fatalf("Verify() with wrong secret = %v, want ErrBadSignature", err)
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	issuer := NewTokenIssuer("secret", -time.Minute) // already expired when issued
	token, err := issuer.Issue("user-1", "dev@example.com")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}
	if _, err := issuer.Verify(token); err != ErrTokenExpired {
		t.Fatalf("Verify() = %v, want ErrTokenExpired", err)
	}
}

func TestVerifyRejectsMalformedToken(t *testing.T) {
	issuer := NewTokenIssuer("secret", time.Hour)
	for _, bad := range []string{"", "a.b", "a.b.c.d"} {
		if _, err := issuer.Verify(bad); err != ErrMalformedToken {
			t.Fatalf("Verify(%q) = %v, want ErrMalformedToken", bad, err)
		}
	}
}
