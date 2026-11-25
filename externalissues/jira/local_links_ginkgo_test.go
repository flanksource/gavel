package jira

import (
	"os"
	"path/filepath"

	"github.com/flanksource/clicky/markdown"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("local Jira issue links", func() {
	It("resolves existing files across projects while preserving other Jira sites and missing targets", func() {
		root := GinkgoT().TempDir()
		for _, path := range []string{"10001/EXAMPLE-2.md", "10002/EXAMPLE-3.md", "10001/.sync-stage/EXAMPLE-4.md", "10001/EXAMPLE-1.attachments/EXAMPLE-5.md"} {
			Expect(os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, path), []byte("Existing issue"), 0o600)).To(Succeed())
		}
		files, err := indexIssueFiles(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(2))
		resolver := localLinkOptions{Site: "https://example.test", Path: filepath.Join(root, "10001", "EXAMPLE-1.md"), Files: files}
		for address, expected := range map[string]string{
			"https://example.test/browse/EXAMPLE-2":         "EXAMPLE-2.md",
			"https://example.test/browse/EXAMPLE-3#details": "../10002/EXAMPLE-3.md#details",
			"https://example.test/browse/EXAMPLE-4":         "https://example.test/browse/EXAMPLE-4",
			"https://other.test/browse/EXAMPLE-2":           "https://other.test/browse/EXAMPLE-2",
		} {
			actual, err := resolver.resolve(address)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(expected))
		}
		node := markdown.Node{Kind: "paragraph", Children: []markdown.Node{{Kind: "link", Href: "https://example.test/browse/EXAMPLE-2", Children: []markdown.Node{{Kind: "text", Text: "EXAMPLE-2"}}}}}
		resolved, err := resolveDocumentLinks(node, resolver.resolve)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Markdown()).To(Equal("[EXAMPLE-2](EXAMPLE-2.md)"))
		Expect(node.Children[0].Href).To(Equal("https://example.test/browse/EXAMPLE-2"))
	})

	It("rejects duplicate issue files and treats a new sync directory as empty", func() {
		root := GinkgoT().TempDir()
		files, err := indexIssueFiles(filepath.Join(root, "new"))
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(BeEmpty())
		for _, project := range []string{"10001", "10002"} {
			Expect(os.MkdirAll(filepath.Join(root, project), 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, project, "EXAMPLE-2.md"), nil, 0o600)).To(Succeed())
		}
		_, err = indexIssueFiles(root)
		Expect(err).To(MatchError(ContainSubstring("ambiguous local issue EXAMPLE-2")))
	})
})
