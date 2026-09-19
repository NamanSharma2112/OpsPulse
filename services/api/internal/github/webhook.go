// Package github verifies and decodes GitHub webhook deliveries.
package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// Errors returned while validating a delivery.
var (
	ErrMissingSignature = errors.New("missing X-Hub-Signature-256 header")
	ErrBadSignature     = errors.New("webhook signature mismatch")
)

// VerifySignature checks the sha256 HMAC GitHub sends with every delivery.
// The header looks like "sha256=<hex digest>".
func VerifySignature(secret string, body []byte, header string) error {
	if header == "" {
		return ErrMissingSignature
	}
	digest, ok := strings.CutPrefix(header, "sha256=")
	if !ok {
		return ErrBadSignature
	}
	sum, err := hex.DecodeString(digest)
	if err != nil {
		return ErrBadSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	if !hmac.Equal(sum, mac.Sum(nil)) {
		return ErrBadSignature
	}
	return nil
}
