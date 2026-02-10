package parsers

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTreeBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tree Builder Suite")
}

var _ = Describe("BuildTestTreeWithVerbosity", func() {
	var tests []Test

	BeforeEach(func() {
		// Create a mix of passing, failing, and skipped tests
		tests = []Test{
			{
				Name:        "TestPassing1",
				Package:     "pkg1",
				PackagePath: "path/to/pkg1",
				File:        "path/to/pkg1/suite1_test.go",
				Suite:       []string{"Suite1"},
				Passed:      true,
				Duration:    100 * time.Millisecond,
			},
			{
				Name:        "TestPassing2",
				Package:     "pkg1",
				PackagePath: "path/to/pkg1",
				File:        "path/to/pkg1/suite1_test.go",
				Suite:       []string{"Suite1"},
				Passed:      true,
				Duration:    150 * time.Millisecond,
			},
			{
				Name:        "TestFailing",
				Package:     "pkg1",
				PackagePath: "path/to/pkg1",
				File:        "path/to/pkg1/suite1_test.go",
				Suite:       []string{"Suite1"},
				Failed:      true,
				Duration:    200 * time.Millisecond,
				Message:     "assertion failed",
			},
			{
				Name:        "TestSkipped",
				Package:     "pkg1",
				PackagePath: "path/to/pkg1",
				File:        "path/to/pkg1/suite2_test.go",
				Suite:       []string{"Suite2"},
				Skipped:     true,
				Duration:    0,
			},
			{
				Name:        "TestPassingAlone",
				Package:     "pkg2",
				PackagePath: "path/to/pkg2",
				File:        "path/to/pkg2/suite3_test.go",
				Suite:       []string{"Suite3"},
				Passed:      true,
				Duration:    50 * time.Millisecond,
			},
		}
	})

	Context("at verbosity 0 (default)", func() {
		It("should build directory tree structure", func() {
			tree := BuildTestTree(tests)

			// Find the path/to/ directory
			Expect(tree).To(HaveLen(1))
			Expect(tree[0].Name).To(Equal("path/"))

			// Navigate to pkg1
			toDir := findChildByName(tree[0].Children, "to/")
			Expect(toDir).NotTo(BeNil())
			pkg1Dir := findChildByName(toDir.Children, "pkg1/")
			Expect(pkg1Dir).NotTo(BeNil())

			// Should have file nodes under pkg1
			suite1File := findChildByName(pkg1Dir.Children, "suite1_test.go")
			Expect(suite1File).NotTo(BeNil(), "Should have suite1_test.go file node")
			suite2File := findChildByName(pkg1Dir.Children, "suite2_test.go")
			Expect(suite2File).NotTo(BeNil(), "Should have suite2_test.go file node")
		})

		It("should group tests by file then by suite", func() {
			tree := BuildTestTree(tests)

			// Navigate to suite1_test.go
			toDir := findChildByName(tree[0].Children, "to/")
			pkg1Dir := findChildByName(toDir.Children, "pkg1/")
			suite1File := findChildByName(pkg1Dir.Children, "suite1_test.go")
			Expect(suite1File).NotTo(BeNil())

			// Should have Suite1 node under the file
			suite1Node := findChildByName(suite1File.Children, "Suite1")
			Expect(suite1Node).NotTo(BeNil(), "Should have Suite1 node under file")

			// Suite1 should have 3 tests (TestPassing1, TestPassing2, TestFailing)
			Expect(len(suite1Node.Children)).To(Equal(3))
		})

		It("should organize tests from different files separately", func() {
			tree := BuildTestTree(tests)

			// Navigate to pkg1
			toDir := findChildByName(tree[0].Children, "to/")
			pkg1Dir := findChildByName(toDir.Children, "pkg1/")

			// suite1_test.go should have Suite1
			suite1File := findChildByName(pkg1Dir.Children, "suite1_test.go")
			Expect(findChildByName(suite1File.Children, "Suite1")).NotTo(BeNil())

			// suite2_test.go should have Suite2 (with the skipped test)
			suite2File := findChildByName(pkg1Dir.Children, "suite2_test.go")
			Expect(findChildByName(suite2File.Children, "Suite2")).NotTo(BeNil())
		})
	})

	Context("at verbosity 1", func() {
		It("should show directory and suite structure with counts", func() {
			tree := BuildTestTreeWithVerbosity(tests, 1)

			// Should have directory nodes
			Expect(tree).NotTo(BeEmpty())

			// Should NOT show individual leaf tests
			leafCount := countLeafTests(tree)
			Expect(leafCount).To(Equal(0), "Should not show individual tests at verbosity 1")
		})
	})

	Context("at verbosity 2", func() {
		It("should show all tests including passing ones", func() {
			// First check the raw tree before verbosity filtering
			rawTree := BuildTestTree(tests)
			rawCount := countLeafTests(rawTree)

			tree := BuildTestTreeWithVerbosity(tests, 2)

			// Should show all 5 tests
			totalShown := countLeafTests(tree)
			Expect(totalShown).To(Equal(5), "Should show all tests at verbosity 2 (raw tree had %d tests)", rawCount)
		})
	})
})

// Helper function to find a child by name
func findChildByName(children []Test, name string) *Test {
	for i := range children {
		if children[i].Name == name {
			return &children[i]
		}
	}
	return nil
}

// Helper function to find a child by name recursively
func findChildByNameRecursive(node Test, name string) *Test {
	if node.Name == name {
		return &node
	}
	for i := range node.Children {
		if found := findChildByNameRecursive(node.Children[i], name); found != nil {
			return found
		}
	}
	return nil
}

// Helper function to count leaf tests in tree
func countLeafTests(nodes []Test) int {
	count := 0
	for _, node := range nodes {
		if len(node.Children) == 0 && !isContainerNode(node) {
			count++
		} else {
			count += countLeafTests(node.Children)
		}
	}
	return count
}

var _ = Describe("makeFileRelativeToPackage", func() {
	DescribeTable("extracts relative file path correctly",
		func(filePath, pkgPath, expected string) {
			result := makeFileRelativeToPackage(filePath, pkgPath)
			Expect(result).To(Equal(expected))
		},
		Entry("absolute path with pkg in middle",
			"/Users/moshe/project/path/to/pkg1/test_test.go",
			"path/to/pkg1",
			"test_test.go"),
		Entry("absolute path - fallback to filename",
			"/Users/moshe/other/project/test_test.go",
			"path/to/pkg1",
			"test_test.go"),
		Entry("relative path with ../ prefix",
			"../other/path/to/pkg1/test_test.go",
			"path/to/pkg1",
			"test_test.go"),
		Entry("relative path matching package prefix",
			"path/to/pkg1/test_test.go",
			"path/to/pkg1",
			"test_test.go"),
		Entry("relative path already stripped",
			"test_test.go",
			"path/to/pkg1",
			"test_test.go"),
		Entry("relative path with subdirectory",
			"path/to/pkg1/subdir/test_test.go",
			"path/to/pkg1",
			"subdir/test_test.go"),
		Entry("empty file path",
			"",
			"path/to/pkg1",
			""),
		Entry("empty package path with absolute file",
			"/Users/moshe/project/test_test.go",
			"",
			"test_test.go"),
		Entry("empty package path with relative file",
			"some/path/test_test.go",
			"",
			"some/path/test_test.go"),
	)
})
