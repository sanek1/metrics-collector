package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintBuildInfo(t *testing.T) {
	originalBuildVersion := buildVersion
	originalBuildDate := buildDate
	originalBuildCommit := buildCommit

	// Restore original values after test
	defer func() {
		buildVersion = originalBuildVersion
		buildDate = originalBuildDate
		buildCommit = originalBuildCommit
	}()

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create temporary file for redirecting output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test 1: set values
	buildVersion = "test-version"
	buildDate = "test-date"
	buildCommit = "test-commit"

	printBuildInfo()

	// Close writer and read output
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Build version: test-version")
	assert.Contains(t, output, "Build date: test-date")
	assert.Contains(t, output, "Build commit: test-commit")

	// Test 2: empty values - create new pipe
	r, w, _ = os.Pipe()
	os.Stdout = w

	buildVersion = ""
	buildDate = ""
	buildCommit = ""

	printBuildInfo()

	// Close writer and read output
	w.Close()
	buf.Reset()
	_, _ = buf.ReadFrom(r)
	output = buf.String()

	assert.Contains(t, output, "Build version: N/A")
	assert.Contains(t, output, "Build date: N/A")
	assert.Contains(t, output, "Build commit: N/A")
}
