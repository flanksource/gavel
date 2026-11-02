//go:build document_live

package jira

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func metafileDOCX() []byte {
	// An EMF header, rectangle and EOF, using a 100 by 100 device/frame.
	image := make([]byte, 152)
	for offset, value := range map[int]uint32{0: 1, 4: 108, 16: 100, 20: 100, 32: 2540, 36: 2540, 40: 0x464d4520, 44: 0x10000, 48: 152, 52: 3, 56: 1, 72: 100, 76: 100, 80: 25, 84: 25, 108: 43, 112: 24, 124: 90, 128: 90, 132: 14, 136: 20, 148: 20} {
		binary.LittleEndian.PutUint32(image[offset:], value)
	}
	var result bytes.Buffer
	archive := zip.NewWriter(&result)
	source, err := zip.NewReader(bytes.NewReader(extractionDOCX()), int64(len(extractionDOCX())))
	Expect(err).NotTo(HaveOccurred())
	for _, file := range source.File {
		reader, err := file.Open()
		Expect(err).NotTo(HaveOccurred())
		data, err := io.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		Expect(reader.Close()).To(Succeed())
		name := strings.ReplaceAll(file.Name, "image1.png", "image1.emf")
		if file.Name == "word/media/image1.png" {
			data = image
		} else {
			data = bytes.ReplaceAll(data, []byte("image1.png"), []byte("image1.emf"))
		}
		writer, err := archive.Create(name)
		Expect(err).NotTo(HaveOccurred())
		_, err = writer.Write(data)
		Expect(err).NotTo(HaveOccurred())
	}
	Expect(archive.Close()).To(Succeed())
	return result.Bytes()
}

var _ = Describe("real metafile document extraction", func() {
	It("reports invalid metafiles with source context", func() {
		err := convertDocumentMetafile(context.Background(), metafileOptions{Data: []byte("invalid metafile"), Source: "media/broken.wmf", Output: filepath.Join(GinkgoT().TempDir(), "broken.svg")})
		Expect(err).To(MatchError(ContainSubstring("media/broken.wmf")))
	})

	It("extracts DOCX EMF media as SVG with CAS references", func() {
		path, output := filepath.Join(GinkgoT().TempDir(), "metafile.docx"), GinkgoT().TempDir()
		Expect(os.WriteFile(path, metafileDOCX(), 0o600)).To(Succeed())
		result, err := extractDocument(context.Background(), documentExtractOptions{Attachment: externalissues.Attachment{Name: "metafile.docx"}, Path: path, OutputDir: output})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Images).To(HaveLen(1))
		Expect(filepath.Ext(result.Images[0])).To(Equal(".svg"))
		Expect(result.Markdown).To(ContainSubstring("](" + result.Images[0] + ")"))
		data, err := os.ReadFile(filepath.Join(output, result.Images[0]))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("<svg"))
	})
})
