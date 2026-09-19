package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// SecretBox seals short secrets — GitHub access tokens — before they are
// written to the database, so a dump on its own does not hand an attacker
// working credentials to somebody's repositories.
//
// AES-256-GCM, with the nonce prefixed to the ciphertext. The key lives in
// the environment rather than the database, which is what makes the two
// separable in the first place.
type SecretBox struct{ aead cipher.AEAD }

// ErrNoEncryptionKey reports that token encryption was not configured.
var ErrNoEncryptionKey = errors.New("token encryption key is not configured")

// NewSecretBox builds a box from a 32-byte key, hex encoded.
func NewSecretBox(hexKey string) (*SecretBox, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes (64 hex characters), got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	return &SecretBox{aead: aead}, nil
}

// Seal encrypts plaintext and returns it base64 encoded. An empty input
// returns an empty string so callers can store "no token" without a branch.
func (b *SecretBox) Seal(plaintext string) (string, error) {
	if b == nil {
		return "", ErrNoEncryptionKey
	}
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	sealed := b.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Open reverses Seal.
func (b *SecretBox) Open(encoded string) (string, error) {
	if b == nil {
		return "", ErrNoEncryptionKey
	}
	if encoded == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode sealed value: %w", err)
	}
	if len(raw) < b.aead.NonceSize() {
		return "", errors.New("sealed value is too short")
	}
	nonce, ciphertext := raw[:b.aead.NonceSize()], raw[b.aead.NonceSize():]
	plaintext, err := b.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// Wrong key, or the value was tampered with. Both are the same
		// answer to the caller.
		return "", fmt.Errorf("open sealed value: %w", err)
	}
	return string(plaintext), nil
}
