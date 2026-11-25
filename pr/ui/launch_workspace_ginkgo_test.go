package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/launch"
	"github.com/flanksource/gavel/resultstore"
	"github.com/flanksource/gavel/workspacefiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type workspaceRouteSchemas struct{}

func (workspaceRouteSchemas) Schema(string) (map[string]any, error) {
	return map[string]any{"type": "object", "properties": map[string]any{}}, nil
}

var _ = Describe("launch workspace routing", func() {
	DescribeTable("serves workspace files through the dashboard router", func(development bool) {
		root := GinkgoT().TempDir()
		output, err := exec.Command("git", "init", root).CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		Expect(os.Mkdir(filepath.Join(root, "docs"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "docs", "guide.md"), []byte("# Guide\n"), 0o600)).To(Succeed())
		workspaces, err := launch.NewWorkspaces(launch.WorkspacePolicy{Root: root, AllowLocal: true, LocalRoots: []string{root}})
		Expect(err).NotTo(HaveOccurred())
		store, err := resultstore.OpenSnapshots(filepath.Join(root, ".gavel", "runs"))
		Expect(err).NotTo(HaveOccurred())
		service, err := launch.NewService(launch.ServiceOptions{Workspaces: workspaces, Store: store, Schemas: workspaceRouteSchemas{}, Executor: launch.NativeExecutor{}})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(service.Close()).To(Succeed()) })
		server := &Server{}
		server.SetLaunchService(service)
		if development {
			server.SetDevProxy(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(viteUpstreamMarker)) }))
		}
		handler := server.Handler()
		body, err := json.Marshal(workspacefiles.EntriesRequest{FileRequest: workspacefiles.FileRequest{Root: root, Path: "docs"}})
		Expect(err).NotTo(HaveOccurred())
		request := httptest.NewRequest(http.MethodPost, "/api/workspace/entries", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		listed := httptest.NewRecorder()
		handler.ServeHTTP(listed, request)
		Expect(listed.Code).To(Equal(http.StatusOK), listed.Body.String())
		Expect(listed.Header().Get("Content-Type")).To(Equal("application/json"))
		Expect(listed.Body.String()).To(MatchJSON(`[{"path":"docs/guide.md","name":"guide.md","directory":false,"visible":true}]`))

		read := httptest.NewRecorder()
		handler.ServeHTTP(read, httptest.NewRequest(http.MethodGet, "/api/workspace/file?"+url.Values{"root": {root}, "path": {"docs/guide.md"}}.Encode(), nil))
		Expect(read.Code).To(Equal(http.StatusOK), read.Body.String())
		var file workspacefiles.File
		Expect(json.Unmarshal(read.Body.Bytes(), &file)).To(Succeed())
		Expect(file.Source).To(Equal("# Guide\n"))

		outside := httptest.NewRecorder()
		handler.ServeHTTP(outside, httptest.NewRequest(http.MethodGet, "/api/workspace/file?"+url.Values{"root": {filepath.Dir(root)}, "path": {"guide.md"}}.Encode(), nil))
		Expect(outside.Code).To(Equal(http.StatusUnprocessableEntity), outside.Body.String())
		Expect(outside.Body.String()).To(ContainSubstring("outside enabled roots"))
	}, Entry("embedded mode", false), Entry("development mode", true))
})
