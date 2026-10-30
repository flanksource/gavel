//go:build document_live

package jira

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func extractionDOCX() []byte {
	var data bytes.Buffer
	archive := zip.NewWriter(&data)
	for name, content := range map[string][]byte{
		"word/document.xml":            []byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><w:body><w:p><w:r><w:t>Extracted DOCX body</w:t></w:r></w:p><w:p><w:r><w:drawing><a:blip r:embed="rId1"/></w:drawing></w:r></w:p></w:body></w:document>`),
		"word/_rels/document.xml.rels": []byte(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image1.png"/></Relationships>`),
		"word/media/image1.png":        documentTestImage(),
	} {
		writer, err := archive.Create(name)
		Expect(err).NotTo(HaveOccurred())
		_, err = writer.Write(content)
		Expect(err).NotTo(HaveOccurred())
	}
	Expect(archive.Close()).To(Succeed())
	return data.Bytes()
}

func extractionPDF() []byte {
	content := "BT /F1 18 Tf 40 350 Td (Extracted PDF body) Tj ET\nq 100 0 0 100 40 180 cm /Image1 Do Q\n"
	image := strings.Repeat("\xff\x00\x00", 4)
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 400 400] /Resources << /Font << /F1 4 0 R >> /XObject << /Image1 6 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content), content),
		fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width 2 /Height 2 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Length %d >>\nstream\n%s\nendstream", len(image), image),
	}
	var data bytes.Buffer
	data.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects))
	for i, object := range objects {
		offsets[i] = data.Len()
		fmt.Fprintf(&data, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := data.Len()
	fmt.Fprintf(&data, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&data, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&data, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return data.Bytes()
}

var _ = Describe("real uvx document extraction", func() {
	It("extracts PDF and DOCX text with images referencing attachments/cas", func() {
		for _, document := range []struct {
			Name string
			Body []byte
			Text string
		}{{"report.docx", extractionDOCX(), "Extracted DOCX body"}, {"report.pdf", extractionPDF(), "Extracted PDF body"}} {
			path := filepath.Join(GinkgoT().TempDir(), document.Name)
			Expect(os.WriteFile(path, document.Body, 0o600)).To(Succeed())
			output := GinkgoT().TempDir()
			body, err := extractDocument(context.Background(), documentExtractOptions{Attachment: externalissues.Attachment{Name: document.Name}, Path: path, OutputDir: output})
			Expect(err).NotTo(HaveOccurred(), document.Name)
			Expect(body.Markdown).To(SatisfyAll(ContainSubstring(document.Text), ContainSubstring("](attachments/cas/"), Not(ContainSubstring("data:image"))))
			Expect(body.Images).NotTo(BeEmpty())
			for _, image := range body.Images {
				Expect(filepath.Join(output, image)).To(BeAnExistingFile())
			}
		}
	})

	It("rejects invalid DOCX files before launching the converter", func() {
		path := filepath.Join(GinkgoT().TempDir(), "invalid.docx")
		Expect(os.WriteFile(path, []byte("invalid Word document"), 0o600)).To(Succeed())
		_, err := extractDocument(context.Background(), documentExtractOptions{Attachment: externalissues.Attachment{Name: "invalid.docx"}, Path: path, OutputDir: GinkgoT().TempDir()})
		Expect(err).To(MatchError(ContainSubstring("read DOCX archive")))
	})

	It("makes extract imply attachment downloads and inlines extracted documents in the issue", func() {
		document := extractionDOCX()
		originalDocument := bytes.Clone(document)
		contentRequests := 0
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1":
				fmt.Fprint(w, `{"key":"EXAMPLE-1","fields":{"summary":"Extraction sync","project":{"id":"10001"},"attachment":[{"id":"20001","filename":"report.docx","content":"https://example.test/report.docx"}]}}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case "/rest/api/3/field", "/rest/api/3/issue/EXAMPLE-1/remotelink":
				fmt.Fprint(w, `[]`)
			case "/rest/api/3/issue/EXAMPLE-1/comment":
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			case "/rest/api/3/attachment/content/20001":
				contentRequests++
				_, err := w.Write(document)
				Expect(err).NotTo(HaveOccurred())
			default:
				http.NotFound(w, r)
			}
		})
		dir := GinkgoT().TempDir()
		options := SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: dir, Extract: true, Store: CatalogStore{Dir: dir}}
		result, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(result[0].Extracted).To(Equal([]string{"20001"}))
		Expect(result[0].Attachments).To(HaveLen(1))
		data, err := os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring("## Extracted attachments"), ContainSubstring("[report.docx](EXAMPLE-1.attachments/20001/report.docx)"), ContainSubstring("Extracted DOCX body"), ContainSubstring("](attachments/cas/"), Not(ContainSubstring("data:image"))))
		Expect(result[0].Images).To(HaveLen(1))
		Expect(result[0].Images[0]).To(BeAnExistingFile())
		_, err = client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(contentRequests).To(Equal(1))
		for _, damaged := range []string{"markdown", "image", "pipeline"} {
			doc, root := readCachedIssue(result[0].Path)
			extraction := doc.Metadata.Attachments["20001"].Extraction
			switch damaged {
			case "markdown":
				Expect(root.Remove(extraction.Markdown.Path)).To(Succeed())
			case "image":
				Expect(root.Remove(extraction.Images[0].Path)).To(Succeed())
			case "pipeline":
				extraction.PipelineFingerprint = "outdated converter"
				writeCachedIssue(result[0].Path, doc)
			}
			_, err = client.Sync(context.Background(), options)
			Expect(err).NotTo(HaveOccurred(), damaged)
			Expect(contentRequests).To(Equal(1), damaged)
		}
		data, err = os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		before := data
		document = []byte("corrupted document")
		options.Force = true
		_, err = client.Sync(context.Background(), options)
		Expect(err).To(MatchError(ContainSubstring("20001")))
		data, err = os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(before))
		original, err := os.ReadFile(result[0].Attachments[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(original).To(Equal(originalDocument))
	})
})
