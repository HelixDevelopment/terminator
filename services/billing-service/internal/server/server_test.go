package server_test

import (
	"testing"

	"github.com/helixdevelopment/billing-service/internal/repository"
	"github.com/helixdevelopment/billing-service/internal/server"
)

func TestServerCreation(t *testing.T) {
	repo := repository.New(nil)
	srv := server.New(repo)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestServerRouterExposed(t *testing.T) {
	repo := repository.New(nil)
	srv := server.New(repo)
	if srv.Router() == nil {
		t.Fatal("expected non-nil router")
	}
}
