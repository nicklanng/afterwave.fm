package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strings"
)

const CodeChallengeMethodS256 = "S256"

// ComputeCodeChallenge returns the S256 code_challenge for a code_verifier (base64url(sha256(verifier))).
func ComputeCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// VerifyCodeVerifier checks that code_verifier produces the stored code_challenge (S256 only).
// Uses constant-time comparison to avoid timing leaks.
func VerifyCodeVerifier(codeVerifier, codeChallenge, method string) bool {
	method = strings.TrimSpace(method)
	if method == "" {
		method = CodeChallengeMethodS256
	}
	if method != CodeChallengeMethodS256 {
		return false
	}
	hash := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])
	computedB := []byte(computed)
	challengeB := []byte(codeChallenge)
	if len(computedB) != len(challengeB) {
		return false
	}
	return subtle.ConstantTimeCompare(computedB, challengeB) == 1
}
