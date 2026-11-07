package main

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const nested = `before
// --- begin outer ---
a
// --- begin inner ---
b
/* --- end inner --- */
<!--- begin other key 2 --->
c
// --- end outer ---
after
`

var _ = Describe("commands", func() {
	var dir string

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "src.go"), []byte(nested), 0o644)).To(Succeed())
	})

	substitute := func(cmd Command) string {
		data, err := cmd.GetSubstitution(filepath.Join(dir, "doc.md"), Options{})
		Expect(err).To(Succeed())
		return string(data)
	}

	include := func(args string, options []string) Command {
		cmd, err := NewInclude(Position{}, []byte(args), false)
		Expect(err).To(Succeed())
		if options != nil {
			Expect(cmd.ApplyOptions(options)).To(Succeed())
		}
		return cmd
	}

	Context("include", func() {
		It("keeps nested markers without options", func() {
			Expect(substitute(include("{src.go}{outer}", nil))).To(Equal("a\n// --- begin inner ---\nb\n/* --- end inner --- */\n<!--- begin other key 2 --->\nc"))
			Expect(substitute(include("{src.go}{inner}", nil))).To(Equal("b"))
		})

		It("removes markers with the default options", func() {
			Expect(substitute(include("{src.go}{outer}", []string{}))).To(Equal("a\nb\nc"))
		})

		It("removes markers with option nomarkers", func() {
			Expect(substitute(include("{src.go}{outer}", []string{"nomarkers"}))).To(Equal("a\nb\nc"))
			Expect(substitute(include("{src.go}{outer}", []string{"nomarkers", "nomarkers"}))).To(Equal("a\nb\nc"))
		})

		It("removes markers for line ranges and complete files", func() {
			Expect(substitute(include("{src.go}{2:7}", []string{}))).To(Equal("a\nb\n"))
			Expect(substitute(include("{src.go}", []string{}))).To(Equal("before\na\nb\nc\nafter\n"))
		})

		It("keeps other lines with dashes", func() {
			Expect(os.WriteFile(filepath.Join(dir, "x.go"), []byte("// --- begin x ---\n--- a ---\n// --- end x ---\n"), 0o644)).To(Succeed())
			Expect(substitute(include("{x.go}{x}", []string{}))).To(Equal("--- a ---"))
		})

		It("applies the filter after removing the markers", func() {
			Expect(substitute(include("{src.go}{outer}{(?m)^(b|c)$}", []string{}))).To(Equal("b\nc\n"))
		})

		It("rejects unsupported options and reports all of them", func() {
			cmd, err := NewInclude(Position{}, []byte("{src.go}{outer}"), false)
			Expect(err).To(Succeed())
			Expect(cmd.ApplyOptions([]string{"other", "nomarkers", "more"})).To(MatchError(`unsupported option(s) "other", "more" (supported: nomarkers)`))
			// no option is applied
			Expect(substitute(cmd)).To(ContainSubstring("--- begin inner ---"))
		})
	})

	Context("execute", func() {
		execute := func(args string, options []string) Command {
			cmd, err := NewExecute(Position{}, []byte(args), false)
			Expect(err).To(Succeed())
			if options != nil {
				Expect(cmd.ApplyOptions(options)).To(Succeed())
			}
			return cmd
		}

		It("removes markers from the extracted output", func() {
			Expect(substitute(execute("{cat}{src.go}{<extract>}{outer}", nil))).To(ContainSubstring("--- begin inner ---"))
			Expect(substitute(execute("{cat}{src.go}{<extract>}{outer}", []string{}))).To(Equal("a\nb\nc"))
		})

		It("removes markers from the complete output", func() {
			Expect(substitute(execute("{cat}{src.go}", []string{"nomarkers"}))).To(Equal("before\na\nb\nc\nafter\n"))
		})

		It("accepts keys with digits like include", func() {
			Expect(os.WriteFile(filepath.Join(dir, "steps.txt"), []byte("// --- begin step 2 ---\nx\n// --- end step 2 ---\n"), 0o644)).To(Succeed())
			Expect(substitute(execute("{cat}{steps.txt}{<extract>}{step 2}", nil))).To(Equal("x"))
		})

		It("rejects unsupported options", func() {
			cmd, err := NewExecute(Position{}, []byte("{cat}{src.go}"), false)
			Expect(err).To(Succeed())
			Expect(cmd.ApplyOptions([]string{"other"})).To(MatchError(`unsupported option(s) "other" (supported: nomarkers)`))
		})

		It("rejects invalid keys", func() {
			_, err := NewExecute(Position{}, []byte("{cat}{file}{<extract>}{-outer}"), false)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("commands without options", func() {
		It("rejects options and the request for default options", func() {
			cmd, err := NewVariable(Position{}, []byte("{name}{value}"), false, false)
			Expect(err).To(Succeed())
			Expect(cmd.ApplyOptions([]string{})).To(MatchError(errNoOptions))
			Expect(cmd.ApplyOptions([]string{"nomarkers"})).To(MatchError("no options supported"))
		})
	})
})
