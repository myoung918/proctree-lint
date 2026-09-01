package main

import (
	"fmt"
	"io"
	"sort"
)

// Change describes how a single pid differs between two snapshots.
type Change struct {
	PID    int
	Status string // "added", "removed", or "changed"
	Old    Process
	New    Process
}

// Diff compares two snapshots by pid and reports what changed: pids
// present in only one of them, and pids present in both whose ppid or
// command differs. It works on the flat Process lists rather than
// validated trees, since a diff between two well-formed snapshots is
// still meaningful pid by pid.
func Diff(oldProcs, newProcs []Process) []Change {
	oldByPID := make(map[int]Process, len(oldProcs))
	for _, p := range oldProcs {
		oldByPID[p.PID] = p
	}
	newByPID := make(map[int]Process, len(newProcs))
	for _, p := range newProcs {
		newByPID[p.PID] = p
	}

	pids := make(map[int]bool, len(oldByPID)+len(newByPID))
	for pid := range oldByPID {
		pids[pid] = true
	}
	for pid := range newByPID {
		pids[pid] = true
	}

	var changes []Change
	for pid := range pids {
		o, inOld := oldByPID[pid]
		n, inNew := newByPID[pid]
		switch {
		case inOld && !inNew:
			changes = append(changes, Change{PID: pid, Status: "removed", Old: o})
		case !inOld && inNew:
			changes = append(changes, Change{PID: pid, Status: "added", New: n})
		case o.PPID != n.PPID || o.Command != n.Command:
			changes = append(changes, Change{PID: pid, Status: "changed", Old: o, New: n})
		}
	}

	sort.Slice(changes, func(i, j int) bool { return changes[i].PID < changes[j].PID })
	return changes
}

// PrintDiff writes changes one pid at a time: a "-" line for what left,
// a "+" line for what arrived, both for a pid whose ppid or command
// changed.
func PrintDiff(w io.Writer, changes []Change) {
	for _, c := range changes {
		switch c.Status {
		case "added":
			fmt.Fprintf(w, "+ %d %d %s\n", c.New.PID, c.New.PPID, c.New.Command)
		case "removed":
			fmt.Fprintf(w, "- %d %d %s\n", c.Old.PID, c.Old.PPID, c.Old.Command)
		case "changed":
			fmt.Fprintf(w, "- %d %d %s\n", c.Old.PID, c.Old.PPID, c.Old.Command)
			fmt.Fprintf(w, "+ %d %d %s\n", c.New.PID, c.New.PPID, c.New.Command)
		}
	}
}
