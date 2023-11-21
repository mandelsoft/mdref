package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

type Refs map[string]string

type File struct {
	relpath string
	refs    Refs
	terms   Refs
	targets Refs
}

func (f *File) Resolve(ref string, src string) (string, string) {
	if src == f.relpath {
		return "#" + ref, f.targets[ref]
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
	return r + "#" + ref, f.targets[ref]
}

func scan(src, rel string) ([]*File, error) {
	var result []*File

	list, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", src, err)
	}
	for _, e := range list {
		rp := path.Join(rel, e.Name())
		ep := path.Join(src, e.Name())
		if e.IsDir() {
			r, err := scan(ep, rp)
			if err != nil {
				return nil, err
			}
			result = append(result, r...)
		} else {
			if strings.HasSuffix(e.Name(), ".md") {
				refs, trms, tgts, err := scanRefs(ep)
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

func scanRefs(src string) (Refs, Refs, Refs, error) {
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
		refs[key] = ""
	}
	matches = trmExp.FindAllSubmatch(data, -1)
	for _, m := range matches {
		key := string(m[1])
		if strings.HasPrefix(key, "*") {
			key = key[1:]
		}
		key = strings.ToLower(key)
		// fmt.Printf("%s: found term ref %s\n", src, key)
		trms[key] = ""
	}
	matches = tgtExp.FindAllSubmatch(data, -1)
	for _, m := range matches {
		key := string(m[1])
		// fmt.Printf("%s: found ref %s[%s]\n", src, key, string(m[3]))
		if _, ok := targets[key]; ok {
			return nil, nil, nil, fmt.Errorf("duplicate use of target %s", key)
		}
		targets[key] = string(m[3])
	}
	return refs, trms, targets, nil
}
