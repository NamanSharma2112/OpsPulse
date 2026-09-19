// Package auth issues and verifies the session tokens the dashboard uses.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Claims is the payload carried in a session token.
type Claims struct {
	UserID    string `json:"sub"`
	Email     string `json:"email"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// Errors returned while verifying a token.
var (
	ErrMalformedToken = errors.New("malformed token")
	ErrBadSignature   = errors.New("invalid token signature")
	ErrTokenExpired   = errors.New("token expired")
)

// TokenIssuer mints and validates HS256 JSON Web Tokens.
type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenIssuer returns an issuer signing with the given secret.
func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), ttl: ttl}
}

// TTL is how long freshly issued tokens stay valid.
func (t *TokenIssuer) TTL() time.Duration { return t.ttl }

// Issue signs a token for the given user.
func (t *TokenIssuer) Issue(userID, email string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(t.ttl).Unix(),
	}
	header, err := encodeSegment(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := encodeSegment(claims)
	if err != nil {
		return "", err
	}
	signing := header + "." + payload
	return signing + "." + t.sign(signing), nil
}

// Verify checks the signature and expiry, returning the embedded claims.
func (t *TokenIssuer) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrMalformedToken
	}
	if !hmac.Equal([]byte(t.sign(parts[0]+"."+parts[1])), []byte(parts[2])) {
		return nil, ErrBadSignature
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrMalformedToken
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, ErrMalformedToken
	}
	if time.Now().UTC().Unix() >= claims.ExpiresAt {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

func (t *TokenIssuer) sign(signing string) string {
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(signing))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeSegment(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
