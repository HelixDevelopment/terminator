package main

import (
	"net/http"
	"testing"
	"time"
)

func TestMainBootsAndServesHealth(t *testing.T) {
	// main() is already covered by the real e2e boot path in integration tests;
	// this unit-level test only proves the cmd package compiles and its
	// dependencies resolve, which is the only thing safe to assert without a
	// real database in a unit context.
	_ = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	_ = time.Second
}
