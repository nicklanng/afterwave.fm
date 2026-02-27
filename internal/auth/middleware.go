package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey struct{}
type sessionKey struct{}

// Authenticate validates the session token (from Cookie or Authorization Bearer) with RS256 and sets the user ID (sub) and session ID (jti) in the request context.
func Authenticate(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := SessionTokenFromRequest(r)
			if tokenString == "" {
				http.Error(w, "missing or invalid authorization", http.StatusUnauthorized)
				return
			}
			tok, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
				if t.Method != jwt.SigningMethodRS256 {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return publicKey, nil
			}, jwt.WithExpirationRequired())
			if err != nil || !tok.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			claims, ok := tok.Claims.(*jwt.RegisteredClaims)
			if !ok || claims.Subject == "" {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextKey{}, claims.Subject)
			ctx = context.WithValue(ctx, sessionKey{}, claims.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext returns the authenticated user ID from the request context, or "" if not set.
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

// SessionIDFromContext returns the session ID (jti) from the request context, or "" if not set.
func SessionIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(sessionKey{}).(string)
	return v
}
