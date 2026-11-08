package main

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("scanning commands", func() {
	var dir string

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "src.go"), []byte(nested), 0o644)).To(Succeed())
	})

	scanDoc := func(content string) (Commands, error) {
		p := filepath.Join(dir, "doc.md")
		Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())
		_, _, _, cmds, err := scanRefs(NewDocContext(p, Options{}))
		return cmds, err
	}

	It("applies command options", func() {
		cmds, err := scanDoc("text\n{{include!}{src.go}{outer}}\n")
		Expect(err).To(Succeed())
		cmd := cmds["{{include!}{src.go}{outer}}"]
		Expect(cmd).NotTo(BeNil())
		data, err := cmd.GetSubstitution(docContext(filepath.Join(dir, "doc.md"), dir))
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("a\nb\nc"))
	})

	It("reports unsupported options with the command", func() {
		_, err := scanDoc("{{include!other}{src.go}{outer}}\n")
		Expect(err).To(MatchError(ContainSubstring(`command "include": unsupported option(s) "other" (supported: nomarkers)`)))
	})

	It("accepts a command at the beginning of a document", func() {
		cmds, err := scanDoc("{{include}{src.go}{inner}}\n")
		Expect(err).To(Succeed())
		Expect(cmds).To(HaveLen(1))
	})

	It("reports options for commands without options", func() {
		_, err := scanDoc("{{variable!}{name}{value}}\n")
		Expect(err).To(MatchError(ContainSubstring(`command "variable": no options supported`)))
		_, err = scanDoc("{{term!}{word}}\n")
		Expect(err).To(MatchError(ContainSubstring(`command "term": no options supported`)))
	})

	DescribeTable("determines heading anchors",
		func(text string, anchor string, isDefault bool, heading string) {
			data := []byte(text)
			// the anchor annotation {{x}} with the preceding and the following line end
			beg := strings.Index(text, "{{x}}")
			ok, beg, end := adjustAnchor(data, beg, beg+len("{{x}}"))
			Expect(ok).To(BeTrue())
			a, d, h := determineAnchor(data, beg, end, "x")
			Expect(a).To(Equal(anchor))
			Expect(d).To(Equal(isDefault))
			Expect(h).To(Equal(heading))
		},
		Entry("after a heading", "text\n\n## Second Heading\n{{x}}\nmore\n", "second-heading", false, ""),
		Entry("before a heading", "text\n\n{{x}}\n## Second Heading\nmore\n", "second-heading", false, ""),
		Entry("after the first heading (document)", "## First Heading\n{{x}}\nmore\n", "", false, "first-heading"),
		Entry("before the first heading (document)", "{{x}}\n## First Heading\nmore\n", "", false, "first-heading"),
		Entry("without heading", "text\n{{x}}\nmore\n", "x", true, ""),
	)

	Context("resolving references", func() {
		f := &File{
			relpath: "doc/chapters/commands.md",
			targets: Refs{
				"commands": &Ref{anchor: "", heading: "commands"},
				"include":  &Ref{anchor: "include"},
				"plain":    &Ref{anchor: ""},
			},
		}

		It("uses the heading anchor for document anchors in the same document", func() {
			r, _ := f.Resolve("commands", "doc/chapters/commands.md")
			Expect(r).To(Equal("#commands"))
		})

		It("refers to the document from other documents", func() {
			r, _ := f.Resolve("commands", "doc/chapters/terms.md")
			Expect(r).To(Equal("commands.md"))
			r, _ = f.Resolve("commands", "README.md")
			Expect(r).To(Equal("doc/chapters/commands.md"))
		})

		It("uses other anchors unchanged", func() {
			r, _ := f.Resolve("include", "doc/chapters/commands.md")
			Expect(r).To(Equal("#include"))
			r, _ = f.Resolve("include", "README.md")
			Expect(r).To(Equal("doc/chapters/commands.md#include"))
			r, _ = f.Resolve("plain", "doc/chapters/commands.md")
			Expect(r).To(Equal("#"))
		})
	})
})
