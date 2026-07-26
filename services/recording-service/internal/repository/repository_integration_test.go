//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/recording-service/internal/model"
	"github.com/helixdevelopment/recording-service/internal/repository"
	"github.com/helixdevelopment/recording-service/internal/testutil"
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

func TestCreateAndGetRecording_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	recording := &model.Recording{
		ID:            uuid.New(),
		SessionID:     uuid.New(),
		HostID:        uuid.New(),
		UserID:        uuid.New(),
		FilePath:      "/recordings/test.cast",
		Format:        model.RecordingFormatAsciinema,
		Status:        model.RecordingStatusCompleted,
		DurationSec:   120,
		FileSizeBytes: 1024,
	}
	if err := repo.CreateRecording(ctx, recording); err != nil {
		t.Fatalf("CreateRecording failed: %v", err)
	}

	got, err := repo.GetRecordingByID(ctx, recording.ID)
	if err != nil {
		t.Fatalf("GetRecordingByID failed: %v", err)
	}
	if got.ID != recording.ID || got.FilePath != recording.FilePath {
		t.Fatalf("recording round-trip mismatch: got %+v", got)
	}
}
