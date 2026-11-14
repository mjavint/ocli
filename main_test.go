package main

import "testing"

func TestMainPackage(t *testing.T) {
	// Main is now a thin wrapper over internal.Execute; keep package compilable.
	t.Log("main package structure valid")
}
