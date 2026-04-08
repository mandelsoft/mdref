package main

import (
	"fmt"
	"iter"
	"regexp"
	"strings"

	"github.com/mandelsoft/goutils/topo"
)

var Variables = map[string]string{}

var substPattern = regexp.MustCompile(`\$\(([a-zA-Z]+)\)`)

func Subst(s string) (string, error) {
	s, _, err := subst(s)
	return s, err
}

func subst(s string) (string, bool, error) {
	indices := substPattern.FindAllStringSubmatchIndex(s, -1)
	result := []byte(s)
	for i := len(indices) - 1; i >= 0; i-- {
		idx := indices[i]
		key := string(result[idx[2]:idx[3]])
		if val, ok := Variables[key]; ok {
			result = append(result[:idx[0]], append([]byte(val), result[idx[1]:]...)...)
		} else {
			return "", false, fmt.Errorf("variable %q not found", key)
		}
	}
	return string(result), len(indices) > 0, nil
}

func dependencies(s string) iter.Seq2[string, string] {
	v := Variables[s]
	indices := substPattern.FindAllStringSubmatchIndex(v, -1)
	return func(yield func(string, string) bool) {
		for _, idx := range indices {
			key := string(v[idx[2]:idx[3]])
			if !yield(key, key) {
				return
			}
		}
	}
}

func elements(yield func(string, string) bool) {
	for k := range Variables {
		if !yield(k, k) {
			return
		}
	}
}

func NormalizeVariables() error {
	order, cycle := topo.Sort[string, string](elements, dependencies)
	if len(cycle) > 0 {
		return fmt.Errorf("cycle detected in variable definitions: %s", strings.Join(cycle, "->"))
	}
	for _, k := range order {
		v, err := Subst(Variables[k])
		if err != nil {
			return fmt.Errorf("variable %q:", k, err)
		}
		Variables[k] = v
	}
	return nil
}
