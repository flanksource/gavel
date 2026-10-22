package changegraph_test

import (
	"github.com/flanksource/gavel/internal/changegraph"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileSet", func() {
	It("deduplicates paths without stripping literal filename whitespace", func() {
		files := changegraph.NewFileSet()
		files.Add(" file\t\n.go ")
		files.Add("")
		files.Union(changegraph.FileSet{" file\t\n.go ": {}, "other.go": {}})
		Expect(files.Sorted()).To(Equal([]string{" file\t\n.go ", "other.go"}))
		Expect(files.Has(" file\t\n.go ")).To(BeTrue())
	})
})
