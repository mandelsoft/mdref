package regexp

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Settings describes the syntactical elements for
// braces substitution variable and substitution patterns.
type Settings struct {
	lbrace rune
	rbrace rune
	subst  rune
}

var settings = Settings{lbrace: '{', rbrace: '}', subst: '/'}

func DefaultSettings() Settings {
	return settings
}

func NewSettings(l, r, s rune) Settings {
	return Settings{
		lbrace: l,
		rbrace: r,
		subst:  s,
	}
}

type Regexp struct {
	*regexp.Regexp
	runes Settings
}

func Compile(pattern string, braces ...Settings) (*Regexp, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	if len(braces) > 0 {
		return &Regexp{
			Regexp: re,
			runes:  braces[0],
		}, nil
	}
	return &Regexp{
		Regexp: re,
		runes:  DefaultSettings(),
	}, nil
}

// Expand appends template to dst and returns the result; during the
// append, Expand replaces variables in the template with corresponding
// matches drawn from src. The match slice should have been returned by
// [Regexp.FindSubmatchIndex].
//
// In the template, a variable is denoted by a substring of the form
// $name or ${name}, where name is a non-empty sequence of letters,
// digits, and underscores. A purely numeric name like $1 refers to
// the submatch with the corresponding index; other names refer to
// capturing parentheses named with the (?P<name>...) syntax. A
// reference to an out of range or unmatched index or a name that is not
// present in the regular expression is replaced with an empty slice.
//
// In the $name form, name is taken to be as long as possible: $1x is
// equivalent to ${1x}, not ${1}x, and, $10 is equivalent to ${10}, not ${1}0.
//
// To insert a literal $ in the output, use $$ in the template.
func (re *Regexp) Expand(dst []byte, template []byte, src []byte, match []int) []byte {
	return re.runes.expand(dst, string(template), src, "", match, re.SubexpNames())
}

func (re *Regexp) ExpandString(dst []byte, template string, src string, match []int) []byte {
	return re.runes.expand(dst, template, nil, src, match, re.SubexpNames())
}

func (b Settings) Expand(dst []byte, template []byte, src []byte, match []int, names []string) []byte {
	return b.expand(dst, string(template), src, "", match, names)
}

func (b Settings) ExpandString(dst []byte, template string, src string, match []int, names []string) []byte {
	return b.expand(dst, template, nil, src, match, names)
}

func (b Settings) expand(dst []byte, template string, bsrc []byte, src string, match []int, names []string) []byte {
	for len(template) > 0 {
		before, after, ok := strings.Cut(template, "$")
		if !ok {
			break
		}
		dst = append(dst, before...)
		template = after
		if template != "" && template[0] == '$' {
			// Treat $$ as $.
			dst = append(dst, '$')
			template = template[1:]
			continue
		}
		expr, rest, ok := b.extract(template)
		if !ok {
			// Malformed; treat $ as raw text.
			dst = append(dst, '$')
			continue
		}
		dst = append(dst, expr.Eval(bsrc, src, match, names)...)
		template = rest
	}
	dst = append(dst, template...)
	return dst
}

