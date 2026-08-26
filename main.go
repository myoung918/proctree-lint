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
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), `usage: proctree [-json] [-find pid] [file]

reads a process tree snapshot from file, or from stdin if file is
omitted or "-"

flags:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	var r io.Reader
	switch fs.NArg() {
	case 0:
		r = os.Stdin
	case 1:
		if fs.Arg(0) == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(fs.Arg(0))
			if err != nil {
				return err
			}
			defer f.Close()
			r = f
		}
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
