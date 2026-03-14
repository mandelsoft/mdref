package regexp_test

import (
	"strings"

	"github.com/mandelsoft/mdref/regexp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Expand Test Environment", func() {
	test := "data"
	one := "one"
	two := "two"
	data := []byte(test + ": this is a test with three matching groups: " + one + " and " + two)
	matches := []int{
		0,
		len(data),
		strings.Index(string(data), one),
		strings.Index(string(data), one) + len(one),
		strings.Index(string(data), two),
		strings.Index(string(data), two) + len(two),
		strings.Index(string(data), test),
		strings.Index(string(data), test) + len(test),
	}
	names := []string{
		"0",
		"g1",
		"g2",
		"g3",
	}

	cfg := regexp.NewSettings('(', ')', '/')
	Context("numbers", func() {
		Context("without braces", func() {
			It("substitute one group", func() {
				r := cfg.Expand([]byte{}, []byte("prefix$1suffix"), data, matches, names)
				Expect(string(r)).To(Equal("prefixonesuffix"))
			})
			It("substitute multiple groups", func() {
				r := cfg.Expand([]byte{}, []byte("prefix$1suffix$2"), data, matches, names)
				Expect(string(r)).To(Equal("prefixonesuffixtwo"))
			})
		})
		Context("with braces", func() {
			It("substitute one group", func() {
				r := cfg.Expand([]byte{}, []byte("prefix$(1)suffix"), data, matches, names)
				Expect(string(r)).To(Equal("prefixonesuffix"))
			})
			It("substitute multiple groups", func() {
				r := cfg.Expand([]byte{}, []byte("prefix$(1)suffix$(2)"), data, matches, names)
				Expect(string(r)).To(Equal("prefixonesuffixtwo"))
			})
		})
	})

	Context("names", func() {
		Context("without braces", func() {
			It("substitute one group", func() {
				r := cfg.Expand([]byte{}, []byte("prefix$g1-suffix"), data, matches, names)
				Expect(string(r)).To(Equal("prefixone-suffix"))
			})
			It("substitute multiple groups", func() {
				r := cfg.Expand([]byte{}, []byte("prefix$g1-suffix$g2"), data, matches, names)
				Expect(string(r)).To(Equal("prefixone-suffixtwo"))
			})
		})
		Context("with braces", func() {
			It("substitute one group", func() {
				r := cfg.Expand([]byte{}, []byte("prefix$(g1)suffix"), data, matches, names)
				Expect(string(r)).To(Equal("prefixonesuffix"))
			})
			It("substitute multiple groups", func() {
				r := cfg.Expand([]byte{}, []byte("prefix$(g1)suffix$(g2)"), data, matches, names)
				Expect(string(r)).To(Equal("prefixonesuffixtwo"))
			})
		})
	})

	Context("substitution", func() {
		It("with numbers", func() {
			r := cfg.Expand([]byte{}, []byte("prefix$(1/n/N)suffix"), data, matches, names)
			Expect(string(r)).To(Equal("prefixoNesuffix"))
		})
		It("empty", func() {
			r := cfg.Expand([]byte{}, []byte("prefix$(1/n/)suffix"), data, matches, names)
			Expect(string(r)).To(Equal("prefixoesuffix"))
		})
	})

	Context("invalid", func() {
		It("invalid /(", func() {
			r := cfg.Expand([]byte{}, []byte("prefix$(1/n/Nsuffix"), data, matches, names)
			Expect(string(r)).To(Equal("prefix$(1/n/Nsuffix"))
		})
		It("invalid (", func() {
			r := cfg.Expand([]byte{}, []byte("prefix$(1suffix"), data, matches, names)
			Expect(string(r)).To(Equal("prefix$(1suffix"))
		})
		It("invalid /", func() {
			r := cfg.Expand([]byte{}, []byte("prefix$(1/suffix"), data, matches, names)
			Expect(string(r)).To(Equal("prefix$(1/suffix"))
		})
	})

	Context("replacement", func() {
		It("replaces", func() {
			data := []byte("this string matches two strings")
			re, err := regexp.Compile("string")
			Expect(err).To(Succeed())
			r := re.ReplaceAllFunc(data, func(matches [][]byte) []byte {
				return []byte("STRING")
			})
			Expect(string(r)).To(Equal("this STRING matches two STRINGs"))
		})

		It("replace all", func() {
			data := []byte("two calls f(a) and g(b)")
			re, err := regexp.Compile("(.)\\(([^)]*)\\)", cfg)
			Expect(err).To(Succeed())
			r := re.ReplaceAll(data, []byte("function $1 with $2"))
			Expect(string(r)).To(Equal("two calls function f with a and function g with b"))
		})
	})

	Context("recursion", func() {
		It("handles recursive substitution", func() {
			data := []byte("two calls f(aa) and g(ab)")
			re, err := regexp.Compile("(.)\\(([^)]*)\\)", cfg)
			Expect(err).To(Succeed())
			r := re.ReplaceAll(data, []byte("function $1 with $(2/a/A)"))
			Expect(string(r)).To(Equal("two calls function f with AA and function g with Ab"))
		})
	})
})
