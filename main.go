package main

import (
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
	var r io.Reader
	switch len(args) {
	case 0:
		r = os.Stdin
	case 1:
		if args[0] == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()
			r = f
		}
	default:
		return fmt.Errorf(`usage: proctree [file]

reads a process tree snapshot from file, or from stdin if file is
omitted or "-"`)
	}

	procs, err := Parse(r)
	if err != nil {
		return err
	}
	roots, err := Validate(procs)
	if err != nil {
		return err
	}
	Print(os.Stdout, roots)
	return nil
}
