package main

import "fmt"

// Node is a Process with its children attached, built by Validate.
type Node struct {
	Process
	Children []*Node
}

// Validate checks that procs forms a well-formed process forest:
//   - every pid is unique
//   - every nonzero ppid refers to a pid present in procs
//   - there are no cycles
//
// On success it returns the roots (pid entries with ppid 0), each with
// its descendants attached.
func Validate(procs []Process) ([]*Node, error) {
	nodes := make(map[int]*Node, len(procs))
	for _, p := range procs {
		if existing, ok := nodes[p.PID]; ok {
			return nil, fmt.Errorf("line %d: duplicate pid %d (first seen at line %d)", p.Line, p.PID, existing.Line)
		}
		nodes[p.PID] = &Node{Process: p}
	}

	var roots []*Node
	for _, p := range procs {
		n := nodes[p.PID]
		if p.PPID == 0 {
			roots = append(roots, n)
			continue
		}
		parent, ok := nodes[p.PPID]
		if !ok {
			return nil, fmt.Errorf("line %d: pid %d has ppid %d, which does not appear in the input", p.Line, p.PID, p.PPID)
		}
		parent.Children = append(parent.Children, n)
	}

	// A pid whose ppid chain never reaches a root (ppid 0) is part of a
	// cycle - every acyclic chain terminates there since pids are finite.
	// Walk every node's ancestry rather than just descending from roots,
	// so a cycle disconnected from any root is still caught.
	for _, p := range procs {
		seen := make(map[int]bool)
		pid := p.PID
		for pid != 0 {
			if seen[pid] {
				return nil, fmt.Errorf("pid %d is part of a cycle", pid)
			}
			seen[pid] = true
			pid = nodes[pid].PPID
		}
	}

	return roots, nil
}

// Find returns the node for pid, searching the whole forest rooted at
// roots, not just the roots themselves.
func Find(roots []*Node, pid int) (*Node, bool) {
	for _, r := range roots {
		if n, ok := find(r, pid); ok {
			return n, true
		}
	}
	return nil, false
}

func find(n *Node, pid int) (*Node, bool) {
	if n.PID == pid {
		return n, true
	}
	for _, c := range n.Children {
		if found, ok := find(c, pid); ok {
			return found, true
		}
	}
	return nil, false
}
