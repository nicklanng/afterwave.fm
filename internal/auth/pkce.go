package auth

import (
	"crypto/sha256"
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
	return computed == codeChallenge
}
