package main

import "testing"

func TestMainCompiles(t *testing.T) {
	// The real main() boot path requires a live PostgreSQL instance and is
	// covered by integration/e2e tests. This unit-level test asserts the
	// cmd package compiles and its imports resolve.
}
