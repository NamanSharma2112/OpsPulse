package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Parameters for password derivation. Kept in the encoded hash so they can be
// raised later without invalidating existing credentials.
const (
	pbkdf2Iterations = 210_000
	pbkdf2KeyLen     = 32
	saltLen          = 16
)

// ErrInvalidHash is returned when a stored hash cannot be parsed.
var ErrInvalidHash = errors.New("invalid password hash")

// HashPassword derives a salted PBKDF2-HMAC-SHA256 hash, encoded as
// pbkdf2_sha256$iterations$salt$key.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, pbkdf2Iterations, pbkdf2KeyLen)
	if err != nil {
		return "", fmt.Errorf("derive key: %w", err)
	}
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s",
		pbkdf2Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword reports whether the password matches the encoded hash. The
// comparison is constant time.
func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false, ErrInvalidHash
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false, ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, ErrInvalidHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, ErrInvalidHash
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(want))
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
