package main

import (
	"testing"
)

func TestMainCompiles(t *testing.T) {
	// The real main() boot path requires a live PostgreSQL instance and is
	// exercised by integration/e2e tests. This unit-level test asserts the cmd
	// package still compiles and its imports resolve.
}
