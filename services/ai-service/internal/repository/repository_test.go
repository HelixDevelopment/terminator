package repository_test

import (
	"context"
	"testing"

	"github.com/helixdevelopment/ai-service/internal/repository"
)

func TestRepositoryPingWithoutPool(t *testing.T) {
	r := repository.New(nil)
	if err := r.Ping(context.Background()); err == nil {
		t.Fatal("expected error when pinging repository without a pool")
	}
}
