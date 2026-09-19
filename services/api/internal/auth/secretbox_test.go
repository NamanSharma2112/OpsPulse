package auth

import (
	"strings"
	"testing"
)

const testKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func TestSecretBoxRoundTrip(t *testing.T) {
	box, err := NewSecretBox(testKey)
	if err != nil {
		t.Fatalf("NewSecretBox() error: %v", err)
	}

	const token = "gho_examplegithubtokenvalue"
	sealed, err := box.Seal(token)
	if err != nil {
		t.Fatalf("Seal() error: %v", err)
	}
	if strings.Contains(sealed, token) {
		t.Fatal("sealed value contains the plaintext token")
	}

	opened, err := box.Open(sealed)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	if opened != token {
		t.Fatalf("Open() = %q, want %q", opened, token)
	}
}

func TestSecretBoxUsesFreshNonce(t *testing.T) {
	box, _ := NewSecretBox(testKey)
	a, _ := box.Seal("same-token")
	b, _ := box.Seal("same-token")
	if a == b {
		t.Fatal("identical plaintexts sealed identically; nonce is not random")
	}
}

func TestSecretBoxRejectsForeignKey(t *testing.T) {
	sealed, _ := mustBox(t, testKey).Seal("secret")
	other := mustBox(t, strings.Repeat("ab", 32))
	if _, err := other.Open(sealed); err == nil {
		t.Fatal("a box with a different key opened the value")
	}
}

func TestSecretBoxRejectsTampering(t *testing.T) {
	box := mustBox(t, testKey)
	sealed, _ := box.Seal("secret")

	// Flip a character in the ciphertext body.
	corrupted := []byte(sealed)
	i := len(corrupted) - 2
	if corrupted[i] == 'A' {
		corrupted[i] = 'B'
	} else {
		corrupted[i] = 'A'
	}
	if _, err := box.Open(string(corrupted)); err == nil {
		t.Fatal("Open() accepted a tampered value")
	}
}

func TestSecretBoxEmptyValues(t *testing.T) {
	box := mustBox(t, testKey)
	sealed, err := box.Seal("")
	if err != nil || sealed != "" {
		t.Fatalf("Seal(\"\") = %q, %v; want empty, nil", sealed, err)
	}
	opened, err := box.Open("")
	if err != nil || opened != "" {
		t.Fatalf("Open(\"\") = %q, %v; want empty, nil", opened, err)
	}
}

func TestNewSecretBoxRejectsBadKeys(t *testing.T) {
	for _, key := range []string{"", "not-hex", strings.Repeat("00", 16)} {
		if _, err := NewSecretBox(key); err == nil {
			t.Fatalf("NewSecretBox(%q) accepted an invalid key", key)
		}
	}
}

func TestNilBoxIsSafe(t *testing.T) {
	var box *SecretBox
	if _, err := box.Seal("x"); err != ErrNoEncryptionKey {
		t.Fatalf("Seal on nil box = %v, want ErrNoEncryptionKey", err)
	}
	if _, err := box.Open("x"); err != ErrNoEncryptionKey {
		t.Fatalf("Open on nil box = %v, want ErrNoEncryptionKey", err)
	}
}

func mustBox(t *testing.T, key string) *SecretBox {
	t.Helper()
	box, err := NewSecretBox(key)
	if err != nil {
		t.Fatalf("NewSecretBox(%q) error: %v", key, err)
	}
	return box
}
