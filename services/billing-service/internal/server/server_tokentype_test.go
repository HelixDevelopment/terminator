package server_test

// HEL-001 — Fleet-wide JWT tokenType confusion (billing-service leg).
//
// auth-service signs BOTH access tokens (tokenType "access") AND refresh
// tokens (tokenType "refresh", much longer TTL) with the SAME Ed25519
// private key (services/auth-service/internal/crypto/crypto.go), so a
// refresh token is signature-valid here and validateToken admits it.
//
// HONEST SCOPE OF THIS LEG (§11.4.6 / §11.4.194 — each factor verified
// independently): unlike the other seven JWT-validating services,
// billing-service was NOT exploitable by a refresh token before the fix.
// Its authMiddleware performs a SECOND check after validateToken —
// `if claims.OrgID == "" { 403 "token missing org identity" }` — and
// auth-service's GenerateRefreshToken happens to omit orgId, so a
// refresh token was already rejected with 403. That protection is
// INCIDENTAL, not a token-type control: it depends entirely on refresh
// tokens continuing to omit orgId. The moment a refresh token carried an
// orgId claim, billing-service would admit it. This leg is therefore
// DEFENCE IN DEPTH placed at the correct layer (the token's declared
// type), turning an accidental 403 into a deliberate 401.
//
// §11.4.10: keys are generated in-test; no signed token value is ever
// logged — assertions compare status codes only.

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/helixdevelopment/billing-service/internal/server"
)

// hel001ProbeRoute is behind authMiddleware, so a non-401 answer proves
// the middleware did not reject the caller on token type.
const hel001ProbeRoute = "/api/v1/subscriptions"

// This service has no shared non-integration test keypair, so HEL-001
// generates its own. Never a committed key (§11.4.10).
var hel001PubKey ed25519.PublicKey
var hel001PrivKey ed25519.PrivateKey

func init() {
	var err error
	hel001PubKey, hel001PrivKey, err = ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
}

// hel001Token mints a token whose wire claims replicate what auth-service
// really emits for the given tokenType, signed with the TRUSTED test key
// so the signature genuinely verifies.
//
// NOTE the deliberate orgId choice below: the refresh token here CARRIES
// an orgId. A faithful auth-service refresh token omits it, but omitting
// it would let billing's pre-existing `OrgID == ""` 403 mask the result,
// so this test would pass for a reason that has nothing to do with token
// type — the exact §11.4.201 false-null this test must avoid. Supplying
// orgId isolates the tokenType check as the ONLY thing that can reject
// this request, and simultaneously models the forward-looking case in
// which refresh tokens gain an orgId claim.
func hel001Token(t *testing.T, tokenType string) string {
	t.Helper()
	now := time.Now().UTC()
	ttl := time.Hour
	if tokenType == "refresh" {
		ttl = 720 * time.Hour // refresh tokens live far longer — the point of the attack
	}
	claims := jwt.MapClaims{
		"userId":    "11111111-1111-1111-1111-111111111111",
		"orgId":     "22222222-2222-2222-2222-222222222222",
		"sessionId": "33333333-3333-3333-3333-333333333333",
		"tokenType": tokenType,
		"exp":       jwt.NewNumericDate(now.Add(ttl)),
		"iat":       jwt.NewNumericDate(now),
		"sub":       "11111111-1111-1111-1111-111111111111",
		"iss":       "helixterminator",
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(hel001PrivKey)
	require.NoError(t, err)
	return tokenString
}

// TestHEL001_RefreshTokenRejectedOnAccessPath is the HEL-001 regression
// guard. Pre-fix this FAILED: an org-carrying refresh token cleared both
// validateToken and the OrgID check and was ADMITTED.
func TestHEL001_RefreshTokenRejectedOnAccessPath(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(hel001PubKey))
	srv := server.New(nil)

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
	t.Setenv("JWT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(hel001PubKey))
	srv := server.New(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, hel001ProbeRoute, nil)
	req.Header.Set("Authorization", "Bearer "+hel001Token(t, "access"))
	srv.Router().ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code,
		"a genuine access token (tokenType=\"access\") MUST still authenticate; "+
			"a 401 here would mean the HEL-001 guard over-rejects and has broken "+
			"real authentication")
}
