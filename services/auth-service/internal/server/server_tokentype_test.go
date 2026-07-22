package server_test

// HEL-001 — Fleet-wide JWT tokenType confusion (auth-service's OWN legs).
//
// The fleet-wide sweep for HEL-001 initially enumerated the eight
// downstream JWT-validating services. auth-service itself turned out to
// hold TWO more validation paths that admitted a refresh token
// (§11.4.118 discovery pressure — the reported set was not the complete
// set):
//
//	(1) server.jwtValidationMiddleware, which protects auth-service's own
//	    authenticated group (/logout, /mfa/setup, /me). It called
//	    jwtManager.ValidateToken and never inspected TokenType.
//
//	    Its IsAccessTokenActive revocation check resolves to
//	    repo.GetSessionByTokenHash, whose SQL matches
//	    `WHERE access_token_hash = $1` — a refresh token's hash misses,
//	    so in a DB-backed deployment the refresh token was rejected
//	    INCIDENTALLY. But IsAccessTokenActive opens with
//	    `if h.repo == nil { return true }`, so in the no-database
//	    configuration the revocation check FAILS OPEN and the refresh
//	    token was genuinely admitted. These tests run exactly that
//	    no-database configuration — the vulnerable one.
//
//	(2) handler.ValidateToken (POST /validate), which answered
//	    `{"valid":true, userId, orgId, permissions, expiresAt}` for ANY
//	    signature-valid token, refresh tokens included, with no
//	    incidental protection of any kind. The response is an
//	    authentication oracle: a caller asking "is this a usable access
//	    token?" was told yes for a refresh token.
//
// Both tokens here are minted by the REAL production minting code
// (srv.JWTManager().GenerateRefreshToken / GenerateAccessToken) rather
// than hand-assembled claims, so the fixtures cannot drift from what
// auth-service actually issues.
//
// §11.4.10: the keypair is generated in-test and no signed token value
// is ever logged — assertions compare status codes and the boolean
// `valid` field only.

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/helixdevelopment/auth-service/internal/server"
)

// newHEL001Server builds a real auth-service with a real provisioned
// Ed25519 keypair and NO database — the configuration in which
// IsAccessTokenActive fails open, isolating the tokenType check as the
// only thing that can reject a refresh token.
func newHEL001Server(t *testing.T) *server.Server {
	t.Helper()
	clearJWTEnv(t)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey (test-only keypair) failed: %v", err)
	}
	t.Setenv("JWT_PRIVATE_KEY", base64.StdEncoding.EncodeToString(priv))
	t.Setenv("JWT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(pub))

	srv, err := server.New(&recordingLogger{})
	if err != nil {
		t.Fatalf("server.New with a valid provisioned keypair failed: %v", err)
	}
	return srv
}

// TestHEL001_AuthService_RefreshTokenRejectedOnProtectedRoute is the
// regression guard for leg (1). Pre-fix this FAILED: the refresh token
// was admitted to /me and answered 200.
func TestHEL001_AuthService_RefreshTokenRejectedOnProtectedRoute(t *testing.T) {
	srv := newHEL001Server(t)

	refreshToken, _, err := srv.JWTManager().GenerateRefreshToken("hel001-user", "hel001-session")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("a refresh-typed token authenticated to the protected route /me: got status %d, want 401. "+
			"A refresh token is signature-valid but MUST NOT authenticate on an access-token path (HEL-001)", w.Code)
	}
}

// TestHEL001_AuthService_AccessTokenStillAcceptedOnProtectedRoute is the
// positive control needle (§11.4.201(1)): it proves the guard is
// SPECIFIC and has not simply broken authentication for everyone. A
// guard that rejected every token would satisfy the negative test above
// while making auth-service unusable — a false-positive refusal, which
// is a FAIL-bluff of the same severity as the original defect.
func TestHEL001_AuthService_AccessTokenStillAcceptedOnProtectedRoute(t *testing.T) {
	srv := newHEL001Server(t)

	accessToken, _, err := srv.JWTManager().GenerateAccessToken(
		"hel001-user", "hel001-org", "hel001@example.com", "user", "hel001-session", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	srv.Router().ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("a genuine access token was rejected from /me (status 401) — the HEL-001 guard "+
			"over-rejects and has broken real authentication")
	}
}

// TestHEL001_ValidateEndpoint_RejectsRefreshToken is the regression guard
// for leg (2). Pre-fix this FAILED: POST /validate answered
// {"valid":true} for a refresh token, telling every caller that a
// refresh token was a usable access token.
func TestHEL001_ValidateEndpoint_RejectsRefreshToken(t *testing.T) {
	srv := newHEL001Server(t)

	refreshToken, _, err := srv.JWTManager().GenerateRefreshToken("hel001-user", "hel001-session")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if valid := postValidate(t, srv, refreshToken); valid {
		t.Fatal("POST /validate reported a refresh token as valid. The endpoint returns userId/orgId/" +
			"permissions and is used as an authentication oracle, so it MUST NOT answer valid=true " +
			"for a non-access token (HEL-001)")
	}
}

// TestHEL001_ValidateEndpoint_AcceptsAccessToken is the positive control
// needle for leg (2): /validate must still confirm genuine access tokens.
func TestHEL001_ValidateEndpoint_AcceptsAccessToken(t *testing.T) {
	srv := newHEL001Server(t)

	accessToken, _, err := srv.JWTManager().GenerateAccessToken(
		"hel001-user", "hel001-org", "hel001@example.com", "user", "hel001-session", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	if valid := postValidate(t, srv, accessToken); !valid {
		t.Fatal("POST /validate reported a genuine access token as INVALID — the HEL-001 guard " +
			"over-rejects and has broken token validation")
	}
}

// postValidate posts a token to /validate and returns the response's
// `valid` field. Only that boolean is read; the token itself is never
// logged (§11.4.10).
func postValidate(t *testing.T, srv *server.Server, token string) bool {
	t.Helper()

	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		t.Fatalf("json.Marshal(validate request) failed: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST /validate returned status %d, want 200 (the endpoint reports validity in its "+
			"body, not via the status code)", w.Code)
	}

	var resp struct {
		Valid bool `json:"valid"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding /validate response failed: %v", err)
	}
	return resp.Valid
}
