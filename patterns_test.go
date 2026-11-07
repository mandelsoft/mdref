package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("command syntax", func() {
	DescribeTable("parses commands",
		func(text string, name string, options []string, args string) {
			m := cmdExp.FindSubmatch([]byte(text))
			Expect(m).NotTo(BeNil())
			Expect(string(m[1])).To(Equal(name))
			Expect(parseOptions(m[2])).To(Equal(options))
			Expect(string(m[3])).To(Equal(args))
		},
		Entry("without options", "{{include}{a.go}{key}}", "include", nil, "{a.go}{key}"),
		Entry("default options", "{{include!}{a.go}{key}}", "include", []string{}, "{a.go}{key}"),
		Entry("one option", "{{include!nomarkers}{a.go}}", "include", []string{"nomarkers"}, "{a.go}"),
		Entry("option list", "{{execute!a,b-c,d2}{cmd}}", "execute", []string{"a", "b-c", "d2"}, "{cmd}"),
	)

	DescribeTable("rejects invalid option lists",
		func(text string) {
			Expect(cmdExp.FindSubmatch([]byte(text))).To(BeNil())
		},
		Entry("upper case", "{{include!Bad}{a.go}}"),
		Entry("empty element", "{{include!a,}{a.go}}"),
		Entry("leading digit", "{{include!1a}{a.go}}"),
	)
})
