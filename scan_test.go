package main

import (
	"os"
	"path/filepath"

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
		_, _, _, cmds, err := scanRefs(p, Options{})
		return cmds, err
	}

	It("applies command options", func() {
		cmds, err := scanDoc("text\n{{include!}{src.go}{outer}}\n")
		Expect(err).To(Succeed())
		cmd := cmds["{{include!}{src.go}{outer}}"]
		Expect(cmd).NotTo(BeNil())
		data, err := cmd.GetSubstitution(filepath.Join(dir, "doc.md"), Options{})
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
})