// extract returns the name from a leading "name" or "{name}" in str.
// (The $ has already been removed by the caller.)
// If it is a number, extract returns num set to that number; otherwise num = -1.
func (b Settings) extract(str string) (expr expr, rest string, ok bool) {
	if str == "" {
		return
	}
	brace := false
	rune, size := utf8.DecodeRuneInString(str)
	if rune == b.lbrace {
		brace = true
		str = str[size:]
	}
	i := 0
	isNum := false
	for i < len(str) {
		rune, size = utf8.DecodeRuneInString(str[i:])

		if i == 0 && !brace && unicode.IsDigit(rune) {
			isNum = true
		}
		if !(unicode.IsLetter(rune) && !isNum) && !unicode.IsDigit(rune) && rune != '_' {
			break
		}
		i += size
	}
	if i == 0 {
		// empty name is not okay
		return
	}
	name := str[:i]
	str = str[i:]
	i = 0
	first := 0
	second := 0
	if brace {
		if len(str) == 0 {
			// missing closing brace
			return
		}
		rune, size = utf8.DecodeRuneInString(str)
		switch rune {
		case b.rbrace:
			i += size
		case b.subst:
			i += size
			lvl := 0
			escaped := false
		loop:
			for i < len(str) {
				rune, size = utf8.DecodeRuneInString(str[i:])
				if escaped {
					escaped = false
				} else {
					switch rune {
					case '\\':
						escaped = true
					case '(':
						lvl++
					case b.rbrace:
						if lvl == 0 {
							if first == 0 {
								// invalid brace levels
								return
							}
							break loop
						}
						lvl--

					case b.subst:
						if first != 0 {
							// unexpected third /
							return
						}
						first = i
					}
				} // if
				i += size
			}
			if i >= len(str) {
				return
			}
			second = i
			i += size
		default:
			// missing closing brace
			return
		}
	}

	rest = str[i:]

	// Parse number.
	num := 0
	for i := 0; i < len(name); i++ {
		if name[i] < '0' || '9' < name[i] || num >= 1e8 {
			num = -1
			break
		}
		num = num*10 + int(name[i]) - '0'
	}
	// Disallow leading zeros.
	if name[0] == '0' && len(name) > 1 {
		num = -1
	}
	ok = true
	if num >= 0 {
		expr = &numExpr{num}
	} else {
		expr = &nameExpr{name}
	}

	// now we check for an expression.
	if first > 0 {
		rexp, err := regexp.Compile(str[1:first])
		if err != nil {
			return
		}
		expr = &substExpr{expr, &Regexp{rexp, b}, str[first+1 : second]}
	}
	return
}

type expr interface {
	Eval(bsrc []byte, src string, matches []int, names []string) []byte
}

type nameExpr struct {
	name string
}

func (e *nameExpr) Eval(bsrc []byte, src string, match []int, names []string) []byte {
	for i, namei := range names {
		if e.name == namei && 2*i+1 < len(match) && match[2*i] >= 0 {
			if bsrc != nil {
				return bsrc[match[2*i]:match[2*i+1]]
			} else {
				return []byte(src[match[2*i]:match[2*i+1]])
			}
			break
		}
	}
	return nil
}

type numExpr struct {
	num int
}

func (e *numExpr) Eval(bsrc []byte, src string, match []int, names []string) []byte {
	if 2*e.num+1 < len(match) && match[2*e.num] >= 0 {
		if bsrc != nil {
			return bsrc[match[2*e.num]:match[2*e.num+1]]
		} else {
			return []byte(src[match[2*e.num]:match[2*e.num+1]])
		}
	}
	return nil
}

type substExpr struct {
	expr  expr
	rexp  *Regexp
	subst string
}

func (e *substExpr) Eval(bsrc []byte, src string, match []int, names []string) []byte {
	data := e.expr.Eval(bsrc, src, match, names)
	return e.rexp.ReplaceAll(data, []byte(e.subst))
}

////////////////////////////////////////////////////////////////////////////////

func (re *Regexp) ReplaceAllIndexFunc(data []byte, fn func(data []byte, indices []int) []byte) []byte {
	matches := re.FindAllSubmatchIndex(data, -1)
	last := 0
	result := make([]byte, 0, len(data))
	for _, m := range matches {
		offset := m[0]
		result = append(result, data[last:offset]...)
		in := data[m[0]:m[1]]
		last = m[1]
		for i := range m {
			m[i] -= offset
		}
		r := fn(in, m)
		result = append(result, r...)
	}
	return append(result, data[last:]...)
}

func (re *Regexp) ReplaceAllFunc(data []byte, fn func(match [][]byte) []byte) []byte {
	return re.ReplaceAllIndexFunc(data, func(in []byte, indices []int) []byte {
		matches := make([][]byte, len(indices)/2)
		for i := 0; i < len(indices); i += 2 {
			matches[i/2] = data[indices[i]:indices[i+1]]
		}
		return fn(matches)
	})
}

func (re *Regexp) ReplaceAll(data []byte, template []byte) []byte {
	return re.ReplaceAllIndexFunc(data, func(in []byte, indices []int) []byte {
		return re.Expand([]byte{}, template, in, indices)
	})
}
