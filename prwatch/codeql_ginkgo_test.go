package prwatch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CodeQL status findings", func() {
	It("refreshes findings on each poll and skips analysis jobs and other apps", func() {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if calls == 1 {
				fmt.Fprint(w, `{"errors":[{"message":"Forbidden"}]}`)
				return
			}
			if calls == 2 {
				fmt.Fprint(w, `{"data":{"node":{"isRequired":false}}}`)
				return
			}
			fmt.Fprint(w, `{"data":{"node":{"annotations":{"nodes":[{"path":"service.go","message":"Finding","annotationLevel":"FAILURE","location":{"start":{"line":1},"end":{"line":1}}}],"pageInfo":{"hasNextPage":false}}}}}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		checks := model.StatusChecks{
			{NodeID: "check-node", Name: "CodeQL", AppSlug: "github-advanced-security"},
			{Name: "Analyze (go)", WorkflowName: "CodeQL", AppSlug: "github-actions"},
			{Name: "CodeQL", AppSlug: "another-app"},
		}
		pr := &model.PRInfo{NodeID: "pr-node", StatusCheckRollup: checks}
		fetchCodeQLFindings(github.Options{Token: "test-token"}, pr)
		Expect(checks[0].LintError).To(ContainSubstring("Forbidden"))
		fetchCodeQLFindings(github.Options{Token: "test-token"}, pr)
		Expect(checks[0].LintError).To(BeEmpty())
		Expect(checks[0].Lint).To(HaveLen(1))
		Expect(checks[0].Lint[0].Severity).To(Equal(models.SeverityWarning))
		Expect(calls).To(Equal(3))
	})
	It("enumerates every finding including repeated files in text and JSON", func() {
		check := model.StatusCheck{Name: "CodeQL", AppSlug: "github-advanced-security", Status: "COMPLETED", Conclusion: "FAILURE"}
		for i := range 16 {
			check.Lint = append(check.Lint, models.Violation{File: "service.go", Line: i + 1, Message: models.StringPtr(fmt.Sprintf("Finding %02d", i)), Severity: models.SeverityError})
		}
		result := &PRWatchResult{PR: &model.PRInfo{StatusCheckRollup: model.StatusChecks{check}}}
		rendered := result.Pretty().String()
		for i := range 16 {
			Expect(rendered).To(ContainSubstring(fmt.Sprintf("Finding %02d", i)))
		}
		body, err := json.Marshal(result)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Count(string(body), `"severity":"error"`)).To(Equal(16))
		Expect(statusExitCode(result)).To(Equal(1))
	})

	It("scopes retrieval failures and findings with the originating check", func() {
		result := &PRWatchResult{PR: &model.PRInfo{StatusCheckRollup: model.StatusChecks{
			{Name: "CodeQL", Status: "COMPLETED", Conclusion: "SUCCESS", LintError: "Forbidden"},
			{Name: "Test", Status: "COMPLETED", Conclusion: "SUCCESS"},
		}}}
		Expect(result.Pretty().String()).To(ContainSubstring("Forbidden"))
		Expect(statusExitCode(result)).To(Equal(1))
		newResultFilters(nil, []string{"Test"}).apply(result)
		Expect(result.PR.StatusCheckRollup).To(HaveLen(1))
		Expect(result.Pretty().String()).NotTo(ContainSubstring("Forbidden"))
		Expect(statusExitCode(result)).To(Equal(0))
	})
})
