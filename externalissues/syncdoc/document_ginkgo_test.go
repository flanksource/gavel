package syncdoc

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

func TestDocument(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "sync documents") }

var _ = Describe("sync frontmatter", func() {
	It("preserves unknown nested enrichment data and rejects concurrent edits", func() {
		dir := GinkgoT().TempDir()
		root, err := os.OpenRoot(dir)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(root.Close)
		original := []byte("---\ncustom: retained\ngavel:\n  version: 1\n  attachments:\n    '20001':\n      name: report.pdf\n      enrichments:\n        ocr:\n          source_sha256: abc\n          result:\n            count: 9007199254740993\n          model: example\n---\nBody\n")
		Expect(root.WriteFile("issue.md", original, 0600)).To(Succeed())
		doc, err := Read(root, "issue.md")
		Expect(err).NotTo(HaveOccurred())
		encoded, err := doc.Markdown("Updated")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(SatisfyAll(ContainSubstring("retained"), ContainSubstring("9007199254740993"), ContainSubstring("model: example")))
		Expect(doc.CheckUnchanged(root, "issue.md")).To(Succeed())
		Expect(root.WriteFile("issue.md", encoded, 0600)).To(Succeed())
		Expect(doc.CheckUnchanged(root, "issue.md")).To(HaveOccurred())
		var node yaml.Node
		Expect(yaml.Unmarshal(encoded[:len(encoded)-len("\nUpdated\n")], &node)).To(Succeed())
	})
	It("checks artifacts by content and confines reads to the document directory", func() {
		dir := GinkgoT().TempDir()
		root, err := os.OpenRoot(dir)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(root.Close)
		data := []byte("extracted content")
		Expect(root.WriteFile("asset.md", data, 0600)).To(Succeed())
		_, err = ReadArtifact(root, Artifact{Path: "asset.md", SHA256: Hash(data)})
		Expect(err).NotTo(HaveOccurred())
		for _, path := range []string{"../asset.md", filepath.Join(GinkgoT().TempDir(), "asset.md"), "asset.md"} {
			_, err = ReadArtifact(root, Artifact{Path: path, SHA256: "incorrect"})
			Expect(err).To(HaveOccurred())
		}
	})
})
