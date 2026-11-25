package fileread_test

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/internal/fileread"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFileread(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "fileread") }

var _ = Describe("Read", func() {
	pngBytes := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "a.png"), pngBytes, 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "notes.txt"), []byte("plain text"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "huge.png"), bytes.Repeat([]byte{0}, fileread.MaxImageBytes+1), 0o644)).To(Succeed())
	})

	It("returns binary image bytes in image mode", func() {
		Expect(fileread.Read(root, "a.png", fileread.Options{Image: true})).To(Equal(pngBytes))
	})

	It("rejects binary content in text mode", func() {
		_, err := fileread.Read(root, "a.png", fileread.Options{})
		Expect(err).To(MatchError(fileread.ErrBinary))
	})

	It("rejects a non-image extension in image mode", func() {
		_, err := fileread.Read(root, "notes.txt", fileread.Options{Image: true})
		Expect(err).To(MatchError(fileread.ErrNotImage))
	})

	It("names the 10 MiB cap when an image is too large", func() {
		_, err := fileread.Read(root, "huge.png", fileread.Options{Image: true})
		Expect(err).To(MatchError(fileread.ErrTooLarge))
		Expect(err.Error()).To(ContainSubstring("10 MiB"))
	})
})

var _ = Describe("WriteImage", func() {
	It("serves an svg with its MIME type, nosniff and a script-blocking CSP", func() {
		w := httptest.NewRecorder()
		Expect(fileread.WriteImage(w, "logo.SVG", []byte("<svg/>"))).To(Succeed())
		Expect(map[string]string{
			"type":    w.Header().Get("Content-Type"),
			"nosniff": w.Header().Get("X-Content-Type-Options"),
			"csp":     w.Header().Get("Content-Security-Policy"),
			"body":    w.Body.String(),
		}).To(Equal(map[string]string{
			"type":    "image/svg+xml",
			"nosniff": "nosniff",
			"csp":     "default-src 'none'; style-src 'unsafe-inline'; sandbox",
			"body":    "<svg/>",
		}))
	})

	It("refuses a non-image path", func() {
		Expect(fileread.WriteImage(httptest.NewRecorder(), "notes.txt", nil)).To(MatchError(fileread.ErrNotImage))
	})
})
