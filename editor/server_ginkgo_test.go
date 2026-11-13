package editor

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing/fstest"
)

var _ = Describe("Editor HTTP API", func() {
	var handler http.Handler
	var doc Document
	var store *Store
	BeforeEach(func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.json")
		Expect(os.WriteFile(path, []byte(`{"name":"initial"}`), 0600)).To(Succeed())
		var err error
		store, err = NewStore([]string{path})
		Expect(err).NotTo(HaveOccurred())
		doc, err = store.Read(store.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		handler = (&Server{Store: store, Context: context.Background(), Assets: fstest.MapFS{"index.html": {Data: []byte("Editor")}}}).Handler()
	})
	It("GET /api/me answers the actor an asserted comment is authored as", func() {
		author := comments.Actor{Kind: "agent", ID: "claude:session-1", Label: "Acme Reviewer"}
		request := httptest.NewRequest("GET", "/api/me", nil).WithContext(commentapi.WithActor(context.Background(), author))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var me comments.Actor
		Expect(json.Unmarshal(response.Body.Bytes(), &me)).To(Succeed())
		Expect(me).To(Equal(author))
	})
	Describe("POST /api/lint/ignore", func() {
		var server *Server
		var workDir string
		post := func(body string) *httptest.ResponseRecorder {
			request := httptest.NewRequest("POST", "/api/lint/ignore", bytes.NewBufferString(body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, request)
			return response
		}
		BeforeEach(func() {
			workDir = GinkgoT().TempDir()
			server = &Server{Store: store, Context: context.Background()}
			server.run = LintRun{Results: []*linters.LinterResult{{
				Linter:     "json",
				WorkDir:    workDir,
				Violations: []models.Violation{{File: "config.json", Line: 3, Source: "json"}},
			}}}
		})
		It("saves the ignore to the run's .gavel.yaml and drops the violation from the run", func() {
			response := post(`{"source":"json","file":"config.json","line":3,"work_dir":"` + workDir + `"}`)
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			config, err := os.ReadFile(filepath.Join(workDir, ".gavel.yaml"))
			Expect(err).NotTo(HaveOccurred())

			status := httptest.NewRecorder()
			server.Handler().ServeHTTP(status, httptest.NewRequest("GET", "/api/lint", nil))
			var run LintRun
			Expect(json.Unmarshal(status.Body.Bytes(), &run)).To(Succeed())
			Expect(string(config)).To(SatisfyAll(ContainSubstring("source: json"), ContainSubstring("file: config.json"), ContainSubstring("line: 3")))
			Expect(run.Results[0].Violations).To(BeEmpty())
		})
		It("refuses a work_dir that is not part of the current lint run", func() {
			response := post(`{"source":"json","work_dir":"` + GinkgoT().TempDir() + `"}`)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("not part of the current lint run"))
		})
	})
	It("lists running lint tasks on the task-manager API the navbar indicator reads", func() {
		clicky.ClearGlobalTasks()
		DeferCleanup(clicky.ClearGlobalTasks)
		release := make(chan struct{})
		DeferCleanup(func() { close(release) })
		group := clicky.StartGroup[string](testui.LintTaskGroupName, clickytask.WithConcurrency(1))
		group.Add("json", func(_ commonsContext.Context, task *clickytask.Task) (string, error) {
			<-release
			task.Success()
			return "ok", nil
		})

		Eventually(func() []clickytask.RunMeta {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/tasks", nil))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			var runs []clickytask.RunMeta
			Expect(json.Unmarshal(response.Body.Bytes(), &runs)).To(Succeed())
			return runs
		}).Should(ContainElement(SatisfyAll(
			HaveField("Name", testui.LintTaskGroupName),
			HaveField("Status", "running"),
		)))
	})
	It("serves the embedded page and scoped files without running lint", func() {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
		Expect(response.Code).To(Equal(200))
		Expect(response.Body.String()).To(Equal("Editor"))
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/api/lint", nil))
		Expect(response.Body.String()).To(ContainSubstring(`"running":false`))
	})
	It("round-trips the file write contract and rejects the stale revision", func() {
		data, err := json.Marshal(map[string]string{"source": "{", "revision": doc.Revision})
		Expect(err).NotTo(HaveOccurred())
		for _, status := range []int{200, 409} {
			request := httptest.NewRequest("PUT", "/api/file/"+doc.ID, bytes.NewReader(data))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		}
	})
	It("serves the page from the dev proxy while the API and unknown API paths stay on this server", func() {
		proxied := []string{}
		proxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxied = append(proxied, r.URL.Path)
			_, _ = w.Write([]byte("vite"))
		})
		dev := (&Server{Store: store, Context: context.Background(), DevProxy: proxy}).Handler()
		serve := func(path string) *httptest.ResponseRecorder {
			response := httptest.NewRecorder()
			dev.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
			return response
		}

		page, module, lint, unknown := serve("/"), serve("/src/index.tsx"), serve("/api/lint"), serve("/api/nope")

		Expect(lint.Body.String()).To(ContainSubstring(`"running":false`))
		Expect(map[string]any{
			"page":    page.Body.String(),
			"module":  module.Body.String(),
			"unknown": []any{unknown.Code, unknown.Header().Get("Content-Type"), unknown.Body.String()},
			"proxied": proxied,
		}).To(Equal(map[string]any{
			"page":    "vite",
			"module":  "vite",
			"unknown": []any{404, "application/json", "{\"error\":\"no API route for GET /api/nope\"}\n"},
			"proxied": []string{"/", "/src/index.tsx"},
		}))
	})
	Describe("GET /api/file/{id}/image", func() {
		const png = "\x89PNG\r\n\x1a\n\x00\x00"
		var images http.Handler
		var ids map[string]string
		BeforeEach(func() {
			parent := GinkgoT().TempDir()
			root := filepath.Join(parent, "docs")
			Expect(os.MkdirAll(filepath.Join(root, "img"), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(parent, "x.png"), []byte(png), 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, "img", "a.png"), []byte(png), 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, "notes.txt"), []byte("notes"), 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, "guide.md"), []byte("![](img/a.png)"), 0o600)).To(Succeed())
			s, err := NewStore([]string{root})
			Expect(err).NotTo(HaveOccurred())
			ids = map[string]string{}
			for _, f := range s.Files() {
				ids[filepath.Base(f.Path)] = f.ID
			}
			images = (&Server{Store: s, Context: context.Background()}).Handler()
		})
		serve := func(target string) []any {
			response := httptest.NewRecorder()
			images.ServeHTTP(response, httptest.NewRequest("GET", target, nil))
			if response.Code != 200 {
				return []any{response.Code}
			}
			return []any{response.Code, response.Header().Get("Content-Type"), response.Body.String()}
		}
		It("serves an image entry's own bytes", func() {
			Expect(serve("/api/file/" + ids["a.png"] + "/image")).To(Equal([]any{200, "image/png", png}))
		})
		It("serves an image relative to a markdown document", func() {
			Expect(serve("/api/file/" + ids["guide.md"] + "/image?path=img/a.png")).To(Equal([]any{200, "image/png", png}))
		})
		DescribeTable("rejects paths it must not serve",
			func(id, rel string, want int) {
				Expect(serve("/api/file/" + ids[id] + "/image?path=" + rel)).To(Equal([]any{want}))
			},
			Entry("escape above the root", "guide.md", "../x.png", 400),
			Entry("non-image file", "guide.md", "notes.txt", 400),
			Entry("missing image", "guide.md", "img/missing.png", 404),
			Entry("document without a path", "guide.md", "", 400),
		)
	})
	It("rejects cross-origin writes and unknown files", func() {
		request := httptest.NewRequest("PUT", "/api/file/"+doc.ID, bytes.NewBufferString(`{}`))
		request.Header.Set("Origin", "https://example.com")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(403))
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/api/file/unknown", nil))
		Expect(response.Code).To(Equal(404))
	})
})
