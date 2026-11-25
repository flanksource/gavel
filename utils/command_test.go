package utils

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type recordingTee struct {
	bytes.Buffer
	cwd  string
	argv []string
}

func (r *recordingTee) RecordCommand(cwd string, argv []string) {
	r.cwd, r.argv = cwd, argv
}

var _ = Describe("ShellQuote", func() {
	DescribeTable("quotes only what a shell would split or expand",
		func(in, want string) { Expect(ShellQuote(in)).To(Equal(want)) },
		Entry("plain", "plain", "plain"),
		Entry("dash", "with-dash", "with-dash"),
		Entry("relative path", "./pkg/a", "./pkg/a"),
		Entry("empty", "", "''"),
		Entry("space", "has space", "'has space'"),
		Entry("single quote", "has'quote", `'has'\''quote'`),
		Entry("variable", "$VAR", "'$VAR'"),
		Entry("pipe", "a|b", "'a|b'"),
		Entry("glob", "glob*.md", "'glob*.md'"),
		Entry("anchored regex", "^TestFoo$", "'^TestFoo$'"),
	)
})

var _ = Describe("ShellJoin", func() {
	It("quotes each argument that needs it", func() {
		Expect(ShellJoin([]string{"go", "test", "-json", "./pkg with space"})).To(Equal("go test -json './pkg with space'"))
	})
})

var _ = Describe("AnnounceCommand", func() {
	argv := []string{"ginkgo", "--focus", "Suite name"}

	It("writes the command line ahead of the output and records it on a CommandRecorder", func() {
		tee := &recordingTee{}
		AnnounceCommand(tee, "/repo/pkg", argv)
		Expect(tee.String()).To(Equal("▶ ginkgo --focus 'Suite name'\n"))
		Expect(tee.cwd).To(Equal("/repo/pkg"))
		Expect(tee.argv).To(Equal(argv))
	})

	It("only writes the command line to a plain writer", func() {
		var tee bytes.Buffer
		AnnounceCommand(&tee, "/repo/pkg", argv)
		Expect(tee.String()).To(Equal("▶ ginkgo --focus 'Suite name'\n"))
	})

	It("ignores a nil tee", func() {
		Expect(func() { AnnounceCommand(nil, "", argv) }).NotTo(Panic())
	})
})
