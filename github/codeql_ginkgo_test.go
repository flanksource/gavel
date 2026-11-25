package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CodeQL findings", func() {
	DescribeTable("maps annotation levels", func(level string, severity models.ViolationSeverity) {
		a := checkAnnotation{Path: "service.go", Message: "Finding", AnnotationLevel: level}
		a.Location.Start.Line, a.Location.End.Line = 1, 1
		finding, err := a.violation()
		Expect(err).NotTo(HaveOccurred())
		Expect(finding.Severity).To(Equal(severity))
	},
		Entry("failure", "FAILURE", models.SeverityError),
		Entry("warning", "WARNING", models.SeverityWarning),
		Entry("notice", "NOTICE", models.SeverityInfo),
	)
	It("fetches every page and preserves severity, messages and ranges", func() {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			var request struct {
				Variables map[string]any `json:"variables"`
			}
			Expect(json.NewDecoder(r.Body).Decode(&request)).To(Succeed())
			Expect(request.Variables["id"]).To(Equal("check-node"))
			Expect(r.Header.Get("Authorization")).To(Equal("Bearer test-token"))
			count, level, next := 100, "FAILURE", true
			if calls == 2 {
				Expect(request.Variables["after"]).To(Equal("cursor-1"))
				count, level, next = 1, "WARNING", false
			}
			nodes := make([]map[string]any, count)
			for i := range nodes {
				nodes[i] = map[string]any{"path": "src/service.go", "title": "Unsafe path", "message": "Untrusted input\nreaches the sink", "annotationLevel": level, "location": map[string]any{"start": map[string]int{"line": i + 1, "column": 2}, "end": map[string]int{"line": i + 2, "column": 8}}}
			}
			Expect(json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"node": map[string]any{"annotations": map[string]any{"nodes": nodes, "pageInfo": map[string]any{"hasNextPage": next, "endCursor": "cursor-1"}}}}})).To(Succeed())
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		findings, err := FetchCodeQLFindings(Options{Token: "test-token"}, "check-node")
		Expect(err).NotTo(HaveOccurred())
		Expect(findings).To(HaveLen(101))
		Expect(findings[0]).To(Equal(models.Violation{File: "src/service.go", Line: 1, Column: 2, EndLine: 2, EndColumn: 8, Message: models.StringPtr("Unsafe path: Untrusted input\nreaches the sink"), Source: "codeql", Severity: models.SeverityError}))
		Expect(findings[100].Severity).To(Equal(models.SeverityWarning))
		Expect(calls).To(Equal(2))
	})

	DescribeTable("rejects incomplete evidence", func(response string) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, response) }))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		_, err := FetchCodeQLFindings(Options{Token: "test-token"}, "check-node")
		Expect(err).To(HaveOccurred())
	},
		Entry("GraphQL errors", `{"errors":[{"message":"Forbidden"}]}`),
		Entry("missing check", `{"data":{"node":null}}`),
		Entry("non advancing cursor", `{"data":{"node":{"annotations":{"nodes":[],"pageInfo":{"hasNextPage":true,"endCursor":""}}}}}`),
	)

	It("accepts an empty completed result", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":{"node":{"annotations":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		findings, err := FetchCodeQLFindings(Options{Token: "test-token"}, "check-node")
		Expect(err).NotTo(HaveOccurred())
		Expect(findings).To(BeEmpty())
	})
})
