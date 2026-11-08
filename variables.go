package main

import (
	"fmt"
	"iter"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/topo"
)

// Variables are the variables defined by variable commands.
// After NormalizeVariables, their values only contain references
// to predefined variables.
var Variables = map[string]string{}

// varNameExp is the syntax of variable names, used for
// definitions and substitutions.
const varNameExp = "[a-zA-Z][a-zA-Z0-9_]*"

var substPattern = regexp.MustCompile(`\$\((` + varNameExp + `)\)`)

// PredefinedVariables provide the values of the predefined variables
// for the document context of a substitution. Their names are reserved.
var PredefinedVariables = map[string]func(ctx *DocContext) (string, error){
	// DOC_ROOT is the absolute path of the source folder.
	"DOC_ROOT": func(ctx *DocContext) (string, error) {
		return ctx.DocRoot(), nil
	},
	// REL_DIR_TO_DOC_ROOT is the relative path from the folder of the
	// document to the source folder.
	"REL_DIR_TO_DOC_ROOT": func(ctx *DocContext) (string, error) {
		dir, err := filepath.Abs(filepath.Dir(ctx.Document()))
		if err != nil {
			return "", err
		}
		return filepath.Rel(dir, ctx.DocRoot())
	},
}

// Subst substitutes the variables in a command argument
// for the given document context.
func Subst(s string, ctx *DocContext) (string, error) {
	return substWith(s, func(key string) (string, bool, error) {
		if f, ok := PredefinedVariables[key]; ok {
			v, err := f(ctx)
			return v, true, err
		}
		v, ok := Variables[key]
		if !ok {
			return "", false, nil
		}
		// the normalized value may refer to predefined variables
		v, err := Subst(v, ctx)
		return v, true, err
	})
}

// substWith substitutes the variables found by the lookup function.
func substWith(s string, lookup func(key string) (string, bool, error)) (string, error) {
	indices := substPattern.FindAllStringSubmatchIndex(s, -1)
	result := []byte(s)
	for i := len(indices) - 1; i >= 0; i-- {
		idx := indices[i]
		key := string(result[idx[2]:idx[3]])
		val, ok, err := lookup(key)
		if err != nil {
			return "", fmt.Errorf("variable %q: %w", key, err)
		}
		if !ok {
			return "", fmt.Errorf("variable %q not found", key)
		}
		result = append(result[:idx[0]], append([]byte(val), result[idx[1]:]...)...)
	}
	return string(result), nil
}

// dependencies provides the variables used by a defined variable,
// predefined variables are not considered.
func dependencies(s string) iter.Seq2[string, string] {
	v := Variables[s]
	indices := substPattern.FindAllStringSubmatchIndex(v, -1)
	return func(yield func(string, string) bool) {
		for _, idx := range indices {
			key := string(v[idx[2]:idx[3]])
			if _, ok := PredefinedVariables[key]; ok {
				continue
			}
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

// checkReferences reports the first (in the order of the variable names)
// use of an undefined variable by a variable definition.
func checkReferences() error {
	for _, k := range slices.Sorted(maps.Keys(Variables)) {
		for d := range dependencies(k) {
			if _, ok := Variables[d]; !ok {
				return fmt.Errorf("variable %q: variable %q not found", k, d)
			}
		}
	}
	return nil
}

func NormalizeVariables() error {
	if err := checkReferences(); err != nil {
		return err
	}
	order, cycle := topo.Sort[string, string](elements, dependencies)
	if len(cycle) > 0 {
		return fmt.Errorf("cycle detected in variable definitions: %s", strings.Join(cycle, "->"))
	}
	for _, k := range order {
		// references to predefined variables are kept, they
		// are substituted for the document using the variable.
		v, err := substWith(Variables[k], func(key string) (string, bool, error) {
			if _, ok := PredefinedVariables[key]; ok {
				return "$(" + key + ")", true, nil
			}
			v, ok := Variables[key]
			return v, ok, nil
		})
		if err != nil {
			return fmt.Errorf("variable %q: %w", k, err)
		}
		Variables[k] = v
	}
	return nil
}
