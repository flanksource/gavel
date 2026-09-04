package outline

import (
	"go/parser"
	"go/token"

	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("extractGinkgoTests", func() {
	var root *Entry

	find := func(parent *Entry, name string) *Entry {
		for _, child := range parent.Children {
			if child.Name == name {
				return child
			}
		}
		Fail("no child named " + name + " under " + parent.Name)
		return nil
	}

	BeforeEach(func() {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "testdata/ginkgo/sample_ginkgo_test.go", nil, parser.ParseComments)
		Expect(err).NotTo(HaveOccurred())
		entries, err := extractGinkgoTests(fset, file, "testdata/ginkgo/sample_ginkgo_test.go")
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		root = entries[0]
	})

	It("builds the container hierarchy with suite chains", func() {
		Expect(root.Name).To(Equal("Calculator"))
		Expect(root.Container).To(BeTrue())
		ctx := find(root, "when adding")
		Expect(ctx.Container).To(BeTrue())
		spec := find(ctx, "adds two positive numbers")
		Expect(spec.Container).To(BeFalse())
		Expect(spec.Framework).To(Equal(parsers.Ginkgo))
		Expect(spec.Suite).To(Equal([]string{"Calculator", "when adding"}))
	})

	It("measures spec closure size and complexity", func() {
		ctx := find(root, "when adding")
		simple := find(ctx, "adds two positive numbers")
		Expect(simple.SizeLines).To(Equal(3))
		Expect(simple.Complexity).To(Equal(1))
		branchy := find(ctx, "handles negatives")
		Expect(branchy.Complexity).To(Equal(2))
		Expect(branchy.calls).To(ContainElement("add"))
	})

	It("flags pending and dynamic specs", func() {
		ctx := find(root, "when adding")
		Expect(find(ctx, "rounds floats someday").Pending).To(BeTrue())
		dynamic := find(root, dynamicName)
		Expect(dynamic.Dynamic).To(BeTrue())
		Expect(dynamic.Container).To(BeFalse())
	})

	It("treats DescribeTable entries as leaves without closures", func() {
		table := find(root, "doubling")
		Expect(table.Container).To(BeTrue())
		zero := find(table, "zero")
		Expect(zero.Container).To(BeFalse())
		Expect(zero.Complexity).To(BeZero())
		Expect(zero.SizeLines).To(Equal(1))
		Expect(zero.Suite).To(Equal([]string{"Calculator", "doubling"}))
	})
})
