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
	file string
}

func (i *Include) getData(p string) ([]byte, error) {
	if !filepath.IsAbs(i.file) {
		p = filepath.Join(filepath.Dir(p), i.file)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("cannot read include file %q", i.file)
	}
	return data, nil
}

type IncludeNum struct {
	Include
	start int
	end   int
}

type IncludePat struct {
	Include
	pattern string
}

// --- begin example ---
// --- begin include args ---
var includeExpNum = regexp.MustCompile("^{([^}]+)}(?:{([0-9]+)?(?::([0-9]+)?)?})?$")
var includeExpPat = regexp.MustCompile("^{([^}]+)}{([a-zA-Z -]+)}$")

// --- end include args ---
// --- end example ---

func NewInclude(args []byte) (Command, error) {
	var err error

	matches := includeExpNum.FindSubmatch(args)
	if len(matches) != 0 {
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
		return &IncludeNum{Include{string(matches[1])}, int(start), int(end)}, nil
	}

	matches = includeExpPat.FindSubmatch(args)
	if len(matches) != 0 {
		return &IncludePat{Include{string(matches[1])}, string(matches[2])}, nil
	}

	return nil, fmt.Errorf("invalid include arguments %q", string(args))
}

func (i *IncludeNum) GetSubstitution(p string) ([]byte, error) {
	data, err := i.getData(p)
	if err != nil {
		return nil, err
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

func (i *IncludePat) GetSubstitution(p string) ([]byte, error) {
	data, err := i.getData(p)
	if err != nil {
		return nil, err
	}

	_, start, err := i.match(data, "begin")
	if err != nil {
		return nil, err
	}
	end, _, err := i.match(data, "end")
	if err != nil {
		return nil, err
	}
	return data[start:end], nil
}

func (i *IncludePat) match(data []byte, key string) (int, int, error) {
	exp := regexp.MustCompile(fmt.Sprintf("(?m)^.*--- %s %s ---.*$", key, regexp.QuoteMeta(i.pattern)))

	matches := exp.FindAllIndex(data, -1)
	if len(matches) == 0 {
		return -1, -1, fmt.Errorf("%s pattern (%s) not found in %q", key, i.pattern, i.file)
	}
	if len(matches) != 1 {
		return -1, -1, fmt.Errorf("%s pattern (%s) in %q is not unique", key, i.pattern, i.file)
	}

	start := matches[0][0]
	if start > 0 && data[start-1] == '\n' {
		start--
	}
	if start > 0 && data[start-1] == '\r' {
		start--
	}

	end := matches[0][1]
	if len(data) > end && data[end] == '\r' {
		end++
	}
	if len(data) > end && data[end] == '\n' {
		end++
	}
	return start, end, nil
}