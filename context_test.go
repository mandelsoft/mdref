package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// docContext provides the document context for a document
// with the given source folder.
func docContext(doc, root string) *DocContext {
	return NewDocContext(doc, Options{DocRoot: root})
}

var _ = Describe("document context", func() {
	It("provides the document and the options", func() {
		opts := Options{DocRoot: "/src", SkipExecute: true, Headings: true}
		ctx := NewDocContext("/src/doc/a.md", opts)
		Expect(ctx.Document()).To(Equal("/src/doc/a.md"))
		Expect(ctx.DocRoot()).To(Equal("/src"))
		Expect(ctx.Options()).To(Equal(opts))
	})
})
