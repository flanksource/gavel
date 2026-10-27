package launch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/resultstore"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type heldExecutor struct{ starts atomic.Int64 }
type heldExecution struct{ owner *heldExecutor }

func (e *heldExecutor) Prepare(context.Context, PreparedWorkspace, Configuration) (Execution, error) {
	return heldExecution{e}, nil
}
func (e heldExecution) Run(ctx context.Context, _ Publish) error {
	e.owner.starts.Add(1)
	<-ctx.Done()
	return ctx.Err()
}

var _ = Describe("launch admission", func() {
	var service *Service
	var executor *heldExecutor
	var root string
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		workspaces, err := NewWorkspaces(WorkspacePolicy{Root: root, AllowLocal: true, LocalRoots: []string{root}})
		Expect(err).NotTo(HaveOccurred())
		store, err := resultstore.OpenSnapshots(filepath.Join(root, ".gavel"))
		Expect(err).NotTo(HaveOccurred())
		executor = &heldExecutor{}
		service, err = NewService(ServiceOptions{Workspaces: workspaces, Store: store, Schemas: schemaFixture{}, Executor: executor})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(service.Close()).To(Succeed()) })
	})
	It("persists before dispatch, reuses root and child retries, and stops the exact run", func() {
		request := Request{Workspace: &shell.Setup{Cwd: root}, Config: Configuration{Lint: &LintConfig{}}}
		body, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())
		post := func(path string, data []byte) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
			req.Header.Set("Idempotency-Key", "same-operation")
			response := httptest.NewRecorder()
			service.Handler().ServeHTTP(response, req)
			return response
		}
		first := post("/api/launch", body)
		Expect(first.Code).To(Equal(http.StatusAccepted), first.Body.String())
		var accepted struct {
			ExecutionID string `json:"executionId"`
		}
		Expect(json.Unmarshal(first.Body.Bytes(), &accepted)).To(Succeed())
		stored, err := service.options.Store.Get(context.Background(), accepted.ExecutionID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.Metadata.Revision).To(BeNumerically(">=", 1))
		second := post("/api/launch/lint", body)
		Expect(second.Code).To(Equal(http.StatusAccepted), second.Body.String())
		Expect(second.Body.String()).To(ContainSubstring(accepted.ExecutionID))
		Eventually(executor.starts.Load).Should(Equal(int64(1)))
		request.Labels = map[string]string{"changed": "true"}
		changed, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())
		Expect(post("/api/launch", changed).Code).To(Equal(http.StatusConflict))
		Expect(service.Stop(context.Background(), accepted.ExecutionID)).To(Succeed())
		Eventually(func() string {
			run, err := service.Get(context.Background(), accepted.ExecutionID)
			Expect(err).NotTo(HaveOccurred())
			return run.Status.State
		}, 5*time.Second).Should(Equal("cancelled"))
	})
	It("serves exact schema documents and rejects forbidden child sections", func() {
		for _, kind := range []string{"", "/test", "/lint"} {
			response := httptest.NewRecorder()
			service.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/launch"+kind+"?schema", nil))
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal("application/schema+json"))
		}
		response := httptest.NewRecorder()
		service.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/launch/lint", bytes.NewBufferString(`{"workspace":{"cwd":"."},"config":{"test":{},"lint":{}}}`)))
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity), response.Body.String())
		Expect(response.Body.String()).To(ContainSubstring("instancePath"))
		document := map[string]any{"paths": map[string]any{}}
		Expect(service.AddOpenAPI(document)).To(Succeed())
		paths := document["paths"].(map[string]any)
		Expect(paths).To(HaveKey("/api/launch/prepare"))
		Expect(paths).To(HaveKey("/api/launch/prompts"))
		Expect(paths).To(HaveKey("/api/runs/{executionId}/stream"))
	})
	It("accepts a stop from another dispatcher without replaying the execution", func() {
		request := Request{Workspace: &shell.Setup{Cwd: root}, Config: Configuration{Lint: &LintConfig{}}}
		run, err := service.Start(context.Background(), request, "cross-dispatcher", "reviewer")
		Expect(err).NotTo(HaveOccurred())
		workspaces, err := NewWorkspaces(WorkspacePolicy{Root: root, AllowLocal: true, LocalRoots: []string{root}})
		Expect(err).NotTo(HaveOccurred())
		store, err := resultstore.OpenSnapshots(filepath.Join(root, ".gavel"))
		Expect(err).NotTo(HaveOccurred())
		other, err := NewService(ServiceOptions{Workspaces: workspaces, Store: store, Schemas: schemaFixture{}, Executor: executor})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(other.Close()).To(Succeed()) })
		Expect(other.Stop(context.Background(), run.Metadata.ExecutionID)).To(Succeed())
		Eventually(func() string {
			result, err := service.Get(context.Background(), run.Metadata.ExecutionID)
			Expect(err).NotTo(HaveOccurred())
			return result.Status.State
		}, 5*time.Second).Should(Equal("cancelled"))
		Expect(executor.starts.Load()).To(BeNumerically("<=", 1))
	})
	It("marks an abandoned owner interrupted and never replays it", func() {
		run := resultstore.Run{Metadata: resultstore.Metadata{ExecutionID: uuid.NewString(), Workspace: root, Started: time.Now()}, Status: resultstore.Status{Running: true, State: "running"}, Admission: "abandoned", InputHash: "immutable", Dispatcher: json.RawMessage(`{}`)}
		stored, _, err := service.options.Store.Create(context.Background(), run)
		Expect(err).NotTo(HaveOccurred())
		recovered, err := service.Get(context.Background(), stored.Metadata.ExecutionID)
		Expect(err).NotTo(HaveOccurred())
		Expect(recovered.Status.State).To(Equal("interrupted"))
		Expect(recovered.Terminal()).To(BeTrue())
		Expect(executor.starts.Load()).To(BeZero())
	})
	It("reports a disallowed workspace as an input error", func() {
		body, err := json.Marshal(Request{Workspace: &shell.Setup{Cwd: filepath.Dir(root)}, Config: Configuration{Test: map[string]any{}}})
		Expect(err).NotTo(HaveOccurred())
		req := httptest.NewRequest(http.MethodPost, "/api/launch/test", bytes.NewReader(body))
		req.Header.Set("Idempotency-Key", "invalid-workspace")
		response := httptest.NewRecorder()
		service.Handler().ServeHTTP(response, req)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity), response.Body.String())
	})

	It("keeps prepared handles and result reads scoped to the caller", func() {
		prepared, err := service.options.Workspaces.Prepare(context.Background(), shell.Setup{Cwd: root}, "reviewer")
		Expect(err).NotTo(HaveOccurred())
		_, err = service.options.Workspaces.Get(prepared.ID, "another-reviewer", true)
		Expect(err).To(HaveOccurred())
		run, err := service.Start(context.Background(), Request{PreparationID: prepared.ID, Config: Configuration{Test: map[string]any{}}}, "owned", "reviewer")
		Expect(err).NotTo(HaveOccurred())
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/runs/"+run.Metadata.ExecutionID, nil)
		req.Header.Set("Authorization", "another-reviewer")
		service.Handler().ServeHTTP(response, req)
		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(service.Stop(context.Background(), run.Metadata.ExecutionID)).To(Succeed())
		Eventually(func() bool {
			stored, err := service.Get(context.Background(), run.Metadata.ExecutionID)
			Expect(err).NotTo(HaveOccurred())
			return stored.Terminal()
		}).Should(BeTrue())
		replayed, err := service.Start(context.Background(), Request{PreparationID: prepared.ID, Config: Configuration{Test: map[string]any{}}}, "owned", "reviewer")
		Expect(err).NotTo(HaveOccurred())
		Expect(replayed.Metadata.ExecutionID).To(Equal(run.Metadata.ExecutionID))
	})
})
