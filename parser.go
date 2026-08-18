package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Process is one entry from a process tree snapshot.
type Process struct {
	PID     int
	PPID    int
	Command string
	Line    int // source line, for error messages
}

// ParseError reports a problem with a specific line of input.
type ParseError struct {
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

// Parse reads a process tree snapshot and returns the entries it found.
//
// The format is one process per line: "pid ppid command", fields
// separated by a single space, command running to the end of the line.
// A ppid of 0 marks a root. Blank lines and lines starting with # are
// ignored.
//
// Parse only checks that individual lines are well formed. It does not
// check that the file as a whole describes a valid tree (unique pids,
// resolvable parents, no cycles) - that is Validate's job.
func Parse(r io.Reader) ([]Process, error) {
	var procs []Process
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1<<20)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.SplitN(line, " ", 3)
		if len(fields) < 3 {
			return nil, &ParseError{lineNo, `expected "pid ppid command"`}
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil || pid <= 0 {
			return nil, &ParseError{lineNo, fmt.Sprintf("invalid pid %q", fields[0])}
		}

		ppid, err := strconv.Atoi(fields[1])
		if err != nil || ppid < 0 {
			return nil, &ParseError{lineNo, fmt.Sprintf("invalid ppid %q", fields[1])}
		}

		command := strings.TrimSpace(fields[2])
		if command == "" {
			return nil, &ParseError{lineNo, "empty command"}
		}

		procs = append(procs, Process{PID: pid, PPID: ppid, Command: command, Line: lineNo})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(procs) == 0 {
		return nil, fmt.Errorf("no processes found in input")
	}
	return procs, nil
}
