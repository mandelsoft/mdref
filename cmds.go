package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Command interface {
	Position() string
	GetSubstitution(path string, opts Options) ([]byte, error)
}

type Commands map[string]Command

////////////////////////////////////////////////////////////////////////////////

type Include struct {
	_Position
	file      string
	filter    *filter
	extractor extractor
}

type extractor interface {
	extract(data []byte) ([]byte, error)
}

type filter struct {
	filter *regexp.Regexp
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

// --- begin filter ---
// An optional third argument can be used to specify a filter regular
// expression. It must contain one matching group. The
// selected file range is matched by this regular expression and
// the content of the first matching group of the all matches is
// concatenated. If the expression uses the multi-line mode, the matches
// are suffixed with a newline.
// --- end filter ---

func (i *filter) Filter(data []byte) ([]byte, error) {
	if i.filter == nil {
		return data, nil
	}
	sep := ""
	if strings.HasPrefix(i.filter.String(), "(?m)") {
		sep = "\n"
	}
	matches := i.filter.FindAllSubmatch(data, -1)
	var result []byte
	for _, m := range matches {
		if len(m) != 2 {
			return nil, fmt.Errorf("regular expression must contain one matching group")
		}
		result = append(result, m[1]...)
		result = append(result, []byte(sep)...)
	}
	return result, nil
}

func (i *Include) GetSubstitution(p string, opts Options) ([]byte, error) {
	data, err := i.getData(p)
	if err != nil {
		return nil, err
	}

	data, err = i.extractor.extract(data)
	if err != nil {
		return nil, fmt.Errorf("include file %q: %w", i.file, err)
	}
	return i.filter.Filter(data)
}

type NumExtractor struct {
	start int
	end   int
}

type PatternExtractor struct {
	pattern string
}

// --- begin example ---
// --- begin include args ---
var includeExpNum = regexp.MustCompile("^{([^}]+)}(?:{([0-9]+)?(?:(:)([0-9]+)?)?}(?:{(.*)})?)?$")
var includeExpPat = regexp.MustCompile("^{([^}]+)}{([a-zA-Z][a-zA-Z0-9 -]*)}(?:{(.*)})?$")

// --- end include args ---
// --- end example ---

func NewInclude(line, col int, args []byte) (Command, error) {
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
			end = start
		}
		if matches[3] != nil {
			if matches[4] != nil {
				end, err = strconv.ParseInt(string(matches[4]), 10, 32)
				if err != nil {
					return nil, fmt.Errorf("invalid start line: %w", err)
				}
			} else {
				end = 0
			}
		}
		var fexp *regexp.Regexp
		if matches[5] != nil {
			fexp, err = regexp.Compile(string(matches[5]))
			if err != nil {
				return nil, fmt.Errorf("invalid filter expression: %w", err)
			}
		}
		return &Include{Position{line, col}, string(matches[1]), &filter{fexp}, &NumExtractor{int(start), int(end)}}, nil
	}

	matches = includeExpPat.FindSubmatch(args)
	if len(matches) != 0 {
		var fexp *regexp.Regexp
		if matches[3] != nil {
			fexp, err = regexp.Compile(string(matches[3]))
			if err != nil {
				return nil, fmt.Errorf("invalid filter expression: %w", err)
			}
		}
		return &Include{Position{line, col}, string(matches[1]), &filter{fexp}, &PatternExtractor{string(matches[2])}}, nil
	}

	return nil, fmt.Errorf("invalid include arguments %q", string(args))
}

func (i *NumExtractor) extract(data []byte) ([]byte, error) {
	lines := strings.Split(string(data), "\n")
	start := 0
	if i.start > 0 {
		start = i.start - 1
	}
	if start >= len(lines) {
		return nil, fmt.Errorf("start line %d after end of data (%d lines)", start, len(lines))
	}
	end := len(lines)
	if i.end > 0 {
		end = i.end
	}
	if end > len(lines) {
		return nil, fmt.Errorf("end line %d after end of file (%d lines)", end, len(lines))
	}
	return []byte(strings.Join(lines[start:end], "\n")), nil
}

