package main

import (
	"bytes"
	"fmt"
	"html"
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

type Format struct {
	begin string
	end   string
}

var formats = map[string]*Format{
	"":  {"", ""},
	"*": {"**", "**"},
	"_": {"_", "_"},
	"`": {"<code>", "</code>"},
}

type Ref struct {
	_Position
	raw      []byte
	format   string
	text     string
	anchor   string
	generate bool
	// heading is the heading anchor of an anchor for the complete
	// document (empty anchor) located at the first heading of the
	// document. It is used for references from the same document.
	heading string
}

func (r *Ref) Raw() []byte {
	raw := r.raw
	if r.generate {
		return bytes.Trim(raw, "\n")
	}
	if raw[0] == '\n' && raw[len(raw)-1] == '\n' {
		return raw[:len(raw)-1]
	}
	return raw
}

func (r *Ref) Format(s string) string {
	if r == nil {
		return s
	}
	return formats[r.format].begin + mdescape(s) + formats[r.format].end
}

func (r *Ref) Plural(s string) string {
	if r == nil {
		return Plural(s)
	}
	if r.format == "`" {
		return r.Format(s) + "s"
	}
	return r.Format(Plural(s))
}

func (r *Ref) AsFormatted() string {
	if r == nil {
		return ""
	}
	return r.Format(r.text)
}

func (r *Ref) AsUpper() string {
	return strings.ToUpper(r.text[:1]) + r.text[1:]
}

func (r *Ref) AsPlural() string {
	if r == nil {
		return ""
	}
	return r.Plural(r.text)
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

func (f *File) Resolve(ref string, src string) (string, *Ref) {
	tgt := f.targets[ref]
	if src == f.relpath {
		if tgt.anchor == "" && tgt.heading != "" {
			// the document itself, referenced by its first heading
			return "#" + tgt.heading, tgt
		}
		return "#" + tgt.anchor, tgt
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
		return r + "#" + tgt.anchor, tgt
	}
	return r, tgt
}

func prescan(srcdir, rel string, opts Options) error {
	list, err := os.ReadDir(srcdir)
	if err != nil {
		return fmt.Errorf("%s: %w", srcdir, err)
	}
	for _, e := range list {
		if e.Name() == "local" {
			continue
		}
		rp := filepath.Join(rel, e.Name())
		ep := filepath.Join(srcdir, e.Name())
		if e.IsDir() {
			err := prescan(ep, rp, opts)
			if err != nil {
				return err
			}
		} else {
			if strings.HasSuffix(e.Name(), ".md") {
				err = scanDefs(NewDocContext(ep, opts))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func scanDefs(ctx *DocContext) error {
	srcfile := ctx.Document()
	data, err := os.ReadFile(srcfile)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", srcfile, err)
	}
	data, _ = normalizeLineEndings(data)
	info := NewData(srcfile, data)

	// commands
	matches, indices := info.scanFor(cmdExp)
	for i, m := range matches {
		ok := acceptAnchor(data, indices[i][0], indices[i][1])
		if !ok {
			continue
		}
		pos := info.Position(indices[i][0])
		key := string(m[1])

		switch key {
		case "pattern":
			_, err = NewPattern(pos, m[3], true, true)
		case "variable":
			_, err = NewVariable(pos, m[3], true, true)
		default:
		}
		if err != nil {
			return fmt.Errorf("%s: %s: %w", srcfile, pos.Position(), err)
		}
	}
	return nil
}

func scan(srcdir, rel string, opts Options) ([]*File, error) {
	var result []*File

	list, err := os.ReadDir(srcdir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", srcdir, err)
	}
	for _, e := range list {
		if e.Name() == "local" {
			continue
		}
		rp := filepath.Join(rel, e.Name())
		ep := filepath.Join(srcdir, e.Name())
		if e.IsDir() {
			r, err := scan(ep, rp, opts)
			if err != nil {
				return nil, err
			}
			result = append(result, r...)
		} else {
			if strings.HasSuffix(e.Name(), ".md") {
				refs, trms, tgts, cmds, err := scanRefs(NewDocContext(ep, opts))
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

func normalizeLineEndings(data []byte) ([]byte, bool) {
	r := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	return r, len(r) != len(data)
}

func toLineEndings(data []byte, win bool, opts Options) []byte {
	if (win && !opts.Unix) || opts.Windows {
		return bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
	}
	return data
}

func scanRefs(ctx *DocContext) (Refs, Refs, Refs, Commands, error) {
	srcfile := ctx.Document()
	standard := map[string]struct{}{}
	refs := Refs{}
	trms := Refs{}
	targets := Refs{}

	data, err := os.ReadFile(srcfile)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("cannot read %s: %w", srcfile, err)
	}
	data, _ = normalizeLineEndings(data)

	info := NewData(srcfile, data)

	// reference substitutions
	matches, indices := info.scanFor(refExp)
	for i, m := range matches {
		key := string(m[1])
		pos := info.Position(indices[i][0])
		refs[key] = &Ref{_Position: pos}
	}

	// term substitutions
	matches, indices = info.scanFor(subExp)
	for i, m := range matches {
		if !acceptMatch(data, indices[i][0], indices[i][1]) {
			continue
		}
		key := normKey(string(m[1]))
		pos := info.Position(indices[i][0])
		trms[key] = &Ref{_Position: pos}
	}

	// term links
	matches, indices = info.scanFor(lnkExp)
	for i, m := range matches {
		key := normKey(string(m[1]))
		pos := info.Position(indices[i][0])
		r := trms[key]
		if r != nil {
			r.generate = true
		} else {
			trms[key] = &Ref{_Position: pos, generate: true}
		}
	}

	// anchors
	matches, indices = info.scanFor(tgtExp)
	for i, m := range matches {
		ok, beg, end := adjustAnchor(data, indices[i][0], indices[i][1])
		if !ok {
			continue
		}
		pos := info.Position(indices[i][0])
		key := string(m[1])
		if _, ok := targets[key]; ok {
			return nil, nil, nil, nil, fmt.Errorf("%s: %s: duplicate use of target %q", srcfile, pos.Position(), key)
		}

		ref := createAnchor(ctx, data, pos, beg, end, key, string(m[3]), standard)
		targets[key] = ref
	}

	cmds := Commands{}

	// commands
	matches, indices = info.scanFor(cmdExp)
	for i, m := range matches {
		var cmd Command
		ok, beg, end := adjustAnchor(data, indices[i][0], indices[i][1])
		if !ok {
			continue
		}
		pos := info.Position(indices[i][0])
		key := string(m[1])
		options := parseOptions(m[2])

		switch key {
		case "variable":
			cmd, err = NewVariable(pos, m[3], data[end-1] == '\n', false)
		case "pattern":
			cmd, err = NewPattern(pos, m[3], data[end-1] == '\n', false)
		case "term":
			if options != nil {
				return nil, nil, nil, nil, fmt.Errorf("%s: %s: command %q: %w", srcfile, pos.Position(), key, errNoOptions)
			}
			var term string
			var ref *Ref
			term, ref, err = ParseTerm(ctx, pos, data, beg, end, m[3], standard)
			if ref != nil {
				if targets[term] != nil {
					return nil, nil, nil, nil, fmt.Errorf("%s: %s: duplicate use of term %q", srcfile, pos.Position(), term)
				}
				targets[term] = ref
				continue
			}
		default:
			gap := ""
			i := beg - 1
			for i >= 0 && (data[i] == ' ' || data[i] == '\t') {
				gap = string(data[i]) + gap
				i--
			}
			if i >= 0 && data[i] != '\n' {
				gap = ""
			}
			cmd, err = parseSubstCmd(key, pos, m[3], data[end-1] == '\n', gap)
		}
		if err == nil && options != nil {
			if err = cmd.ApplyOptions(options); err != nil {
				err = fmt.Errorf("command %q: %w", key, err)
			}
		}
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("%s: %s: %w", srcfile, pos.Position(), err)
		}
		cmds[strings.TrimSpace(string(m[0]))] = cmd
	}

	return refs, trms, targets, cmds, nil
}

func adjustAnchor(data []byte, beg, end int) (bool, int, int) {
	if !acceptAnchor(data, beg, end) {
		return false, beg, end
	}

	if beg > 0 {
		if data[beg-1] == '\n' {
			beg--
		}
	}
	if end < len(data) {
		if data[end] == '\n' {
			end++
		}
	}
	return true, beg, end
}

func acceptMatch(data []byte, beg, end int) bool {
	if data[beg] != '{' || beg == 0 {
		return true
	}
	if data[beg-1] == '{' {
		return false
	}
	if end >= len(data) {
		return true
	}
	if data[end] == '}' {
		return false
	}
	return true
}

func acceptAnchor(data []byte, beg, end int) bool {
	if beg == 0 {
		return true
	}
	match := string(data[beg:end])
	_ = match
	if data[beg-1] == '[' || data[beg-1] == '(' || data[beg-1] == '{' {
		return false
	}
	if end >= len(data) {
		return true
	}
	if data[end] == '}' {
		return false
	}
	return true
}

func parseSubstCmd(key string, pos Position, def []byte, nl bool, gap ...string) (Command, error) {
	var cmd Command
	var err error
	switch key {
	case "include":
		cmd, err = NewInclude(pos, def, nl, gap...)
	case "execute":
		cmd, err = NewExecute(pos, def, nl, gap...)
	default:
		err = fmt.Errorf("invalid command %q", key)
	}
	return cmd, err
}

func createAnchor(ctx *DocContext, data []byte, pos Position, begin, end int, key string, text string, standard map[string]struct{}) *Ref {
	anchor, gen, heading := key, true, ""
	if ctx.Options().Headings {
		anchor, gen, heading = determineAnchor(data, begin, end, key)
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

	return &Ref{
		_Position: pos,
		raw:       data[begin:end],
		text:      text,
		anchor:    anchor,
		generate:  gen,
		heading:   heading,
	}
}

func isEOL(data []byte, last int) bool {
	nl := false
	if len(data) > last {
		nl = data[last] == '\n'
	}
	return nl
}

func normKey(key string) string {
	if strings.HasPrefix(key, "*") {
		key = key[1:]
	}
	return strings.ToLower(key[0:1]) + key[1:]
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
	return Position{len(l.lines), idx - l.lines[len(l.lines)-1] + 1}
}

func (l *Data) Location(idx int) string {
	pos := l.Position(idx)
	return fmt.Sprintf("%s: %s", l.relpath, pos.Position())
}

func (l *Data) scanFor(exp *regexp.Regexp) ([][][]byte, [][]int) {
	return exp.FindAllSubmatch(l.data, -1), exp.FindAllIndex(l.data, -1)
}

func first(data []byte, pos int) bool {
	if pos == 0 && data[pos] == '{' {
		return true
	}
	return strings.TrimSpace(string(data[:pos])) == ""
}

// determineAnchor determines the anchor for an anchor annotation:
// the heading anchor of a heading directly before or after it,
// or the given default (generate=true). An anchor at the first
// heading of the document refers to the complete document: the
// anchor is empty and the heading anchor is provided separately.
func determineAnchor(data []byte, beg, end int, def string) (anchor string, generate bool, heading string) {
	var title []byte
	if len(data) > end && data[end-1] == '\n' && data[end] == '#' {
		// before heading
		document := first(data, beg)
		s := end + 1
		// skip section depth
		for s < len(data) {
			if data[s] != '#' {
				break
			}
			s++
		}
		e := s
		// look for end of title line
		for e < len(data) {
			if data[e] == '\n' {
				break
			}
			e++
		}
		if s < len(data) {
			title = data[s:e]
		}
		if document {
			return "", false, headingAnchor(title)
		}
	} else {
		if data[beg] == '\n' {
			// possibly after heading (beg is the end of the heading line)
			e := beg
			s := e
			// get previous line
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
				if first(data, s) {
					return "", false, headingAnchor(line)
				}
				title = line
			}
		}
	}

	if l := headingAnchor(title); l != "" {
		return l, false, ""
	}
	return def, true, ""
}

// headingAnchor provides the standard Markdown anchor for a heading title.
func headingAnchor(title []byte) string {
	l := ""
	for _, c := range strings.ToLower(strings.TrimSpace(string(title))) {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			l += string(c)
		}
		if c == ' ' {
			l += "-"
		}
	}
	return l
}

var termCmd = regexp.MustCompile("^{([`*_]?)([a-z][a-zA-Z0-9-_]+)}{([a-z]+)}((?:{[^{}]+})+)$")

// --- begin symbol ---

func ParseTerm(ctx *DocContext, pos Position, data []byte, begin, end int, args []byte, standard map[string]struct{}) (string, *Ref, error) {
	path, opts := ctx.Document(), ctx.Options()
	matches := termCmd.FindSubmatch(args)
	if len(matches) != 0 {
		qual := string(matches[1])
		name := strings.TrimSpace(string(matches[2]))
		key := string(matches[3])
		def := matches[4]

		switch qual {
		case "*":
			qual = "**"
		}
		incl, err := parseSubstCmd(key, pos, def, true)
		if err != nil {
			return "", nil, fmt.Errorf("term %q: %w", name, err)
		}
		// the text of a term is always required, even if
		// the execution of commands is skipped otherwise.
		topts := opts
		topts.SkipExecute = false
		tdata, err := incl.GetSubstitution(NewDocContext(path, topts))
		if err != nil {
			return "", nil, fmt.Errorf("cannot extract term %q: %w", name, err)
		}
		lines := strings.Split(string(tdata), "\n")
		text := strings.TrimSpace(lines[0])
		// fmt.Printf("identifier %q=%q\n", name, value)

		ref := createAnchor(ctx, data, pos, begin, end, name, text, standard)
		ref.format = qual
		return name, ref, nil
	}

	return "", nil, fmt.Errorf("invalid term arguments %q", string(args))
}

// --- end symbol ---

func mdescape(s string) string {
	s = html.EscapeString(s)
	s = strings.Replace(s, `*`, `&ast;`, -1)
	s = strings.Replace(s, `_`, `&lowbar;`, -1)
	s = strings.Replace(s, "`", `&grave;`, -1)
	return s
}
