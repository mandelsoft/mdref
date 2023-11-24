package main

import (
	"fmt"
	"strings"
)

func Print(files []*File, resolution Resolution) {

	for _, f := range files {
		if strings.HasSuffix(f.relpath, ".md") {
			fmt.Printf("*** %s: markdown\n", f.relpath)
			if len(f.targets) > 0 {
				fmt.Printf("  targets:\n")
				for k, ref := range f.targets {
					if ref.generate {
						fmt.Printf("   - %s: %s\n", k, ref.text)
					} else {
						fmt.Printf("   - %s->%s: %s\n", k, ref.anchor, ref.text)
					}
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
			if len(f.commands) > 0 {
				fmt.Printf("  commands:\n")
				for k, _ := range f.commands {
					fmt.Printf("   - %s\n", k)
				}
			}
		} else {
			fmt.Printf("*** %s: additional file\n", f.relpath)
		}
	}
}
