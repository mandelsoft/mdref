package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func resolve(files []*File) (Resolution, error) {
	resolution := Resolution{}

	failed := 0

	for _, f := range files {
		for k, d := range f.targets {
			if r, ok := resolution[k]; ok {
				fmt.Fprintf(os.Stderr, "%s: %s: duplicate reference target %s in %s\n", f.relpath, d.Position(), k, r.relpath)
				failed++
			} else {
				resolution[k] = f
			}
		}
	}

	for _, f := range files {
		for k, d := range f.refs {
			if resolution[k] == nil {
				fmt.Fprintf(os.Stderr, "%s: %s reference %q not found\n", f.relpath, d.Position(), k)
				failed++
			}
		}
		for k, d := range f.terms {
			if r := resolution[k]; r == nil {
				fmt.Fprintf(os.Stderr, "%s: %s: term reference %q not found\n", f.relpath, d.Position(), k)
				failed++
			} else {
				if r.targets[k].text == "" {
					fmt.Fprintf(os.Stderr, "%s: %s: term anchor %q in %s without term\n", f.relpath, d.Position(), k, r.relpath)
					failed++
				}
			}
		}
	}

	if failed > 0 {
		return nil, fmt.Errorf("failed with %d resolution error(s)", failed)
	}
	return resolution, nil
}

func checkCommands(src string, files []*File) error {
	failed := 0

	for _, f := range files {
		p := filepath.Join(src, f.relpath)
		for k, d := range f.commands {
			_, err := d.GetSubstitution(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s: command %q: %s\n", f.relpath, d.Position(), k, err)
				failed++
			}
		}
	}
	if failed > 0 {
		return fmt.Errorf("failed with %d resolution error(s)", failed)
	}
	return nil
}
