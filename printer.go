package main

import (
	"fmt"
	"io"
	"sort"
)

// Print writes roots as an indented tree, in the style of pstree(1).
func Print(w io.Writer, roots []*Node) {
	sort.Slice(roots, func(i, j int) bool { return roots[i].PID < roots[j].PID })
	for _, r := range roots {
		printNode(w, r, "", "", true)
	}
}

func printNode(w io.Writer, n *Node, prefix, connector string, last bool) {
	fmt.Fprintf(w, "%s%s%d %s\n", prefix, connector, n.PID, n.Command)

	childPrefix := prefix
	if connector != "" {
		if last {
			childPrefix += "   "
		} else {
			childPrefix += "│  "
		}
	}

	sort.Slice(n.Children, func(i, j int) bool { return n.Children[i].PID < n.Children[j].PID })
	for i, c := range n.Children {
		childLast := i == len(n.Children)-1
		conn := "├─ "
		if childLast {
			conn = "└─ "
		}
		printNode(w, c, childPrefix, conn, childLast)
	}
}
