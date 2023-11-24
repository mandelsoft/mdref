package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Command interface {
	GetSubstitution(path string) ([]byte, error)
}

type Commands map[string]Command

type Include struct {
	file  string
	start int
	end   int
}

var includeExp = regexp.MustCompile("^{([^}]+)}(?:{([0-9]+)?(?::([0-9]+)?)?})?$")

func NewInclude(args []byte) (Command, error) {
	var err error

	matches := includeExp.FindSubmatch(args)

	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid include arguments %q", string(args))
	}
	start := int64(0)
	end := int64(0)
	if matches[2] != nil {
		start, err = strconv.ParseInt(string(matches[2]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid start line: %w", err)
		}
	}
	if matches[3] != nil {
		end, err = strconv.ParseInt(string(matches[3]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid start line: %w", err)
		}
	}
	return &Include{string(matches[1]), int(start), int(end)}, nil
}

func (i *Include) GetSubstitution(p string) ([]byte, error) {
	if !filepath.IsAbs(i.file) {
		p = filepath.Join(filepath.Dir(p), i.file)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("cannot read include file %q", i.file)
	}
	lines := strings.Split(string(data), "\n")
	start := 0
	if i.start > 0 {
		start = i.start - 1
	}
	if start >= len(lines) {
		return nil, fmt.Errorf("start line %d after end of file (%q %d lines", start, i.file, len(lines))
	}
	end := len(lines)
	if i.end > 0 {
		end = i.end
	}
	if end > len(lines) {
		return nil, fmt.Errorf("end line %d after end of file (%q %d lines", end, i.file, len(lines))
	}
	return []byte(strings.Join(lines[start:end], "\n")), nil
}
