package main

import (
	"encoding/json"
	"io"
	"sort"
)

// jsonNode is the JSON representation of a Node. It drops the Line field,
// which is only meaningful for error messages during parsing/validation.
type jsonNode struct {
	PID      int         `json:"pid"`
	PPID     int         `json:"ppid"`
	Command  string      `json:"command"`
	Children []*jsonNode `json:"children,omitempty"`
}

func toJSONNode(n *Node) *jsonNode {
	sort.Slice(n.Children, func(i, j int) bool { return n.Children[i].PID < n.Children[j].PID })

	children := make([]*jsonNode, len(n.Children))
	for i, c := range n.Children {
		children[i] = toJSONNode(c)
	}
	return &jsonNode{PID: n.PID, PPID: n.PPID, Command: n.Command, Children: children}
}

// PrintJSON writes roots as an indented JSON array, sorted by pid at every
// level to match Print's ordering.
func PrintJSON(w io.Writer, roots []*Node) error {
	sort.Slice(roots, func(i, j int) bool { return roots[i].PID < roots[j].PID })

	out := make([]*jsonNode, len(roots))
	for i, r := range roots {
		out[i] = toJSONNode(r)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