func (i *PatternExtractor) extract(data []byte) ([]byte, error) {
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

func (i *PatternExtractor) match(data []byte, key string) (int, int, error) {
	exp := regexp.MustCompile(fmt.Sprintf("(?m)^.*--- %s %s ---.*$", key, regexp.QuoteMeta(i.pattern)))

	matches := exp.FindAllIndex(data, -1)
	if len(matches) == 0 {
		return -1, -1, fmt.Errorf("%s pattern (%s) not found", key, i.pattern)
	}
	if len(matches) != 1 {
		return -1, -1, fmt.Errorf("%s pattern (%s) is not unique", key, i.pattern)
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

////////////////////////////////////////////////////////////////////////////////

type Execute struct {
	_Position
	cmd       []string
	filter    *filter
	extractor extractor
}

var _ Command = (*Execute)(nil)

func (e *Execute) GetSubstitution(path string, opts Options) ([]byte, error) {

	if opts.SkipExecute {
		return []byte{}, nil
	}
	cmd := exec.Command(e.cmd[0], e.cmd[1:]...)
	cmd.Dir = filepath.Dir(path)
	r, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot execute %v: %w", e.cmd, err)
	}
	if e.extractor != nil {
		r, err = e.extractor.extract(r)
		if err != nil {
			return nil, fmt.Errorf("extract failed %v: %w", e.cmd, err)
		}
	}
	if e.filter != nil {
		r, err = e.filter.Filter(r)
		if err != nil {
			return nil, fmt.Errorf("extract failed %v: %w", e.cmd, err)
		}
	}
	return r, nil
}

var nextarg = regexp.MustCompile("^{([^}]+)}(.*)$")
var extractExpNum = regexp.MustCompile("^([0-9]+)?(?:(:)([0-9]+)?)?$")
var extractExpPat = regexp.MustCompile("^([a-zA-Z -]+)$")

func NewExecute(line, col int, args []byte) (Command, error) {
	var cmd []string

	for {
		m := nextarg.FindSubmatch(args)
		if m == nil {
			break
		}
		cmd = append(cmd, string(m[1]))
		args = m[2]
	}

	var extract []string
	for i := range cmd {
		if cmd[i] == "<extract>" {
			extract = cmd[i+1:]
			cmd = cmd[:i]
			break
		}
	}
	if len(cmd) == 0 {
		return nil, fmt.Errorf("command argument required")
	}
	if len(extract) > 2 {
		return nil, fmt.Errorf("extraction mode requires a maximum of 2 arguments (found %d)", len(extract))
	}

	var ext extractor
	var err error
	if len(extract) > 0 {
		m := extractExpNum.FindSubmatch([]byte(extract[0]))
		if m != nil {
			start := int64(0)
			end := int64(0)
			if m[1] != nil {
				start, err = strconv.ParseInt(string(m[1]), 10, 32)
				if err != nil {
					return nil, fmt.Errorf("invalid start line: %w", err)
				}
				end = start
			}
			if m[2] != nil {
				if m[3] != nil {
					end, err = strconv.ParseInt(string(m[3]), 10, 32)
					if err != nil {
						return nil, fmt.Errorf("invalid start line: %w", err)
					}
				} else {
					end = 0
				}
			}
			ext = &NumExtractor{
				start: int(start),
				end:   int(end),
			}
		} else {
			m = extractExpPat.FindSubmatch([]byte(extract[0]))
			if m == nil {
				return nil, fmt.Errorf("invalid range specification (%s)", extract[0])
			}
			ext = &PatternExtractor{extract[0]}
		}
	}

	var fexp *regexp.Regexp
	if len(extract) == 2 {
		fexp, err = regexp.Compile(string(extract[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression: %w", err)
		}
	}
	return &Execute{Position{line, col}, cmd, &filter{fexp}, ext}, nil
}
