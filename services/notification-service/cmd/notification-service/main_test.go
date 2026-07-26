package main

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestMainGracefulShutdown(t *testing.T) {
	// Save and restore the original env to avoid side effects.
	origPort := os.Getenv("PORT")
	defer func() {
		_ = os.Setenv("PORT", origPort)
	}()
	_ = os.Setenv("PORT", "18081")

	// Start the server in a goroutine.
	started := make(chan struct{}, 1)
	go func() {
		started <- struct{}{}
		main()
	}()

	<-started
	// Give main() time to bind the listener.
	time.Sleep(100 * time.Millisecond)

	// The server should be reachable.
	client := &http.Client{Timeout: 500 * time.Millisecond}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:18081/healthz", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("server not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected health status: %d", resp.StatusCode)
	}

	// Trigger graceful shutdown.
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	// Allow shutdown to complete.
	time.Sleep(200 * time.Millisecond)
}
