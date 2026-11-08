package main

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// withVariables runs the test with the given variable
// definitions and restores the global state afterwards.
func withVariables(vars map[string]string) {
	oldVars := Variables
	Variables = map[string]string{}
	for k, v := range vars {
		Variables[k] = v
	}
	DeferCleanup(func() {
		Variables = oldVars
	})
}

var _ = Describe("variables", func() {
	root := filepath.Join(string(filepath.Separator), "work", "src")
	chapter := filepath.Join(root, "doc", "chapters", "commands.md")
	top := filepath.Join(root, "README.md")

	It("provides the source folder", func() {
		withVariables(nil)
		Expect(Subst("$(DOC_ROOT)/x", docContext(chapter, root))).To(Equal(root + "/x"))
	})

	It("provides the relative path to the source folder for the document", func() {
		withVariables(nil)
		Expect(Subst("$(REL_DIR_TO_DOC_ROOT)/x", docContext(chapter, root))).To(Equal("../../x"))
		Expect(Subst("$(REL_DIR_TO_DOC_ROOT)/x", docContext(top, root))).To(Equal("./x"))
	})

	Context("two stage evaluation", func() {
		BeforeEach(func() {
			withVariables(map[string]string{
				"name":    "demo",
				"cmd":     "$(name)cmd",
				"root":    "$(REL_DIR_TO_DOC_ROOT)/..",
				"democmd": "$(root)/$(cmd)",
				"both":    "$(DOC_ROOT):$(root)",
			})
		})

		Context("normalization", func() {
			It("resolves own variables completely", func() {
				Expect(NormalizeVariables()).To(Succeed())
				Expect(Variables["name"]).To(Equal("demo"))
				Expect(Variables["cmd"]).To(Equal("democmd"))
			})

			It("keeps references to predefined variables", func() {
				Expect(NormalizeVariables()).To(Succeed())
				Expect(Variables["root"]).To(Equal("$(REL_DIR_TO_DOC_ROOT)/.."))
				Expect(Variables["democmd"]).To(Equal("$(REL_DIR_TO_DOC_ROOT)/../democmd"))
				Expect(Variables["both"]).To(Equal("$(DOC_ROOT):$(REL_DIR_TO_DOC_ROOT)/.."))
			})

			It("reports unknown variables used by definitions", func() {
				Variables["broken"] = "$(REL_DIR_TO_DOC_ROOT)/$(unknown)"
				Expect(NormalizeVariables()).To(MatchError(`variable "broken": variable "unknown" not found`))
				Expect(Variables).NotTo(HaveKey("unknown"))
			})

			It("reports unknown variables before cycles", func() {
				Variables["a"] = "$(b)"
				Variables["b"] = "$(a)$(unknown)"
				Expect(NormalizeVariables()).To(MatchError(`variable "b": variable "unknown" not found`))
			})

			It("reports the first unknown variable in the order of the names", func() {
				Variables["y"] = "$(unknownb)"
				Variables["x"] = "$(unknowna)"
				Expect(NormalizeVariables()).To(MatchError(`variable "x": variable "unknowna" not found`))
			})
		})

		Context("substitution", func() {
			BeforeEach(func() {
				Expect(NormalizeVariables()).To(Succeed())
			})

			It("evaluates the predefined variables for the document using a variable", func() {
				Expect(Subst("$(democmd)", docContext(chapter, root))).To(Equal("../../../democmd"))
				Expect(Subst("$(democmd)", docContext(top, root))).To(Equal("./../democmd"))
				Expect(Subst("$(both)", docContext(chapter, root))).To(Equal(root + ":../../.."))
				Expect(Subst("$(both)", docContext(top, root))).To(Equal(root + ":./.."))
			})

			It("does not modify the normalized variables", func() {
				Expect(Subst("$(democmd)", docContext(chapter, root))).To(Equal("../../../democmd"))
				Expect(Variables["democmd"]).To(Equal("$(REL_DIR_TO_DOC_ROOT)/../democmd"))
				Expect(Subst("$(democmd)", docContext(top, root))).To(Equal("./../democmd"))
			})

			It("keeps values without predefined variables independent of the document", func() {
				Expect(Subst("$(cmd)", docContext(chapter, root))).To(Equal("democmd"))
				Expect(Subst("$(cmd)", docContext(top, root))).To(Equal("democmd"))
			})
		})
	})

	It("detects cycles", func() {
		withVariables(map[string]string{"a": "$(b)", "b": "$(a)$(DOC_ROOT)"})
		Expect(NormalizeVariables()).To(MatchError(ContainSubstring("cycle detected")))
	})

	It("supports names with digits", func() {
		withVariables(map[string]string{"lib2": "x", "path2": "$(lib2)/y"})
		Expect(NormalizeVariables()).To(Succeed())
		Expect(Subst("$(path2)", docContext(chapter, root))).To(Equal("x/y"))
	})

	It("supports names with underscores", func() {
		withVariables(map[string]string{"my_lib": "x", "my_path_2": "$(my_lib)/y"})
		Expect(NormalizeVariables()).To(Succeed())
		Expect(Subst("$(my_path_2)", docContext(chapter, root))).To(Equal("x/y"))
	})

	It("does not treat names starting with a digit or underscore as variables", func() {
		withVariables(nil)
		Expect(Subst("$(2lib)", docContext(chapter, root))).To(Equal("$(2lib)"))
		Expect(Subst("$(_lib)", docContext(chapter, root))).To(Equal("$(_lib)"))
	})

	It("reports unknown variables", func() {
		withVariables(nil)
		_, err := Subst("$(unknown)", docContext(chapter, root))
		Expect(err).To(MatchError(`variable "unknown" not found`))
	})
})
