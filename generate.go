package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var caser = cases.Title(language.AmericanEnglish)

func comment(s string, l int) []byte {
	return []byte(fmt.Sprintf(fmt.Sprintf("<!-- %%-%ds -->\n", l), s))
}

func generate(files []*File, resolution Resolution, source, target string, opts Options) error {
	for _, f := range files {
		tgt := filepath.Join(target, f.relpath)
		src := filepath.Join(source, f.relpath)
		r, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("cannot read source %s", src, err)
		}
		os.MkdirAll(filepath.Dir(tgt), 0o766)
		w, err := os.OpenFile(tgt, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			r.Close()
			return fmt.Errorf("cannot read source %s", src, err)
		}
		if strings.HasSuffix(f.relpath, ".md") {
			rel, err := filepath.Rel(filepath.Dir(tgt), src)
			l := len(rel) + 5
			if l < 31 {
				l = 31
			}
			w.Write(comment("DO NOT MODIFY", l))
			w.Write(comment("this file is generated by mdref", l))
			if err == nil && !opts.SkipSource {
				w.Write(comment(fmt.Sprintf("from %s", rel), l))
			}
			w.Write([]byte("\n"))
		}
		if !f.HasSubst() {
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
				exp := regexp.MustCompile(`\({{` + k + `}}\)`)

				ref, _ := resolution.Resolve(k, f.relpath)
				data = exp.ReplaceAll(data, []byte("("+ref+")"))
			}

			for k, r := range f.targets {
				if r.generate {
					exp := regexp.MustCompile("{{" + k + "(:[a-zA-Z][a-zA-Z0-9- ]+)?}}")
					data = exp.ReplaceAll(data, []byte(`<a id="`+k+`"></a>`))
				} else {
					exp := regexp.MustCompile("{{" + k + "(:[a-zA-Z][a-zA-Z0-9- ]+)?}}\n?")
					data = exp.ReplaceAll(data, []byte(""))
				}
			}

			for k, c := range f.commands {
				exp := regexp.MustCompile(regexp.QuoteMeta(k))
				sub, err := c.GetSubstitution(src, opts)
				if err != nil {
					return fmt.Errorf("%s: %s: %s; %w", f.relpath, c.Position(), k, err)
				}
				if len(sub) > 0 && c.EOL() && sub[len(sub)-1] == '\n' {
					sub = sub[:len(sub)-1]
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
