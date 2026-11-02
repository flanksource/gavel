//go:build document_acceptance

package jira

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("supplied DOCX acceptance", func() {
	It("publishes eleven resolving SVG links and preserves output on failed extraction", func() {
		document, err := os.ReadFile(os.Getenv("GAVEL_DOCX_ACCEPTANCE"))
		Expect(err).NotTo(HaveOccurred())
		original := bytes.Clone(document)
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1":
				fmt.Fprint(w, `{"key":"EXAMPLE-1","fields":{"summary":"SVG acceptance","project":{"id":"10001"},"attachment":[{"id":"20001","filename":"report.docx","content":"https://example.test/report.docx"}]}}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case "/rest/api/3/field", "/rest/api/3/issue/EXAMPLE-1/remotelink":
				fmt.Fprint(w, `[]`)
			case "/rest/api/3/issue/EXAMPLE-1/comment":
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			case "/rest/api/3/attachment/content/20001":
				_, err := w.Write(document)
				Expect(err).NotTo(HaveOccurred())
			default:
				http.NotFound(w, r)
			}
		})
		dir := os.Getenv("GAVEL_DOCX_OUTPUT")
		Expect(dir).NotTo(BeEmpty())
		options := SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: dir, Extract: true, Store: CatalogStore{Dir: dir}}
		results, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(results[0].Images).To(HaveLen(11))
		before, err := os.ReadFile(results[0].Path)
		Expect(err).NotTo(HaveOccurred())
		for _, image := range results[0].Images {
			Expect(filepath.Ext(image)).To(Equal(".svg"))
			data, err := os.ReadFile(image)
			Expect(err).NotTo(HaveOccurred())
			Expect(validateDocumentSVG(data)).To(Succeed())
			relative, err := filepath.Rel(filepath.Dir(results[0].Path), image)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(before)).To(ContainSubstring("](" + filepath.ToSlash(relative) + ")"))
		}
		repeated, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(repeated[0].Images).To(Equal(results[0].Images))
		document = []byte("invalid document")
		_, err = client.Sync(context.Background(), options)
		Expect(err).To(HaveOccurred())
		after, err := os.ReadFile(results[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
		attachment, err := os.ReadFile(results[0].Attachments[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(attachment).To(Equal(original))
	})
})
