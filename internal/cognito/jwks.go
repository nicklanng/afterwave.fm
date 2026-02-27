package cognito

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwksCacheTTL = 24 * time.Hour

// jwksCache caches Cognito JWKS by (region, userPoolID) with TTL.
type jwksCache struct {
	mu    sync.RWMutex
	entries map[string]jwksEntry
}

type jwksEntry struct {
	keys map[string]*rsa.PublicKey
	exp  time.Time
}

var defaultJWKSCache = &jwksCache{entries: make(map[string]jwksEntry)}

func jwksCacheKey(region, userPoolID string) string {
	return region + "|" + userPoolID
}

// jwksResponse is the JSON shape of Cognito's .well-known/jwks.json
type jwksResponse struct {
	Keys []struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func (c *jwksCache) getKeys(ctx context.Context, region, userPoolID string) (map[string]*rsa.PublicKey, error) {
	key := jwksCacheKey(region, userPoolID)
	c.mu.RLock()
	ent, ok := c.entries[key]
	c.mu.RUnlock()
	if ok && time.Now().Before(ent.exp) {
		return ent.keys, nil
	}

	url := "https://cognito-idp." + region + ".amazonaws.com/" + userPoolID + "/.well-known/jwks.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks: got status %d", resp.StatusCode)
	}
	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}
	keys := make(map[string]*rsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || k.N == "" || k.E == "" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		var e int
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		n := new(big.Int).SetBytes(nBytes)
		pk := &rsa.PublicKey{N: n, E: e}
		keys[k.Kid] = pk
	}

	c.mu.Lock()
	c.entries[key] = jwksEntry{keys: keys, exp: time.Now().Add(jwksCacheTTL)}
	c.mu.Unlock()
	return keys, nil
}

// ValidateIDToken verifies the Cognito ID token (signature, iss, aud, exp) and returns sub and email.
// region and userPoolID are used to fetch JWKS and validate issuer; clientID is the expected audience.
func ValidateIDToken(ctx context.Context, idToken, region, userPoolID, clientID string) (sub, email string, err error) {
	tok, err := jwt.Parse(idToken, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != "RS256" {
			return nil, fmt.Errorf("unexpected alg: %s", t.Method.Alg())
		}
		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, fmt.Errorf("missing kid")
		}
		keys, err := defaultJWKSCache.getKeys(ctx, region, userPoolID)
		if err != nil {
			return nil, err
		}
		k, ok := keys[kid]
		if !ok {
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}
		return k, nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return "", "", err
	}
	if !tok.Valid {
		return "", "", fmt.Errorf("invalid token")
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}
	iss := "https://cognito-idp." + region + ".amazonaws.com/" + userPoolID
	if claims["iss"] != iss {
		return "", "", fmt.Errorf("invalid iss")
	}
	switch v := claims["aud"].(type) {
	case string:
		if v != clientID {
			return "", "", fmt.Errorf("invalid aud")
		}
	case []interface{}:
		var found bool
		for _, a := range v {
			if s, _ := a.(string); s == clientID {
				found = true
				break
			}
		}
		if !found {
			return "", "", fmt.Errorf("invalid aud")
		}
	default:
		return "", "", fmt.Errorf("missing aud")
	}
	sub, _ = claims["sub"].(string)
	email, _ = claims["email"].(string)
	if sub == "" {
		return "", "", fmt.Errorf("missing sub")
	}
	return sub, email, nil
}
