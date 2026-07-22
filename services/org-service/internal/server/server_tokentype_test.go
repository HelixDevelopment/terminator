package server_test

// HEL-001 — Fleet-wide JWT tokenType confusion (org-service leg).
//
// auth-service signs BOTH access tokens (tokenType "access") AND refresh
// tokens (tokenType "refresh", much longer TTL, no orgId/role/permissions)
// with the SAME Ed25519 private key
// (services/auth-service/internal/crypto/crypto.go). A refresh token is
// therefore signature-valid here. Before the HEL-001 fix, this service's
// validateToken checked only the signing method, token.Valid
// (signature + exp) and the claims type assertion — and its Claims struct
// did not even declare a TokenType field, so a stolen or ALREADY-REVOKED
// refresh token authenticated to every /api/v1 org/team route for its
// full (long) lifetime, bypassing auth-service's /refresh revocation seam.
//
// §11.4.10: keys are generated in-test; no signed token value is ever
// logged — assertions compare status codes only.

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/helixdevelopment/org-service/internal/server"
)

// hel001ProbeRoute is behind authMiddleware, so a non-401 answer proves
// the middleware ADMITTED the caller.
const hel001ProbeRoute = "/api/v1/orgs"

// hel001Token mints a token whose wire claims replicate what auth-service
// really emits for the given tokenType, signed with the TRUSTED test key
// so the signature genuinely verifies. jwt.MapClaims is used deliberately
// rather than server.Claims: the attack is defined by the JSON on the
// wire, which keeps this test valid independently of the service's
// internal claims struct.
func hel001Token(t *testing.T, tokenType string) string {
	t.Helper()
	now := time.Now().UTC()
	ttl := time.Hour
	if tokenType == "refresh" {
		ttl = 720 * time.Hour // refresh tokens live far longer — the point of the attack
	}
	claims := jwt.MapClaims{
		"userId":    "11111111-1111-1111-1111-111111111111",
		"sessionId": "33333333-3333-3333-3333-333333333333",
		"tokenType": tokenType,
		"exp":       jwt.NewNumericDate(now.Add(ttl)),
		"iat":       jwt.NewNumericDate(now),
		"sub":       "11111111-1111-1111-1111-111111111111",
		"iss":       "helixterminator",
	}
	if tokenType == "access" {
		// Only access tokens carry tenant/authorisation claims.
		claims["orgId"] = "22222222-2222-2222-2222-222222222222"
		claims["email"] = "test@example.com"
		claims["role"] = "user"
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(testPrivateKey)
	require.NoError(t, err)
	return tokenString
}

// TestHEL001_RefreshTokenRejectedOnAccessPath is the HEL-001 regression
// guard: FAILED on the pre-fix artifact (refresh token admitted), PASSES
// on the fixed one.
func TestHEL001_RefreshTokenRejectedOnAccessPath(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(testPublicKey))
	srv, err := server.New(nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, hel001ProbeRoute, nil)
	req.Header.Set("Authorization", "Bearer "+hel001Token(t, "refresh"))
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"a refresh-typed token (tokenType=\"refresh\") is signature-valid but MUST NOT "+
			"authenticate on the access-token path; anything other than 401 means the "+
			"refresh token was admitted (HEL-001)")
}

// TestHEL001_AccessTokenStillAccepted is the positive control needle
// (§11.4.201(1)): it proves the guard is SPECIFIC and does not simply
// reject every token. A guard that rejected everything would satisfy the
// negative test above while breaking all real authentication — a
// false-positive refusal, a FAIL-bluff of the same severity as the
// original defect.
func TestHEL001_AccessTokenStillAccepted(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(testPublicKey))
	srv, err := server.New(nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, hel001ProbeRoute, nil)
	req.Header.Set("Authorization", "Bearer "+hel001Token(t, "access"))
	srv.Router().ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code,
		"a genuine access token (tokenType=\"access\") MUST still authenticate; "+
			"a 401 here would mean the HEL-001 guard over-rejects and has broken "+
			"real authentication")
}
