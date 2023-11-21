package main

import (
	"fmt"
	"os"
)

func resolve(files []*File) (Resolution, error) {
	resolution := Resolution{}

	failed := 0

	for _, f := range files {
		for k := range f.targets {
			if r, ok := resolution[k]; ok {
				fmt.Fprintf(os.Stderr, "%s: duplicate reference target %s in %s\n", f.relpath, k, r.relpath)
				failed++
			} else {
				resolution[k] = f
			}
		}
	}

	for _, f := range files {
		for k := range f.refs {
			if resolution[k] == nil {
				fmt.Fprintf(os.Stderr, "%s: reference %s not found\n", f.relpath, k)
				failed++
			}
		}
		for k := range f.terms {
			if r := resolution[k]; r == nil {
				fmt.Fprintf(os.Stderr, "%s: term reference %s not found\n", f.relpath, k)
				failed++
			} else {
				if r.targets[k] == "" {
					fmt.Fprintf(os.Stderr, "%s: target %s in %s without term\n", f.relpath, k, r.relpath)
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
