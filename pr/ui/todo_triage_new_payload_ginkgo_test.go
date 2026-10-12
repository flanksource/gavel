package ui

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Triage new payload", func() {
	DescribeTable("decodes triage with body precedence over query", func(contentType, body, query string, want bool) {
		request := httptest.NewRequest(http.MethodPost, "/api/todos/new"+query, strings.NewReader(body))
		request.Header.Set("Content-Type", contentType)
		payload, attachments, err := parseTodoNewPayload(request)
		Expect(err).NotTo(HaveOccurred())
		Expect(attachments).To(BeEmpty())
		Expect(payload.Triage).NotTo(BeNil())
		Expect(*payload.Triage).To(Equal(want))
	},
		Entry("JSON true", "application/json", `{"triage":true}`, "", true),
		Entry("JSON false overrides query with autoSave", "application/json", `{"triage":false,"autoSave":true}`, "?triage=true", false),
		Entry("form true", "application/x-www-form-urlencoded", "triage=true&autoSave=false", "", true),
		Entry("form false overrides query", "application/x-www-form-urlencoded", "triage=false", "?triage=true", false),
		Entry("query true", "", "", "?triage=true", true),
		Entry("query false", "", "", "?triage=false", false),
	)
	It("decodes an explicit unchecked multipart value", func() {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		Expect(writer.WriteField("autoSave", "true")).To(Succeed())
		Expect(writer.WriteField("triage", "false")).To(Succeed())
		Expect(writer.Close()).To(Succeed())
		request := httptest.NewRequest(http.MethodPost, "/api/todos/new?triage=true", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		payload, _, err := parseTodoNewPayload(request)
		Expect(err).NotTo(HaveOccurred())
		Expect(payload.Triage).NotTo(BeNil())
		Expect(*payload.Triage).To(BeFalse())
	})
	DescribeTable("rejects invalid values", func(contentType, body, query string) {
		request := httptest.NewRequest(http.MethodPost, "/api/todos/new"+query, strings.NewReader(body))
		request.Header.Set("Content-Type", contentType)
		_, _, err := parseTodoNewPayload(request)
		Expect(err).To(HaveOccurred())
	},
		Entry("JSON string", "application/json", `{"triage":"yes"}`, ""),
		Entry("form invalid", "application/x-www-form-urlencoded", "triage=yes", ""),
		Entry("empty query", "", "", "?triage="),
	)
})
