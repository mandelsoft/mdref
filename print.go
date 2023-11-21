package main

import (
	"fmt"
)

func Print(files []*File, resolution Resolution) {

	for _, f := range files {
		fmt.Printf("*** %s:\n", f.relpath)
		if len(f.targets) > 0 {
			fmt.Printf("  targets:\n")
			for k, str := range f.targets {
				fmt.Printf("   - %s: %s\n", k, str)
			}
		}
		if len(f.refs) > 0 {
			fmt.Printf("  references:\n")
			for k, _ := range f.refs {
				ref, _ := resolution.Resolve(k, f.relpath)
				fmt.Printf("   - %s: %s\n", k, ref)
			}
		}
		if len(f.terms) > 0 {
			fmt.Printf("  terms refs:\n")
			for k, _ := range f.terms {
				ref, str := resolution.Resolve(k, f.relpath)
				fmt.Printf("   - %s: %s[%s]\n", k, ref, str)
			}
		}
	}
}
