//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/sftp-service/internal/model"
	"github.com/helixdevelopment/sftp-service/internal/repository"
	"github.com/helixdevelopment/sftp-service/internal/testutil"
)

func newRealRepo(t *testing.T) *repository.Repository {
	t.Helper()
	poolURL, available := testutil.StartTestPostgres(t)
	if !available {
		t.Skip("SKIP: podman not available — cannot run repository tests against a real database")
	}
	pool, err := pgxpool.New(context.Background(), poolURL)
	if err != nil {
		t.Fatalf("pgxpool.New failed: %v", err)
	}
	t.Cleanup(pool.Close)
	return repository.New(pool)
}

func TestCreateAndGetSession_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	session := &model.SFTPSession{
		ID:               uuid.New(),
		HostID:           uuid.New(),
		UserID:           uuid.New(),
		RemotePath:       "/remote/test.txt",
		LocalPath:        "/local/test.txt",
		Direction:        model.SFTPDirectionUpload,
		Status:           model.SFTPSessionStatusPending,
		BytesTransferred: 0,
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	got, err := repo.GetSessionByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSessionByID failed: %v", err)
	}
	if got.ID != session.ID || got.RemotePath != session.RemotePath {
		t.Fatalf("session round-trip mismatch: got %+v", got)
	}
}
