package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"
)

var _ = Describe("Triage new creation", func() {
	var called int
	var startError error
	BeforeEach(func() {
		called = 0
		startError = nil
		resolve, start := run.Resolve, run.Start
		DeferCleanup(func() { run.Resolve = resolve; run.Start = start })
		run.Resolve = func(ctx context.Context, req run.Request) (*run.Prepared, error) {
			Expect(req.Options.Step).To(Equal("triage.new"))
			persisted, err := req.Provider.Get(ctx, req.Todo.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(persisted.MarkdownBody).To(ContainSubstring("Persist this body"))
			return &run.Prepared{Resolution: &lifecycle.Resolution{}}, nil
		}
		run.Start = func(req run.Request) (run.StartResult, error) {
			called++
			Expect(req.Approvals).To(BeTrue())
			Expect(req.Prepared).NotTo(BeNil())
			return run.StartResult{Status: "started", SessionID: "approved-session"}, startError
		}
	})
	DescribeTable("starts only explicitly requested triage after saving", func(option string, want int) {
		server := &Server{ghOpts: github.Options{WorkDir: GinkgoT().TempDir()}}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/todos/new", strings.NewReader(`{"title":"New task","body":"Persist this body"`+option+`}`))
		request.Header.Set("Content-Type", "application/json")
		server.handleTodoNew(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusCreated), recorder.Body.String())
		Expect(called).To(Equal(want))
	}, Entry("checked", `,"triage":true`, 1), Entry("unchecked", `,"triage":false`, 0), Entry("omitted", "", 0))
	It("returns the saved TODO and explicit failure when admission fails", func() {
		startError = fmt.Errorf("provider unavailable")
		server := &Server{ghOpts: github.Options{WorkDir: GinkgoT().TempDir()}}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/todos/new", strings.NewReader(`{"title":"New task","body":"Persist this body","triage":true}`))
		request.Header.Set("Content-Type", "application/json")
		server.handleTodoNew(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusCreated))
		var response map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response["todo"]).NotTo(BeNil())
		Expect(response["triage"]).To(HaveKeyWithValue("error", "provider unavailable"))
	})
})
