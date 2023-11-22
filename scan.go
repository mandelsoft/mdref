package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"
)

type Ref struct {
	text     string
	anchor   string
	generate bool
}

type Refs map[string]*Ref

type File struct {
	relpath string
	refs    Refs
	terms   Refs
	targets Refs
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
	r := path.Join(ts...)
	for i := 1; i < len(ss); i++ {
		r = "../" + r
	}
	return r + "#" + tgt.anchor, tgt.text
}

func scan(src, rel string, opts Options) ([]*File, error) {
	var result []*File

	list, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src, err)
	}
	for _, e := range list {
		rp := path.Join(rel, e.Name())
		ep := path.Join(src, e.Name())
		if e.IsDir() {
			r, err := scan(ep, rp, opts)
			if err != nil {
				return nil, err
			}
			result = append(result, r...)
		} else {
			if strings.HasSuffix(e.Name(), ".md") {
				refs, trms, tgts, err := scanRefs(ep, opts)
				if err != nil {
					return nil, err
				}
				result = append(result, &File{
					relpath: rp,
					refs:    refs,
					terms:   trms,
					targets: tgts,
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
var tgtExp = regexp.MustCompile(`[^([]{{([a-z][a-z0-9.-]*)(:([a-zA-Z][a-zA-Z0-9- ]+))?}}`)

func scanRefs(src string, opts Options) (Refs, Refs, Refs, error) {
	standard := map[string]struct{}{}
	refs := Refs{}
	trms := Refs{}
	targets := Refs{}

	data, err := os.ReadFile(src)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot read %s: %w", src, err)
	}

	matches := refExp.FindAllSubmatch(data, -1)
	for _, m := range matches {
		key := string(m[1])
		// fmt.Printf("%s: found ref %s\n", src, key)
		refs[key] = nil
	}
	matches = trmExp.FindAllSubmatch(data, -1)
	for _, m := range matches {
		key := string(m[1])
		if strings.HasPrefix(key, "*") {
			key = key[1:]
		}
		key = strings.ToLower(key)
		// fmt.Printf("%s: found term ref %s\n", src, key)
		trms[key] = nil
	}
	matches = tgtExp.FindAllSubmatch(data, -1)
	indices := tgtExp.FindAllIndex(data, -1)
	for i, m := range matches {
		key := string(m[1])
		// fmt.Printf("%s: found ref %s[%s]\n", src, key, string(m[3]))
		if _, ok := targets[key]; ok {
			return nil, nil, nil, fmt.Errorf("duplicate use of target %s", key)
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
			text:     string(m[3]),
			anchor:   anchor,
			generate: gen,
		}
		targets[key] = ref
	}
	return refs, trms, targets, nil
}

func determineAnchor(data []byte, beg, end int, def string) (string, bool) {
	if data[beg] != '{' {
		beg++
	}
	var title []byte
	if len(data) > end+1 && data[end] == '\n' && data[end+1] == '#' {
		// before heading
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
			for s > 1 {
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
