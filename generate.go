package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var caser = cases.Title(language.AmericanEnglish)

func generate(files []*File, resolution Resolution, source, target string) error {

	for _, f := range files {
		tgt := path.Join(target, f.relpath)
		src := path.Join(source, f.relpath)
		r, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("cannot read source %s", src, err)
		}
		os.MkdirAll(path.Dir(tgt), 0o766)
		w, err := os.OpenFile(tgt, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			r.Close()
			return fmt.Errorf("cannot read source %s", src, err)
		}
		if len(f.refs) == 0 && len(f.targets) == 0 {
			_, err := io.Copy(w, r)
			r.Close()
			w.Close()
			if err != nil {
				return fmt.Errorf("cannot read source %s", src, err)
			}
		} else {
			data, err := io.ReadAll(r)
			if err != nil {
				r.Close()
				w.Close()
				return fmt.Errorf("cannot read source %s", src, err)
			}

			for k := range f.terms {
				ref, term := resolution.Resolve(k, f.relpath)
				exp := regexp.MustCompile(`\[{{` + k + `}}\]`)
				data = exp.ReplaceAll(data, []byte(fmt.Sprintf("[%s](%s)", term, ref)))
				exp = regexp.MustCompile(`\[{{\*` + k + `}}\]`)
				data = exp.ReplaceAll(data, []byte(fmt.Sprintf("[%s](%s)", Plural(term), ref)))
				r := strings.ToUpper(term[:1]) + term[1:]
				k = strings.ToUpper(k[:1]) + k[1:]
				exp = regexp.MustCompile(`\[{{` + k + `}}\]`)
				data = exp.ReplaceAll(data, []byte(fmt.Sprintf("[%s](%s)", r, ref)))
				exp = regexp.MustCompile(`\[{{\*` + k + `}}\]`)
				data = exp.ReplaceAll(data, []byte(fmt.Sprintf("[%s](%s)", Plural(r), ref)))
			}

			for k := range f.refs {
				exp := regexp.MustCompile("({{" + k + "}})")

				ref, _ := resolution.Resolve(k, f.relpath)
				data = exp.ReplaceAll(data, []byte(ref))
			}

			for k, r := range f.targets {
				if r.generate {
					exp := regexp.MustCompile("{{" + k + "(:[a-zA-Z][a-zA-Z0-9- ]+)?}}")
					data = exp.ReplaceAll(data, []byte(`<a id="`+k+`"></a>`))
				} else {
					exp := regexp.MustCompile("\n?{{" + k + "(:[a-zA-Z][a-zA-Z0-9- ]+)?}}")
					data = exp.ReplaceAll(data, []byte(""))
				}
			}

			for k, c := range f.commands {
				exp := regexp.MustCompile(regexp.QuoteMeta(k))
				sub, err := c.GetSubstitution(src)
				if err != nil {
					return fmt.Errorf("%s: %s; %w", f.relpath, k, err)
				}
				data = exp.ReplaceAll(data, sub)
			}

			_, err = w.Write(data)
			r.Close()
			w.Close()
			if err != nil {
				return fmt.Errorf("cannot write target %s: %w", tgt, err)
			}
		}
	}
	return nil
}
