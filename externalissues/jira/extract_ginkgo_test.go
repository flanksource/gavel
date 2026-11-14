package jira

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func documentTestImage() []byte {
	var data bytes.Buffer
	Expect(png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 2, 2)))).To(Succeed())
	return data.Bytes()
}

var _ = Describe("document extraction images", func() {
	It("rejects Word image relationships that would write outside the extraction directory", func() {
		for _, target := range []string{"media/image1.png", "media/../../escape.png", `media/..\escape.png`} {
			var data bytes.Buffer
			archive := zip.NewWriter(&data)
			writer, err := archive.Create("word/_rels/document.xml.rels")
			Expect(err).NotTo(HaveOccurred())
			_, err = fmt.Fprintf(writer, `<Relationships><Relationship Target="%s"/></Relationships>`, target)
			Expect(err).NotTo(HaveOccurred())
			Expect(archive.Close()).To(Succeed())
			path := filepath.Join(GinkgoT().TempDir(), "report.docx")
			Expect(os.WriteFile(path, data.Bytes(), 0o600)).To(Succeed())
			err = validateDOCXMedia(path)
			if target == "media/image1.png" {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError(ContainSubstring("unsafe DOCX image target")))
			}
		}
	})

	It("stores repeated HTML and Markdown images once in attachments/cas with relative Markdown references", func() {
		dir := GinkgoT().TempDir()
		image := documentTestImage()
		Expect(os.Mkdir(filepath.Join(dir, "media"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "media", "image1.png"), image, 0o600)).To(Succeed())
		input := "# Document\n\n**Content** <br> <img src=\"media/image1.png\" id=\"image1\">\n\n![Again](media/image1.png)\n\nEnd"
		output := GinkgoT().TempDir()
		result, err := localizeDocumentImages(documentImageOptions{Markdown: input, SourceDir: dir, OutputDir: output})
		Expect(err).NotTo(HaveOccurred())
		relative := fmt.Sprintf("attachments/cas/%x.png", sha256.Sum256(image))
		Expect(result.Markdown).To(SatisfyAll(ContainSubstring("# Document"), ContainSubstring("**Content**"), ContainSubstring("![image1]("+relative+")"), ContainSubstring("![Again]("+relative+")"), Not(ContainSubstring("data:image"))))
		Expect(result.Images).To(Equal([]string{relative}))
		data, err := os.ReadFile(filepath.Join(output, filepath.FromSlash(relative)))
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(image))
	})

	It("rejects missing, non-image and escaping image paths including symlinks", func() {
		dir := GinkgoT().TempDir()
		outside := filepath.Join(GinkgoT().TempDir(), "outside.png")
		Expect(os.WriteFile(outside, documentTestImage(), 0o600)).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(dir, "escape.png"))).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "text.png"), []byte("not an image"), 0o600)).To(Succeed())
		for _, path := range []string{"../outside.png", outside, "https://example.test/image.png", "missing.png", "escape.png", "text.png"} {
			_, err := localizeDocumentImages(documentImageOptions{Markdown: `<img src="` + path + `">`, SourceDir: dir, OutputDir: GinkgoT().TempDir()})
			Expect(err).To(HaveOccurred(), path)
		}
	})

	It("recognizes PDF and DOCX extensions or MIME types while ignoring other attachments", func() {
		for _, attachment := range []externalissues.Attachment{{Name: "REPORT.PDF"}, {MIMEType: "application/pdf; charset=binary"}} {
			Expect(documentKind(attachment)).To(Equal("pdf"))
		}
		Expect(documentKind(externalissues.Attachment{Name: "Report.DOCX"})).To(Equal("docx"))
		Expect(documentKind(externalissues.Attachment{MIMEType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document"})).To(Equal("docx"))
		Expect(documentKind(externalissues.Attachment{Name: "notes.txt"})).To(BeEmpty())
	})
})
