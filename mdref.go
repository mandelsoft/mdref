package main

import (
	"fmt"
	"os"
	"strings"
)

type Resolution map[string]*File

func (r Resolution) Resolve(ref string, src string) (string, string) {
	f := r[ref]
	return f.Resolve(ref, src)
}

type Options struct {
	SkipExecute bool
	SkipSource  bool
	Headings    bool
	Print       bool
}

func Error(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func Syntax() {
	fmt.Printf("mdref {<options>} [<source dir> [<target dir>]]\n")
	fmt.Printf(`
Flags:
  --version      just print the program version
  --help         this help text
  --headings     prefer using standard heading anchors
  --skip-execute omit the evaluation of the execute statement (for test purposes, only)
  --skip-source  omit source reference in generation comment
  --list         print reference index and usage list
`)
}

func main() {
	var opts Options

	args := os.Args[1:]
	for len(args) > 0 && strings.HasPrefix(args[0], "--") {
		if args[0] == "--version" {
			info := Get()

			fmt.Printf("mdgen version %s.%s.%s (%s) [%s %s]\n", info.Major, info.Minor, info.Patch, info.PreRelease, info.GitTreeState, info.GitCommit)
			os.Exit(0)
		}

		if args[0] == "--syntax" {
			Syntax()
			os.Exit(0)
		}
		if args[0] == "--help" {
			Syntax()
			fmt.Printf(`
mdref evalates a document tree with markdown files containing logical references
and resolves thoses refs to markdown links. The generated tree is written
to a target folder. Directories with the name 'local' will be ignored.
They may be used to provide included content not copied to the output folder.

If no target directory is given, only a consistency check is done.
If the option --headings is given, reference targets before or after
a standard Markdown heading will use the Markdown heading anchor.

If the option --list is given a reference index and usage list is
printed, additionally.
`)
			os.Exit(0)
		}

		switch args[0] {
		case "--list":
			opts.Print = true
			args = args[1:]
		case "--skip-execute":
			opts.SkipExecute = true
			args = args[1:]
		case "--skip-source":
			opts.SkipSource = true
			args = args[1:]
		case "--headings":
			opts.Headings = true
			args = args[1:]
		default:
			Error(fmt.Errorf("invalid option %q", args[0]))
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

	files, err := scan(src, "", opts)
	Error(err)

	resolution, err := resolve(files)
	Error(err)
	Error(checkCommands(src, files, opts))

	if opts.Print {
		Print(files, resolution)
	}

	if dst != "" {
		err := generate(files, resolution, src, dst, opts)
		Error(err)
	}
}
