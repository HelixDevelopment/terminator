//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/helixdevelopment/audit-service/internal/model"
	"github.com/helixdevelopment/audit-service/internal/repository"
	"github.com/helixdevelopment/audit-service/internal/testutil"
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

func TestCreateAndGetAuditLog_RoundTrips(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	log := &model.AuditLog{
		ID:           uuid.New(),
		Action:       model.ActionCreate,
		ResourceType: model.ResourceTypeUser,
		Timestamp:    time.Now().UTC(),
		Severity:     model.SeverityInfo,
		IPAddress:    "127.0.0.1",
	}
	if err := repo.CreateAuditLog(ctx, log); err != nil {
		t.Fatalf("CreateAuditLog failed: %v", err)
	}

	got, err := repo.GetAuditLogByID(ctx, log.ID)
	if err != nil {
		t.Fatalf("GetAuditLogByID failed: %v", err)
	}
	if got.ID != log.ID || got.Action != log.Action {
		t.Fatalf("audit log round-trip mismatch: got %+v", got)
	}
}

func TestListAuditLogs_HonorsLimit(t *testing.T) {
	repo := newRealRepo(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		log := &model.AuditLog{
			ID:           uuid.New(),
			Action:       model.ActionRead,
			ResourceType: model.ResourceTypeHost,
			Timestamp:    time.Now().UTC(),
			Severity:     model.SeverityWarning,
			IPAddress:    "127.0.0.1",
		}
		if err := repo.CreateAuditLog(ctx, log); err != nil {
			t.Fatalf("CreateAuditLog failed: %v", err)
		}
	}

	logs, total, err := repo.ListAuditLogs(ctx, nil, nil, model.ActionRead, model.ResourceTypeHost, model.SeverityWarning, nil, nil, 2, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs failed: %v", err)
	}
	if total < 5 {
		t.Fatalf("expected total >= 5, got %d", total)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs with limit=2, got %d", len(logs))
	}
}
