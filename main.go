package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// version is bumped by hand for each release; there's no build tooling
// here to stamp it from a tag.
const version = "0.5.0"

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
	showVersion := fs.Bool("version", false, "print the version number and exit")
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), `NAME
       proctree - validate and print pid/ppid/command process tree snapshots

SYNOPSIS
       proctree [-json] [-find pid] [file...]
       proctree -diff old-file new-file
       proctree -version

DESCRIPTION
       proctree reads a process tree snapshot from file, or from stdin if
       file is omitted or "-". Each line of the snapshot has the form

              pid ppid command

       If more than one file is given, their processes are merged into a
       single forest before validation, as if the lines had all come
       from one file. At most one file may be "-", since stdin can only
       be read once.

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

       -version
              Print the version number and exit.

       -h, -help
              Print this help and exit.

EXAMPLES
       proctree snapshot.proctree
       proctree host1.proctree host2.proctree
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

	if *showVersion {
		fmt.Fprintln(os.Stdout, "proctree "+version)
		return nil
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

	procs, err := readSnapshots(fs.Args())
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

// readSnapshots parses and concatenates the snapshots at names into one
// list of processes, tagging each with the file it came from so that
// Validate's error messages can tell overlapping pids apart. No names
// means read a single snapshot from stdin.
func readSnapshots(names []string) ([]Process, error) {
	if len(names) == 0 {
		names = []string{"-"}
	}

	stdinCount := 0
	for _, name := range names {
		if name == "-" {
			stdinCount++
		}
	}
	if stdinCount > 1 {
		return nil, fmt.Errorf(`stdin ("-") can only be given once`)
	}

	var all []Process
	for _, name := range names {
		procs, err := parseFile(name)
		if err != nil {
			if name == "-" {
				return nil, err
			}
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if name != "-" {
			for i := range procs {
				procs[i].File = name
			}
		}
		all = append(all, procs...)
	}
	return all, nil
}
