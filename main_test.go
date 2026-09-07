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

// captureStdout redirects os.Stdout for the duration of fn and returns
// what was written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRunVersionFlag(t *testing.T) {
	var err error
	out := captureStdout(t, func() {
		err = run([]string{"-version"})
	})
	if err != nil {
		t.Fatalf("run([-version]) returned error: %v, want nil", err)
	}
	if !strings.Contains(out, version) {
		t.Errorf("version output = %q, want it to contain %q", out, version)
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

func TestRunStdinGivenTwice(t *testing.T) {
	err := run([]string{"-", "-"})
	if err == nil || !strings.Contains(err.Error(), "stdin") {
		t.Fatalf(`run(-, -) error = %v, want mention of stdin`, err)
	}
}

// writeTempFile creates a file under t.TempDir() containing content and
// returns its path.
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.proctree")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	return f.Name()
}

func TestRunMergesMultipleFiles(t *testing.T) {
	a := writeTempFile(t, "1 0 init\n2 1 bash\n")
	b := writeTempFile(t, "3 1 sshd\n")

	var err error
	out := captureStdout(t, func() {
		err = run([]string{a, b})
	})
	if err != nil {
		t.Fatalf("run(a, b) returned error: %v", err)
	}
	if !strings.Contains(out, "2 bash") || !strings.Contains(out, "3 sshd") {
		t.Errorf("output = %q, want it to contain processes from both files", out)
	}
}

func TestRunDuplicatePidAcrossFiles(t *testing.T) {
	a := writeTempFile(t, "1 0 init\n")
	b := writeTempFile(t, "1 0 vim\n")

	err := run([]string{a, b})
	if err == nil || !strings.Contains(err.Error(), "duplicate pid 1") {
		t.Fatalf("run(a, b) error = %v, want mention of duplicate pid 1", err)
	}
}
