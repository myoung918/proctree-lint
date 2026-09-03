package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr redirects os.Stderr for the duration of fn and returns
// what was written to it. run() writes usage output there via the
// flag package's default fs.Output().
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRunHelpFlag(t *testing.T) {
	var err error
	out := captureStderr(t, func() {
		err = run([]string{"-h"})
	})
	if err != nil {
		t.Fatalf("run([-h]) returned error: %v, want nil", err)
	}
	if !strings.Contains(out, "SYNOPSIS") {
		t.Errorf("usage output missing SYNOPSIS section, got: %q", out)
	}
	if !strings.Contains(out, "-diff") {
		t.Errorf("usage output missing -diff, got: %q", out)
	}
}

func TestRunUnknownFlag(t *testing.T) {
	var err error
	captureStderr(t, func() {
		err = run([]string{"-bogus"})
	})
	if err == nil {
		t.Fatal("run([-bogus]) returned nil error, want an error")
	}
}

func TestRunDiffConflictingFlags(t *testing.T) {
	err := run([]string{"-diff", "-json", "a", "b"})
	if err == nil || !strings.Contains(err.Error(), "-diff cannot be combined") {
		t.Fatalf("run(-diff -json a b) error = %v, want mention of -diff cannot be combined", err)
	}
}

func TestRunTooManyArgs(t *testing.T) {
	var err error
	captureStderr(t, func() {
		err = run([]string{"a", "b", "c"})
	})
	if err == nil || !strings.Contains(err.Error(), "too many arguments") {
		t.Fatalf("run(a b c) error = %v, want too many arguments", err)
	}
}
