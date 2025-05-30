package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/mandelsoft/filepath/pkg/filepath"
)

type Position struct {
	line int
	col  int
}

func (r *Position) Position() string {
	return fmt.Sprintf("%d:%d", r.line, r.col)
}

type _Position = Position

type Ref struct {
	_Position
	text     string
	anchor   string
	generate bool
}

type Refs map[string]*Ref

type File struct {
	relpath  string
	refs     Refs
	terms    Refs
	targets  Refs
	commands Commands
}

func (f *File) HasSubst() bool {
	return !(len(f.refs) == 0 && len(f.targets) == 0 && len(f.terms) == 0 && len(f.commands) == 0)
}

func (f *File) Resolve(ref string, src string) (string, string) {
	tgt := f.targets[ref]
	if src == f.relpath {
		return "#" + tgt.anchor, tgt.text
	}

	ss := strings.Split(src, "/")
	ts := strings.Split(f.relpath, "/")
	for len(ss) > 0 {
		if len(ts) == 0 || ss[0] != ts[0] {
			break
		}
		ss = ss[1:]
		ts = ts[1:]
	}
	r := filepath.Join(ts...)
	for i := 1; i < len(ss); i++ {
		r = "../" + r
	}
	if tgt.anchor != "" {
		return r + "#" + tgt.anchor, tgt.text
	}
	return r, tgt.text
}

func scan(src, rel string, opts Options) ([]*File, error) {
	var result []*File

	list, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src, err)
	}
	for _, e := range list {
		if e.Name() == "local" {
			continue
		}
		rp := filepath.Join(rel, e.Name())
		ep := filepath.Join(src, e.Name())
		if e.IsDir() {
			r, err := scan(ep, rp, opts)
			if err != nil {
				return nil, err
			}
			result = append(result, r...)
		} else {
			if strings.HasSuffix(e.Name(), ".md") {
				refs, trms, tgts, cmds, err := scanRefs(ep, opts)
				if err != nil {
					return nil, err
				}
				result = append(result, &File{
					relpath:  rp,
					refs:     refs,
					terms:    trms,
					targets:  tgts,
					commands: cmds,
				})
			} else {
				result = append(result, &File{
					relpath: rp,
				})
			}
		}
	}

	return result, nil
}

var refExp = regexp.MustCompile(`\({{([a-z0-9.-]+)}}\)`)
var trmExp = regexp.MustCompile(`\[{{([*]?[A-Za-z][a-z0-9.-]*)}}\]`)
var tgtExp = regexp.MustCompile(`(?:^|[^([]){{([a-z][a-z0-9.-]*)(:([a-zA-Z][a-zA-Z0-9- ]+))?}}`)
var cmdExp = regexp.MustCompile(`{{([a-z]+)}((?:{[^}]+})+)}`)

func scanRefs(src string, opts Options) (Refs, Refs, Refs, Commands, error) {
	standard := map[string]struct{}{}
	refs := Refs{}
	trms := Refs{}
	targets := Refs{}

	data, err := os.ReadFile(src)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("cannot read %s: %w", src, err)
	}

	info := NewData(src, data)

	// reference substitutions
	matches, indices := info.scanFor(refExp)
	for i, m := range matches {
		key := string(m[1])
		pos := info.Position(indices[i][0])
		refs[key] = &Ref{_Position: pos}
	}

	// term substitutions
	matches, indices = info.scanFor(trmExp)
	for i, m := range matches {
		key := string(m[1])
		if strings.HasPrefix(key, "*") {
			key = key[1:]
		}
		key = strings.ToLower(key)
		pos := info.Position(indices[i][0])
		trms[key] = &Ref{_Position: pos}
	}

	// reference targets
	matches, indices = info.scanFor(tgtExp)
	for i, m := range matches {
		pos := info.Position(indices[i][0])
		key := string(m[1])
		if _, ok := targets[key]; ok {
			return nil, nil, nil, nil, fmt.Errorf("%s: %s: duplicate use of target %q", src, pos.Position(), key)
		}

		anchor, gen := key, true
		if opts.Headings {
			anchor, gen = determineAnchor(data, indices[i][0], indices[i][1], key)
			if !gen {
				if _, ok := standard[anchor]; ok {
					// similar heading used twice in document
					// fall back to anchor generation
					gen = true
					anchor = key
				} else {
					standard[anchor] = struct{}{}
				}
			}
		}
		ref := &Ref{
			_Position: pos,
			text:      string(m[3]),
			anchor:    anchor,
			generate:  gen,
		}
		targets[key] = ref
	}

	cmds := Commands{}

	// commands
	matches, indices = info.scanFor(cmdExp)
	for i, m := range matches {
		var cmd Command
		nl := false
		pos := info.Position(indices[i][0])
		key := string(m[1])
		if len(info.data) > indices[i][0]+len(m[0]) {
			nl = info.data[indices[i][0]+len(m[0])] == '\n'
		}
		switch key {
		case "include":
			cmd, err = NewInclude(pos, m[2], nl)
		case "execute":
			cmd, err = NewExecute(pos, m[2], nl)
		default:
			err = fmt.Errorf("invalid command %q", key)
		}
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%s: %s: %w", src, pos.Position(), err)
		}
		cmds[string(m[0])] = cmd
	}

	return refs, trms, targets, cmds, nil
}

type Data struct {
	relpath string
	data    []byte
	lines   []int
}

func NewData(p string, data []byte) *Data {
	lines := []int{0}

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			lines = append(lines, i+1)
		}
	}
	return &Data{p, data, lines}
}

func (l *Data) Position(idx int) Position {
	for n, i := range l.lines {
		if idx < i {
			return Position{n, idx - l.lines[n-1] + 1}
		}
	}
	return Position{-1, -1}
}

func (l *Data) Location(idx int) string {
	pos := l.Position(idx)
	return fmt.Sprintf("%s: %s", l.relpath, pos.Position())
}

func (l *Data) scanFor(exp *regexp.Regexp) ([][][]byte, [][]int) {
	return exp.FindAllSubmatch(l.data, -1), exp.FindAllIndex(l.data, -1)
}

func determineAnchor(data []byte, beg, end int, def string) (string, bool) {
	if data[beg] != '{' {
		beg++
	}
	var title []byte
	if len(data) > end+1 && data[end] == '\n' && data[end+1] == '#' {
		// before heading
		if beg == 0 {
			return "", false
		}
		s := end + 2
		for s < len(data) {
			if data[s] != '#' {
				break
			}
			s++
		}
		e := s
		for e < len(data) {
			if data[e] == '\n' {
				break
			}
			e++

		}
		if s < len(data) {
			title = data[s:e]
		}
	} else {
		if beg > 0 && data[beg-1] == '\n' {
			// possibly after heading
			e := beg - 1
			s := e
			for s > 0 {
				if data[s-1] == '\n' {
					break
				}
				s--
			}
			line := data[s:e]
			found := false
			for len(line) > 0 && line[0] == '#' {
				line = line[1:]
				found = true
			}
			if found {
				if s == 0 {
					return "", false
				}
				title = line
			}
		}
	}

	link := strings.ToLower(strings.TrimSpace(string(title)))
	if len(link) > 0 {
		l := ""
		for _, c := range link {
			if unicode.IsLetter(c) || unicode.IsDigit(c) {
				l += string(c)
			}
			if c == ' ' {
				l += "-"
			}
		}
		return l, false
	}
	return def, true
}
