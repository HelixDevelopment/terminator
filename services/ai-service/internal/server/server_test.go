package server_test

import (
	"testing"
	"time"

	"github.com/helixdevelopment/ai-service/internal/handler"
	"github.com/helixdevelopment/ai-service/internal/server"
)

func TestServerStubRemoved(t *testing.T) {
	// Re-export the existing real timeout invariant tests already exercise the
	// server package. This test simply confirms the package compiles and the
	// exported constructor can be called.
	h := handler.New(nil, nil)
	_ = h
	_ = server.MinWriteTimeoutMargin
	_ = server.ResolveHTTPWriteTimeout()
	_ = handler.ResolveLLMTimeout()
}

func TestResolveHTTPWriteTimeoutDefaults(t *testing.T) {
	// Sanity check the default write timeout is reasonable.
	d := server.ResolveHTTPWriteTimeout()
	if d <= 0 {
		t.Fatalf("expected positive default write timeout, got %s", d)
	}
}

func TestTimeoutInvariantHoldsForDefaults(t *testing.T) {
	write := server.ResolveHTTPWriteTimeout()
	llm := handler.ResolveLLMTimeout()
	if write <= llm {
		t.Fatalf("default write timeout %s must exceed LLM timeout %s", write, llm)
	}
	if margin := write - llm; margin < server.MinWriteTimeoutMargin {
		t.Fatalf("margin %s < MinWriteTimeoutMargin %s", margin, server.MinWriteTimeoutMargin)
	}
}

func TestValidateTimeoutInvariantAcceptsValid(t *testing.T) {
	if err := server.ValidateTimeoutInvariant(150*time.Second, 90*time.Second); err != nil {
		t.Fatalf("expected valid timeout pair to be accepted: %v", err)
	}
}

func TestValidateTimeoutInvariantRejectsInvalid(t *testing.T) {
	if err := server.ValidateTimeoutInvariant(30*time.Second, 45*time.Second); err == nil {
		t.Fatal("expected invalid timeout pair (LLM > write) to be rejected")
	}
}
