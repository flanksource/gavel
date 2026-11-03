package jira

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SVG extraction", func() {
	It("propagates cancellation before creating converter output", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		path := filepath.Join(GinkgoT().TempDir(), "cancelled.svg")
		Expect(convertDocumentMetafile(ctx, metafileOptions{Output: path})).To(MatchError(context.Canceled))
		Expect(path).NotTo(BeAnExistingFile())
	})

	It("reports the original media path when npm is unavailable", func() {
		GinkgoT().Setenv("PATH", GinkgoT().TempDir())
		err := convertDocumentMetafile(context.Background(), metafileOptions{Source: "media/image1.emf", Output: filepath.Join(GinkgoT().TempDir(), "missing.svg")})
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring("media/image1.emf"), ContainSubstring("npm"))))
	})

	It("stores SVG bytes once and preserves repeated Markdown references", func() {
		data := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><text>Example</text></svg>`)
		source, output := GinkgoT().TempDir(), GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(source, "drawing.svg"), data, 0o600)).To(Succeed())
		options := documentImageOptions{Markdown: "![One](drawing.svg)\n\n![Two](drawing.svg)", SourceDir: source, OutputDir: output}
		for range 2 {
			result, err := localizeDocumentImages(options)
			Expect(err).NotTo(HaveOccurred())
			path := fmt.Sprintf("attachments/cas/%x.svg", sha256.Sum256(data))
			Expect(result.Images).To(Equal([]string{path}))
			Expect(result.Markdown).To(ContainSubstring("![Two](" + path + ")"))
			stored, err := os.ReadFile(filepath.Join(output, path))
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).To(Equal(data))
		}
	})

	It("rejects malformed, incorrectly namespaced and active SVG", func() {
		for _, input := range []string{`<svg xmlns="http://www.w3.org/2000/svg">`, `<svg/>`, `<html/>`, `<svg xmlns="http://www.w3.org/2000/svg"><script/></svg>`, `<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"/>`, `<svg xmlns="http://www.w3.org/2000/svg"><image href="https://example.test/image.png"/></svg>`} {
			_, err := storeDocumentImage(mustSVGRoot(), []byte(input))
			Expect(err).To(HaveOccurred(), input)
		}
	})
})

func mustSVGRoot() *os.Root {
	root, err := os.OpenRoot(GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(root.Close)
	return root
}
