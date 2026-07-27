package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo verification runtime spec", func() {
	It("uses the requested overall timeout", func() {
		timeout, err := validateTodoVerificationSpec(&api.Spec{
			Budget: api.Budget{Timeout: "45s"},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(timeout).To(Equal(45 * time.Second))
	})

	It("rejects commit policies because verification is a single non-mutating fixture run", func() {
		_, err := validateTodoVerificationSpec(&api.Spec{
			Workflow: &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun}}},
		})

		Expect(err).To(MatchError("verification runs cannot commit"))
	})

	It("accepts a nested spec while executing the persisted fixture once", func() {
		workDir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		created, err := uiTestProviderFor(workDir).Create(GinkgoT().Context(), todos.CreateRequest{
			Title: "Run verification with a spec",
			Body: verificationFixtureBody("Description.", `### command: prompt run smoke

`+"```bash"+`
echo prompt-run-verification-ok
`+"```"+`

- contains: prompt-run-verification-ok`),
		})
		Expect(err).NotTo(HaveOccurred())
		payload, err := json.Marshal(todoVerificationRunPayload{
			Ref:  todos.TODOReference(created),
			Spec: &api.Spec{Budget: api.Budget{Timeout: "30s"}},
		})
		Expect(err).NotTo(HaveOccurred())

		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(
			recorder,
			httptest.NewRequest(http.MethodPost, "/api/todos/verification/run", bytes.NewReader(payload)),
		)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var response struct {
			Verification types.CheckResult `json:"verification"`
		}
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Verification.AllPassed).To(BeTrue())
		Expect(response.Verification.Output).NotTo(BeNil())
		Expect(response.Verification.Output.Results).To(HaveLen(1))
	})
})
