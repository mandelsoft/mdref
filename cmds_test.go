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
		data, err := cmd.GetSubstitution(docContext(filepath.Join(dir, "doc.md"), dir))
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

	Context("variables", func() {
		It("includes files given by DOC_ROOT and REL_DIR_TO_DOC_ROOT", func() {
			sub := filepath.Join(dir, "doc", "chapters")
			Expect(os.MkdirAll(sub, 0o755)).To(Succeed())
			withVariables(nil)
			doc := filepath.Join(sub, "x.md")
			for _, args := range []string{"{$(DOC_ROOT)/src.go}{inner}", "{$(REL_DIR_TO_DOC_ROOT)/src.go}{inner}"} {
				cmd, err := NewInclude(Position{}, []byte(args), false)
				Expect(err).To(Succeed())
				data, err := cmd.GetSubstitution(docContext(doc, dir))
				Expect(err).To(Succeed(), args)
				Expect(string(data)).To(Equal("b"), args)
			}
		})

		It("substitutes execute arguments for the document", func() {
			sub := filepath.Join(dir, "doc")
			Expect(os.MkdirAll(sub, 0o755)).To(Succeed())
			withVariables(nil)
			cmd, err := NewExecute(Position{}, []byte("{cat}{$(REL_DIR_TO_DOC_ROOT)/src.go}{<extract>}{inner}"), false)
			Expect(err).To(Succeed())
			data, err := cmd.GetSubstitution(docContext(filepath.Join(sub, "x.md"), dir))
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("b"))
		})

		It("accepts variable names with digits", func() {
			withVariables(nil)
			_, err := NewVariable(Position{}, []byte("{lib2}{x}"), false, true)
			Expect(err).To(Succeed())
			Expect(Variables).To(HaveKeyWithValue("lib2", "x"))
		})

		It("accepts variable names with underscores", func() {
			withVariables(nil)
			_, err := NewVariable(Position{}, []byte("{my_lib}{x}"), false, true)
			Expect(err).To(Succeed())
			Expect(Variables).To(HaveKeyWithValue("my_lib", "x"))
		})

		It("rejects variable names starting with a digit or underscore", func() {
			withVariables(nil)
			_, err := NewVariable(Position{}, []byte("{2lib}{x}"), false, true)
			Expect(err).To(MatchError(`invalid variable arguments "{2lib}{x}"`))
			_, err = NewVariable(Position{}, []byte("{_lib}{x}"), false, true)
			Expect(err).To(MatchError(`invalid variable arguments "{_lib}{x}"`))
		})

		It("rejects the definition of predefined variables", func() {
			withVariables(nil)
			_, err := NewVariable(Position{}, []byte("{DOC_ROOT}{x}"), false, true)
			Expect(err).To(MatchError(`variable name "DOC_ROOT" is reserved for a predefined variable`))
		})
	})
})
