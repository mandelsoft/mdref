package main

import (
	"regexp"
)

var subExp = regexp.MustCompile(`{{{([*]?[A-Za-z][a-z0-9.-]+)}}}`)
var refExp = regexp.MustCompile(`\({{([a-z0-9.-]+)}}\)`)
var lnkExp = regexp.MustCompile(`\[{{([*]?[A-Za-z][a-z0-9.-]*)}}\]`)
var tgtExp = regexp.MustCompile(`{{([a-z][a-z0-9.-]*)(:([a-zA-Z][\p{L}\p{N}- ]+))?}}`)
var cmdExp = regexp.MustCompile(`{{([a-z]+)}((?:{[^}]*})+)}`)

var FilterPattern = map[string]*regexp.Regexp{}

// --- begin go-func ---
const go_func = "(?m)^\\s*func\\s+(?:\\(\\s*\\w+\\s+[\\w*]+\\s*\\)\\s+)?(\\w+)"

// --- end go-func ---

// --- begin go-type ---
const go_type = "(?m)^\\s*([_a-zA-Z]+) *(?:struct|func|interface|\\[|=)"

// --- end go-type ---

// --- begin go-var ---
const go_var = "(?m)^\\s*var +([_a-zA-Z]+) *= *"

// --- end go-var ---

// --- begin go-const ---
const go_const = "(?m)^\\s*const +([_a-zA-Z]+) *= *"

// --- end go-const ---

// --- begin go-const-value ---
const go_const_value = "(?m)^\\s*const +[_a-zA-Z]+ *= *(.*)\n"

// --- end go-const-value ---

// --- begin go-line-comment ---
// extract comment text from a Go source file using line comment syntax.
const go_line_comment = "(?m)^\\s*//\\s*(.*)\n"

// --- end go-line-comment ---

// --- begin go-comment ---
/*
   extract comment text from a Go source file using standard comment syntax.
   It does not support the exclusion of comment lines starting with ` *`.
*/
const go_comment = "/\\*\\s*([^*]*(?:\\*+[^/][^*]*)*)\\s*\\*/"

// --- end go-comment ---

// --- begin html-comment ---
// <!----- extract HTML comment. ---->
const html_comment = "<!--+\\s*([\\s\\S]*?)\\s*-+->"

// --- end html-comment ---

func init() {
	MustAddFilter("go-func", go_func)
	MustAddFilter("go-type", go_type)
	MustAddFilter("go-var", go_var)
	MustAddFilter("go-const", go_const)
	MustAddFilter("go-const-value", go_const_value)

	MustAddFilter("go-line-comment", go_line_comment)
	MustAddFilter("go-comment", go_comment)
	MustAddFilter("html-comment", html_comment)
}

func MustAddFilter(name, expr string) {
	if err := AddFilter(name, expr); err != nil {
		panic(err)
	}
}

func AddFilter(name, expr string) error {
	exp, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	FilterPattern[name] = exp
	return nil
}
