package main

import (
	"strings"
	"testing"
)

func TestParseBasic(t *testing.T) {
	input := "# a comment\n\n1 0 init\n2 1 make -j4\n\n3 1 sshd\n"
	procs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(procs) != 3 {
		t.Fatalf("got %d processes, want 3", len(procs))
	}

	want := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 3},
		{PID: 2, PPID: 1, Command: "make -j4", Line: 4},
		{PID: 3, PPID: 1, Command: "sshd", Line: 6},
	}
	for i, w := range want {
		if procs[i] != w {
			t.Errorf("procs[%d] = %+v, want %+v", i, procs[i], w)
		}
	}
}

func TestParseCommandWithSpaces(t *testing.T) {
	procs, err := Parse(strings.NewReader("1 0 cc -c main.c -o main.o\n"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if got := procs[0].Command; got != "cc -c main.c -o main.o" {
		t.Errorf("Command = %q, want %q", got, "cc -c main.c -o main.o")
	}
}

func TestParseNoProcesses(t *testing.T) {
	_, err := Parse(strings.NewReader("# nothing here\n\n"))
	if err == nil {
		t.Fatal("expected an error for input with no processes, got nil")
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
		line  int
	}{
		{"too few fields", "1 0\n", 1},
		{"non-numeric pid", "x 0 init\n", 1},
		{"zero pid", "0 0 init\n", 1},
		{"negative pid", "-1 0 init\n", 1},
		{"non-numeric ppid", "1 x init\n", 1},
		{"negative ppid", "1 -1 init\n", 1},
		{"error on later line", "1 0 init\n2 bad init\n", 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(c.input))
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			perr, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("error is %T, want *ParseError", err)
			}
			if perr.Line != c.line {
				t.Errorf("error line = %d, want %d", perr.Line, c.line)
			}
		})
	}
}
