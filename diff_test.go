package main

import (
	"bytes"
	"testing"
)

func TestDiffAddedRemovedChanged(t *testing.T) {
	oldProcs := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 1},
		{PID: 2, PPID: 1, Command: "bash", Line: 2},
		{PID: 3, PPID: 1, Command: "sshd", Line: 3},
	}
	newProcs := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 1},
		{PID: 2, PPID: 1, Command: "zsh", Line: 2},
		{PID: 4, PPID: 1, Command: "cron", Line: 3},
	}

	changes := Diff(oldProcs, newProcs)
	if len(changes) != 3 {
		t.Fatalf("got %d changes, want 3: %+v", len(changes), changes)
	}

	// Sorted by pid: 2 (changed), 3 (removed), 4 (added).
	if changes[0].PID != 2 || changes[0].Status != "changed" {
		t.Errorf("changes[0] = %+v, want pid 2 changed", changes[0])
	}
	if changes[0].Old.Command != "bash" || changes[0].New.Command != "zsh" {
		t.Errorf("changes[0] old/new commands = %q/%q, want bash/zsh", changes[0].Old.Command, changes[0].New.Command)
	}
	if changes[1].PID != 3 || changes[1].Status != "removed" {
		t.Errorf("changes[1] = %+v, want pid 3 removed", changes[1])
	}
	if changes[2].PID != 4 || changes[2].Status != "added" {
		t.Errorf("changes[2] = %+v, want pid 4 added", changes[2])
	}
}

func TestDiffReparented(t *testing.T) {
	oldProcs := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 1},
		{PID: 2, PPID: 1, Command: "make -j4", Line: 2},
		{PID: 3, PPID: 2, Command: "cc -c main.c", Line: 3},
	}
	newProcs := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 1},
		{PID: 3, PPID: 1, Command: "cc -c main.c", Line: 2},
	}

	changes := Diff(oldProcs, newProcs)
	if len(changes) != 2 {
		t.Fatalf("got %d changes, want 2: %+v", len(changes), changes)
	}
	if changes[0].PID != 2 || changes[0].Status != "removed" {
		t.Errorf("changes[0] = %+v, want pid 2 removed", changes[0])
	}
	if changes[1].PID != 3 || changes[1].Status != "changed" {
		t.Errorf("changes[1] = %+v, want pid 3 changed", changes[1])
	}
	if changes[1].Old.PPID != 2 || changes[1].New.PPID != 1 {
		t.Errorf("changes[1] old/new ppid = %d/%d, want 2/1", changes[1].Old.PPID, changes[1].New.PPID)
	}
}

func TestDiffNoChanges(t *testing.T) {
	procs := []Process{
		{PID: 1, PPID: 0, Command: "init", Line: 1},
	}
	if changes := Diff(procs, procs); len(changes) != 0 {
		t.Errorf("got %d changes for identical snapshots, want 0: %+v", len(changes), changes)
	}
}

func TestPrintDiff(t *testing.T) {
	changes := []Change{
		{PID: 2, Status: "changed", Old: Process{PID: 2, PPID: 1, Command: "bash"}, New: Process{PID: 2, PPID: 1, Command: "zsh"}},
		{PID: 3, Status: "removed", Old: Process{PID: 3, PPID: 1, Command: "sshd"}},
		{PID: 4, Status: "added", New: Process{PID: 4, PPID: 1, Command: "cron"}},
	}

	var buf bytes.Buffer
	PrintDiff(&buf, changes)

	want := "- 2 1 bash\n+ 2 1 zsh\n- 3 1 sshd\n+ 4 1 cron\n"
	if got := buf.String(); got != want {
		t.Errorf("PrintDiff output =\n%s\nwant\n%s", got, want)
	}
}
