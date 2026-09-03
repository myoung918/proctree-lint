package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "proctree: "+err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("proctree", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "emit the tree as JSON instead of pretty-printing it")
	find := fs.Int("find", 0, "print only the subtree rooted at this pid")
	diff := fs.Bool("diff", false, "compare two snapshots and print what was added, removed, or changed")
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), `NAME
       proctree - validate and print pid/ppid/command process tree snapshots

SYNOPSIS
       proctree [-json] [-find pid] [file]
       proctree -diff old-file new-file

DESCRIPTION
       proctree reads a process tree snapshot from file, or from stdin if
       file is omitted or "-". Each line of the snapshot has the form

              pid ppid command

       proctree validates that the pids are unique, that every non-zero
       ppid resolves to another pid in the snapshot, and that there are
       no cycles, then prints the result indented like pstree(1).

       With -diff, proctree instead compares two snapshots by pid and
       prints what was added, removed, or changed between them. At most
       one of the two files may be "-" for stdin.

OPTIONS
       -json
              Emit the tree as JSON instead of pretty-printing it.

       -find pid
              Print only the subtree rooted at pid instead of the whole
              forest. Combines with -json.

       -diff
              Compare two snapshots given as old-file and new-file and
              print the differences by pid. Cannot be combined with
              -json or -find.

       -h, -help
              Print this help and exit.

EXAMPLES
       proctree snapshot.proctree
       awk '{print $1, $2, $3}' /proc/*/stat | proctree
       proctree -json -find 3 snapshot.proctree
       proctree -diff before.proctree after.proctree
`)
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *diff {
		if *jsonOutput || *find != 0 {
			return fmt.Errorf("-diff cannot be combined with -json or -find")
		}
		if fs.NArg() != 2 {
			fs.Usage()
			return fmt.Errorf("-diff requires exactly two files")
		}
		return runDiff(fs.Arg(0), fs.Arg(1))
	}

	var r io.Reader
	switch fs.NArg() {
	case 0:
		r = os.Stdin
	case 1:
		rd, closeFn, err := openInput(fs.Arg(0))
		if err != nil {
			return err
		}
		defer closeFn()
		r = rd
	default:
		fs.Usage()
		return fmt.Errorf("too many arguments")
	}

	procs, err := Parse(r)
	if err != nil {
		return err
	}
	roots, err := Validate(procs)
	if err != nil {
		return err
	}

	if *find != 0 {
		n, ok := Find(roots, *find)
		if !ok {
			return fmt.Errorf("no such pid: %d", *find)
		}
		roots = []*Node{n}
	}

	if *jsonOutput {
		return PrintJSON(os.Stdout, roots)
	}
	Print(os.Stdout, roots)
	return nil
}

// openInput opens name for reading, treating "-" as stdin. The returned
// close function is always safe to call, even for stdin.
func openInput(name string) (io.Reader, func() error, error) {
	if name == "-" {
		return os.Stdin, func() error { return nil }, nil
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}
	return f, f.Close, nil
}

// runDiff parses and validates the snapshots at oldName and newName, then
// prints what changed between them.
func runDiff(oldName, newName string) error {
	oldProcs, err := parseFile(oldName)
	if err != nil {
		return err
	}
	newProcs, err := parseFile(newName)
	if err != nil {
		return err
	}
	if _, err := Validate(oldProcs); err != nil {
		return fmt.Errorf("%s: %w", oldName, err)
	}
	if _, err := Validate(newProcs); err != nil {
		return fmt.Errorf("%s: %w", newName, err)
	}

	PrintDiff(os.Stdout, Diff(oldProcs, newProcs))
	return nil
}

func parseFile(name string) ([]Process, error) {
	r, closeFn, err := openInput(name)
	if err != nil {
		return nil, err
	}
	defer closeFn()
	return Parse(r)
}
