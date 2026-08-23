package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPrintJSONSingleNode(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintJSON(&buf, []*Node{node(1, "init")}); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	want := "[\n  {\n    \"pid\": 1,\n    \"ppid\": 0,\n    \"command\": \"init\"\n  }\n]\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

func TestPrintJSONNestedAndSorted(t *testing.T) {
	roots := []*Node{
		node(6, "sshd"),
		node(1, "init",
			node(3, "make -j4"),
			node(2, "bash"),
		),
	}

	var buf bytes.Buffer
	if err := PrintJSON(&buf, roots); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}

	var got []struct {
		PID      int  `json:"pid"`
		Children []struct {
			PID int `json:"pid"`
		} `json:"children"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}

	if len(got) != 2 || got[0].PID != 1 || got[1].PID != 6 {
		t.Fatalf("roots = %+v, want pids [1, 6] in order", got)
	}
	if len(got[0].Children) != 2 || got[0].Children[0].PID != 2 || got[0].Children[1].PID != 3 {
		t.Fatalf("init's children = %+v, want pids [2, 3] in order", got[0].Children)
	}
}

func TestPrintJSONLeafHasNoChildrenField(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintJSON(&buf, []*Node{node(1, "init")}); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}
	if bytes.Contains(buf.Bytes(), []byte("children")) {
		t.Errorf("output contains a children field for a leaf node: %s", buf.String())
	}
}
