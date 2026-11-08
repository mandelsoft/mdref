package main

// DocContext describes the context of the processing of a document
// (scanning it and evaluating its commands): the processed document and
// the generation options.
// Additional context information can be added without changing the
// signatures of the functions using the context.
type DocContext struct {
	document string
	options  Options
}

// NewDocContext provides the context for the processing of a document.
func NewDocContext(document string, opts Options) *DocContext {
	return &DocContext{document, opts}
}

// Document provides the path of the evaluated document.
func (c *DocContext) Document() string {
	return c.document
}

// DocRoot provides the absolute path of the source folder.
func (c *DocContext) DocRoot() string {
	return c.options.DocRoot
}

// Options provides the generation options.
func (c *DocContext) Options() Options {
	return c.options
}
