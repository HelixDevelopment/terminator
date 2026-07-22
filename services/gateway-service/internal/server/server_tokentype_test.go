package server_test

// HEL-001 — Fleet-wide JWT tokenType confusion.
//
// auth-service signs BOTH access tokens (tokenType "access", short TTL,
// carries orgId/role/permissions) AND refresh tokens (tokenType "refresh",
// long TTL, carries neither) with the SAME Ed25519 private key
// (services/auth-service/internal/crypto/crypto.go —
// GenerateAccessTokenWithExpiry sets TokenType:"access",
// GenerateRefreshToken sets TokenType:"refresh", both
// token.SignedString(m.privateKey)).
//
// A refresh token is therefore signature-valid at every JWT-validating
// service in the fleet. Before the HEL-001 fix, gateway-service's
// validateToken checked only (a) the signing method is Ed25519, (b)
// token.Valid (signature + exp) and (c) the claims type assertion — it
// never read Claims.TokenType even though the field is declared. A
// stolen OR ALREADY-REVOKED refresh token thus authenticated directly to
// any route behind jwtValidationMiddleware for its much longer lifetime,
// completely bypassing the /refresh endpoint's hash-check + revocation
// seam.
//
// The probe route is GET /api/v1/users/me: it sits behind
// api.Use(s.jwtValidationMiddleware()) but its handler is a local
// notImplemented (501) — no upstream proxying — so the response code
// isolates the AUTH MIDDLEWARE decision rather than any backend's
// availability.
//
// §11.4.10: every key here is generated in-test; no signed token value is
// ever logged — assertions compare status codes only, never token bytes.

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// protectedProbeRoute is behind jwtValidationMiddleware and served by a
// local 501 handler, so a non-401 answer proves the middleware ADMITTED
// the caller.
const protectedProbeRoute = "/api/v1/users/me"

// generateRefreshTypedToken mints a token whose wire claims replicate
// auth-service's GenerateRefreshToken EXACTLY (tokenType "refresh",
// no orgId/role/permissions, jti == sessionId, issuer "helixterminator"),
// signed with the trusted test key so the signature genuinely verifies.
// jwt.MapClaims is used deliberately rather than server.Claims: the
// attack is defined by the JSON on the wire, and this keeps the test
// independent of any one service's internal claims struct.
func generateRefreshTypedToken(t *testing.T) string {
	t.Helper()
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"userId":    "test-user-id",
		"email":     "",
		"role":      "",
		"sessionId": "test-session-id",
		"tokenType": "refresh",
		"exp":       jwt.NewNumericDate(now.Add(720 * time.Hour)),
		"iat":       jwt.NewNumericDate(now),
		"sub":       "test-user-id",
		"iss":       "helixterminator",
		"jti":       "test-session-id",
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(testPrivateKey)
	require.NoError(t, err)
	return tokenString
}

// TestHEL001_RefreshTokenRejectedOnAccessPath is the HEL-001 regression
// guard. It FAILED on the pre-fix artifact (the refresh token was
// admitted and the probe route answered 501) and PASSES on the fixed one.
func TestHEL001_RefreshTokenRejectedOnAccessPath(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(testPublicKey))
	s := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, protectedProbeRoute, nil)
	req.Header.Set("Authorization", "Bearer "+generateRefreshTypedToken(t))
	s.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"a refresh-typed token (tokenType=\"refresh\") is signature-valid but MUST NOT "+
			"authenticate on the access-token path; anything other than 401 means the "+
			"refresh token was admitted (HEL-001)")
}

// TestHEL001_AccessTokenStillAccepted is the positive control needle
// (§11.4.201(1)): it proves the tokenType guard is SPECIFIC and does not
// simply reject every token. Without this control, a guard that rejected
// everything would satisfy the negative test above while breaking all
// real authentication — a false-positive refusal, which is a FAIL-bluff
// of exactly the same severity as the original defect.
func TestHEL001_AccessTokenStillAccepted(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(testPublicKey))
	s := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, protectedProbeRoute, nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken())
	s.Router().ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code,
		"a genuine access token (tokenType=\"access\") MUST still authenticate; "+
			"a 401 here would mean the HEL-001 guard over-rejects and has broken "+
			"real authentication")
}
