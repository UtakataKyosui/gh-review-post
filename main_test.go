package main_test

import (
	"os/exec"
	"testing"
)

// TestBinaryBuilds verifies the project compiles cleanly.
func TestBinaryBuilds(t *testing.T) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
}
