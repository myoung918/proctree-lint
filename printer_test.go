package main

import (
	"bytes"
	"testing"
)

func node(pid int, command string, children ...*Node) *Node {
	return &Node{Process: Process{PID: pid, Command: command}, Children: children}
}

func TestPrintSingleNode(t *testing.T) {
	var buf bytes.Buffer
	Print(&buf, []*Node{node(1, "init")})

	want := "1 init\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

func TestPrintExample(t *testing.T) {
	// Mirrors the tree in the README's usage example.
	tree := node(1, "init",
		node(2, "bash",
			node(3, "make -j4",
				node(4, "cc -c main.c"),
				node(5, "cc -c util.c"),
			),
		),
		node(6, "sshd"),
	)

	var buf bytes.Buffer
	Print(&buf, []*Node{tree})

	want := "" +
		"1 init\n" +
		"├─ 2 bash\n" +
		"│  └─ 3 make -j4\n" +
		"│     ├─ 4 cc -c main.c\n" +
		"│     └─ 5 cc -c util.c\n" +
		"└─ 6 sshd\n"
	if buf.String() != want {
		t.Errorf("output =\n%s\nwant\n%s", buf.String(), want)
	}
}

func TestPrintSortsChildrenAndRoots(t *testing.T) {
	// Roots and children are passed out of pid order; Print should sort both.
	roots := []*Node{
		node(6, "sshd"),
		node(1, "init",
			node(3, "make -j4"),
			node(2, "bash"),
		),
	}

	var buf bytes.Buffer
	Print(&buf, roots)

	want := "" +
		"1 init\n" +
		"├─ 2 bash\n" +
		"└─ 3 make -j4\n" +
		"6 sshd\n"
	if buf.String() != want {
		t.Errorf("output =\n%s\nwant\n%s", buf.String(), want)
	}
}
