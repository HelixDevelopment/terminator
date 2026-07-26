package main

import "testing"

func TestMainCompiles(t *testing.T) {
	// The real main() boot path requires a live PostgreSQL instance and a
	// configured HelixTrack Core sandbox. It is covered by integration/e2e
	// tests and handler_live_test.go. This unit-level test asserts the cmd
	// package compiles and its imports resolve.
}
