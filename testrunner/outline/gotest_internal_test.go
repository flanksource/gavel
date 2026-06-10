package outline

import (
	"go/parser"
	"go/token"

	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var (
	Describe   = ginkgo.Describe
	It         = ginkgo.It
	BeforeEach = ginkgo.BeforeEach
	Fail       = ginkgo.Fail
)

func parseTestdataFile(path string) (*token.FileSet, *Entry, []*Entry) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	Expect(err).NotTo(HaveOccurred())
	entries := extractGoTests(fset, file, path)
	Expect(entries).NotTo(BeEmpty())
	return fset, entries[0], entries
}

var _ = Describe("extractGoTests", func() {
	var entries []*Entry
	byName := func(name string) *Entry {
		for _, e := range entries {
			if e.Name == name {
				return e
			}
		}
		Fail("no entry named " + name)
		return nil
	}

	BeforeEach(func() {
		_, _, entries = parseTestdataFile("testdata/gotest/sample_test.go")
	})

	It("finds only real Test funcs, skipping TestMain and helpers", func() {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name)
		}
		Expect(names).To(ConsistOf("TestAdd", "TestTable", "TestSubtests"))
	})

	It("captures location, size, and complexity for a simple test", func() {
		e := byName("TestAdd")
		Expect(e.Framework).To(Equal(parsers.GoTest))
		Expect(e.Line).To(Equal(8))
		Expect(e.EndLine).To(Equal(12))
		Expect(e.SizeLines).To(Equal(5))
		Expect(e.Complexity).To(Equal(2)) // one if
		Expect(e.calls).To(ContainElement("add"))
	})

	It("flags non-literal t.Run names as dynamic", func() {
		e := byName("TestTable")
		Expect(e.Children).To(HaveLen(1))
		Expect(e.Children[0].Dynamic).To(BeTrue())
		Expect(e.Children[0].Name).To(Equal("TestTable/<dynamic>"))
	})

	It("builds runtime-style names for nested literal subtests", func() {
		e := byName("TestSubtests")
		Expect(e.Children).To(HaveLen(2))
		child := e.Children[0]
		Expect(child.Name).To(Equal("TestSubtests/literal_child"))
		Expect(child.Children).To(HaveLen(1))
		Expect(child.Children[0].Name).To(Equal("TestSubtests/literal_child/nested_grandchild"))
		Expect(child.Children[0].Complexity).To(Equal(3)) // if with &&
		Expect(e.Children[1].Dynamic).To(BeTrue())
	})
})
