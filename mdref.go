package main

import (
	"fmt"
	"os"

	"github.com/mandelsoft/mdref/version"
)

type Resolution map[string]*File

func (r Resolution) Resolve(ref string, src string) (string, string) {
	f := r[ref]
	return f.Resolve(ref, src)
}

func Error(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	print := false

	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "--version" {
			info := version.Get()

			fmt.Printf("mdgen version %s.%s.%s (%s) [%s %s]\n", info.Major, info.Minor, info.Patch, info.PreRelease, info.GitTreeState, info.GitCommit)
			os.Exit(0)
		}

		if args[0] == "--help" {
			fmt.Printf("mdgen [--doc] [--copy] [<source dir> [<target dir>]]\n")
			fmt.Printf(`
Flags:
  --list   print ref list

mdref evalates a document tree with markdown files containing logical references
and resolves thoses refs to markdown links. The generated tree is written
to a target folder.

If no target directory is given, only a consistency check is done.
`)
			os.Exit(0)
		}
		if args[0] == "--list" {
			print = true
			args = args[1:]
		}
	}
	if len(args) > 2 {
		fmt.Printf("use mdref [--liat] [<source> [<target>]]")
		os.Exit(1)
	}
	src := "."
	if len(args) > 0 {
		src = args[0]
	}
	dst := ""
	if len(args) > 1 {
		dst = args[1]
	}

	files, err := scan(src, "")
	Error(err)

	resolution, err := resolve(files)
	Error(err)

	if print {
		Print(files, resolution)
	} else {
		if dst != "" {
			err := generate(files, resolution, src, dst)
			Error(err)
		}
	}
}
