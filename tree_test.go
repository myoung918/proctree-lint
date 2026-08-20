package main

import (
	"strings"
	"testing"
)

func TestValidateBuildsTree(t *testing.T) {
	procs := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 1},
		{PID: 6, PPID: 1, Command: "sshd", Line: 2},
		{PID: 2, PPID: 1, Command: "bash", Line: 3},
		{PID: 3, PPID: 2, Command: "make -j4", Line: 4},
	}
	roots, err := Validate(procs)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(roots) != 1 || roots[0].PID != 1 {
		t.Fatalf("roots = %+v, want a single root with pid 1", roots)
	}
	if len(roots[0].Children) != 2 {
		t.Fatalf("root has %d children, want 2", len(roots[0].Children))
	}

	var bash *Node
	for _, c := range roots[0].Children {
		if c.PID == 2 {
			bash = c
		}
	}
	if bash == nil {
		t.Fatal("pid 2 (bash) not found among root's children")
	}
	if len(bash.Children) != 1 || bash.Children[0].PID != 3 {
		t.Fatalf("bash.Children = %+v, want a single child with pid 3", bash.Children)
	}
}

func TestValidateMultipleRoots(t *testing.T) {
	procs := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 1},
		{PID: 2, PPID: 0, Command: "kthreadd", Line: 2},
	}
	roots, err := Validate(procs)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("got %d roots, want 2", len(roots))
	}
}

func TestValidateDuplicatePid(t *testing.T) {
	procs := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 1},
		{PID: 1, PPID: 0, Command: "vim", Line: 2},
	}
	_, err := Validate(procs)
	if err == nil {
		t.Fatal("expected an error for duplicate pid, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate pid 1") {
		t.Errorf("error = %q, want it to mention duplicate pid 1", err.Error())
	}
}

func TestValidateDanglingParent(t *testing.T) {
	procs := []Process{
		{PID: 2, PPID: 1, Command: "bash", Line: 1},
	}
	_, err := Validate(procs)
	if err == nil {
		t.Fatal("expected an error for a dangling parent, got nil")
	}
	if !strings.Contains(err.Error(), "ppid 1") {
		t.Errorf("error = %q, want it to mention the missing ppid", err.Error())
	}
}

func TestValidateCycle(t *testing.T) {
	procs := []Process{
		{PID: 1, PPID: 2, Command: "a", Line: 1},
		{PID: 2, PPID: 1, Command: "b", Line: 2},
	}
	_, err := Validate(procs)
	if err == nil {
		t.Fatal("expected an error for a cycle, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error = %q, want it to mention a cycle", err.Error())
	}
}

func TestValidateSelfParent(t *testing.T) {
	procs := []Process{
		{PID: 1, PPID: 1, Command: "a", Line: 1},
	}
	_, err := Validate(procs)
	if err == nil {
		t.Fatal("expected an error for a process that is its own parent, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error = %q, want it to mention a cycle", err.Error())
	}
}
