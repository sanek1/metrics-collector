package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type testTable struct {
	url    string
	body   string
	want   string
	status int
}

func TestPrintBuildInfo(t *testing.T) {
	originalBuildVersion := buildVersion
	originalBuildDate := buildDate
	originalBuildCommit := buildCommit

	defer func() {
		buildVersion = originalBuildVersion
		buildDate = originalBuildDate
		buildCommit = originalBuildCommit
	}()

	buildVersion = "test-version"
	buildDate = "test-date"
	buildCommit = "test-commit"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	printBuildInfo()

	w.Close()
	out, _ := io.ReadAll(r)
	output := string(out)

	require.Contains(t, output, "Build version: test-version")
	require.Contains(t, output, "Build date: test-date")
	require.Contains(t, output, "Build commit: test-commit")

	buildVersion = ""
	buildDate = ""
	buildCommit = ""

	r, w, _ = os.Pipe()
	os.Stdout = w

	printBuildInfo()

	w.Close()
	out, _ = io.ReadAll(r)
	output = string(out)

	require.Contains(t, output, "Build version: N/A")
	require.Contains(t, output, "Build date: N/A")
	require.Contains(t, output, "Build commit: N/A")
}
