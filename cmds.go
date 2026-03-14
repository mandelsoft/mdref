package main

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	myregexp "github.com/mandelsoft/mdref/regexp"
)

type Command interface {
	Position() string
	EOL() bool
	GetSubstitution(path string, opts Options) ([]byte, error)
}

type Commands map[string]Command

////////////////////////////////////////////////////////////////////////////////

type Include struct {
	Empty
	file      string
	filter    *filter
	extractor extractor
}

type extractor interface {
	extract(data []byte) ([]byte, error)
}

type filter struct {
	filter *regexp.Regexp
	subst  []byte
}

func NewFilter(def []byte, subst []byte) (*filter, error) {
	var err error
	if def == nil {
		return nil, nil
	}

	fexp := FilterPattern[string(def)]
	if fexp == nil {
		s := html.UnescapeString(string(def))
		fexp, err = regexp.Compile(s)
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression %q: %w", s, err)
		}
	}

	return &filter{fexp, subst}, nil
}

func (i *Include) EOL() bool {
	return i.nl
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

var regexpcfg = myregexp.NewSettings('(', ')', '/')

// --- begin filter ---
// An optional additional such filter argument can be used to specify a
// filter regular expression. The selected
// file range is matched by this regular expression and
// the matched content of the all matches is
// concatenated. If the expression uses the multi-line mode, the matches
// are suffixed with a newline.
// If the expression conitains exactly one capturing group, the matched
// content for this group is taken.
// --- end filter ---

func (i *filter) Filter(data []byte) ([]byte, error) {
	if i == nil || i.filter == nil {
		return data, nil
	}
	sep := ""
	if strings.HasPrefix(i.filter.String(), "(?m)") {
		sep = "\n"
	}
	indices := i.filter.FindAllSubmatchIndex(data, -1)
	var result []byte
	for _, m := range indices {
		var r []byte
		if i.subst != nil {
			// use composition pattern for match
			r = regexpcfg.Expand([]byte{}, i.subst, data, m, i.filter.SubexpNames())
		} else {
			switch len(m) {
			case 2:
				// use complete match
				r = data[m[0]:m[1]]
			case 4:
				r = data[m[2]:m[3]]
			default:
				// multiple matching groups
				return nil, fmt.Errorf("multiple matching groups require a composion argument")
			}
		}
		result = append(result, r...)
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
var includeExpNum = regexp.MustCompile("^{([^}]+)}(?:{([0-9]+)?(?:(:)([0-9]+)?)?}(?:{([^}]+)}(?:{([^}]+)})?)?)?$")
var includeExpPat = regexp.MustCompile("^{([^}]+)}{([a-zA-Z][a-zA-Z0-9- ]*)}(?:{([^}]+)})?(?:{([^}]+)})?$")

// --- end include args ---
// --- end example ---

func NewInclude(pos Position, args []byte, nl bool) (Command, error) {
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
		var filter *filter
		filter, err = NewFilter(matches[5], matches[6])
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression: %w", err)
		}
		return &Include{Empty{pos, nl}, string(matches[1]), filter, &NumExtractor{int(start), int(end)}}, nil
	}

	matches = includeExpPat.FindSubmatch(args)
	if len(matches) != 0 {
		var filter *filter
		filter, err = NewFilter(matches[3], matches[4])
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression: %w", err)
		}
		return &Include{Empty{pos, nl}, string(matches[1]), filter, &PatternExtractor{string(matches[2])}}, nil
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
	Empty
	cmd       []string
	filter    *filter
	extractor extractor
	data      []byte
}

var _ Command = (*Execute)(nil)

func (e *Execute) EOL() bool {
	return e.nl
}

func (e *Execute) GetSubstitution(path string, opts Options) ([]byte, error) {

	if opts.SkipExecute {
		return []byte{}, nil
	}
	if e.data != nil {
		return e.data, nil
	}

	stderr := &bytes.Buffer{}
	cmd := exec.Command(e.cmd[0], e.cmd[1:]...)
	cmd.Dir = filepath.Dir(path)
	cmd.Stderr = stderr
	r, err := cmd.Output()
	if err != nil {
		if len(stderr.Bytes()) > 0 {
			return nil, fmt.Errorf("cannot execute %v: %w (%s)", e.cmd, err, stderr.String())
		}
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
	e.data = r
	return r, nil
}

var nextarg = regexp.MustCompile("^{([^}]+)}(.*)$")
var extractExpNum = regexp.MustCompile("^([0-9]+)?(?:(:)([0-9]+)?)?$")
var extractExpPat = regexp.MustCompile("^([a-zA-Z -]+)$")

func NewExecute(pos Position, args []byte, nl bool) (Command, error) {
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
	if len(extract) > 3 {
		return nil, fmt.Errorf("extraction mode requires a maximum of 3 arguments (found %d)", len(extract))
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

	var filter *filter
	switch len(extract) {
	case 2:
		filter, err = NewFilter([]byte(extract[1]), nil)
	case 3:
		filter, err = NewFilter([]byte(extract[1]), []byte(extract[2]))
	}
	if err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	return &Execute{Empty{pos, nl}, cmd, filter, ext, nil}, nil
}

////////////////////////////////////////////////////////////////////////////////

type Empty struct {
	_Position
	nl bool
}

var _ Command = (*Empty)(nil)

func (e *Empty) EOL() bool {
	return e.nl
}

func (e *Empty) GetSubstitution(path string, opts Options) ([]byte, error) {
	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////

var patternExp = regexp.MustCompile("^{([(a-z][a-z.-]*)}{([^}]+)}$")

func NewPattern(pos Position, args []byte, nl bool) (Command, error) {
	m := patternExp.FindSubmatch(args)
	if m == nil {
		return nil, fmt.Errorf("invalid pattern arguments")
	}

	name := string(m[1])
	pat := html.UnescapeString(string(m[2]))

	exp, err := regexp.Compile(pat)

	if err != nil {
		return nil, fmt.Errorf("invalid regular expression %q: %w", pat, err)
	}

	FilterPattern[name] = exp
	return &Empty{pos, true}, nil
}
