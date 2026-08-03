package ui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"time"

	cexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project status lifecycle", func() {
	var originalProjectsPath string

	BeforeEach(func() {
		originalProjectsPath = projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: "/work/gavel", Repos: []string{"acme/gavel"}}})).To(Succeed())
	})

	AfterEach(func() {
		projectsPath = originalProjectsPath
	})

	It("returns the selected project's gavel status and action state", func() {
		originalGather := gatherProjectStatus
		gatherProjectStatus = func(workDir string, opts status.Options) (*status.Result, error) {
			Expect(workDir).To(Equal("/work/gavel"))
			Expect(opts.NoRepomap).To(BeTrue())
			Expect(opts.NoResults).To(BeTrue())
			return &status.Result{
				WorkDir: "/work/gavel",
				Branch:  "feature/projects",
				Files: []status.FileStatus{{
					Path:     "pr/ui/src/App.tsx",
					State:    status.StateUnstaged,
					WorkKind: status.KindModified,
					Adds:     12,
					Dels:     3,
				}},
			}, nil
		}
		DeferCleanup(func() { gatherProjectStatus = originalGather })

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/projects/gavel/status", nil)
		request.SetPathValue("name", "gavel")
		(&Server{}).handleProjectStatus(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var response projectStatusResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response).To(Equal(projectStatusResponse{
			Project: Project{Name: "gavel", Dir: "/work/gavel", Repos: []string{"acme/gavel"}},
			WorkDir: "/work/gavel",
			Branch:  "feature/projects",
			Files: []projectFileStatus{{
				Path:     "pr/ui/src/App.tsx",
				State:    status.StateUnstaged,
				WorkKind: status.KindModified,
				Adds:     12,
				Dels:     3,
			}},
		}))
		var rawResponse map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &rawResponse)).To(Succeed())
		Expect(rawResponse).NotTo(HaveKey("commitQueue"))
	})

	It("includes snapshot-derived project results only when requested", func() {
		originalGather := gatherProjectStatus
		gatherProjectStatus = func(workDir string, opts status.Options) (*status.Result, error) {
			Expect(workDir).To(Equal("/work/gavel"))
			Expect(opts.NoResults).To(BeFalse())
			return &status.Result{
				WorkDir:      workDir,
				ResultsStale: true,
				Files: []status.FileStatus{{
					Path:         "pr/ui/src/App.tsx",
					TestStatus:   status.TestStatus{Passed: 4},
					ResultsStale: true,
				}},
			}, nil
		}
		DeferCleanup(func() { gatherProjectStatus = originalGather })

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/projects/gavel/status?includeResults=true", nil)
		request.SetPathValue("name", "gavel")
		(&Server{}).handleProjectStatus(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var response projectStatusResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.ResultsStale).To(BeTrue())
		Expect(response.Files).To(HaveLen(1))
		Expect(response.Files[0].TestStatus.Passed).To(Equal(4))
		Expect(response.Files[0].ResultsStale).To(BeTrue())
	})

	It("rejects an invalid includeResults option", func() {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/projects/gavel/status?includeResults=sometimes", nil)
		request.SetPathValue("name", "gavel")
		(&Server{}).handleProjectStatus(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring("includeResults"))
	})

	It("serves deep links for the dedicated projects tab", func() {
		request := httptest.NewRequest(http.MethodGet, "/projects/Clicky%20UI", nil)

		route, ok := parseRouteRequest(request)

		Expect(ok).To(BeTrue())
		Expect(route.Tab).To(Equal(viewTabProjects))
	})

	It("publishes project status and lifecycle actions in OpenAPI", func() {
		paths, ok := projectsOpenAPI()["paths"].(map[string]any)
		Expect(ok).To(BeTrue())

		statusPath, ok := paths["/api/projects/{name}/status"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(statusPath).To(HaveKey("get"))
		statusJSON, err := json.Marshal(statusPath["get"])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(statusJSON)).To(ContainSubstring(`"name":"includeResults"`))

		actionsPath, ok := paths["/api/projects/{name}/actions"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(actionsPath).To(HaveKey("post"))

		ignorePath, ok := paths["/api/projects/{name}/ignore"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(ignorePath).To(HaveKey("post"))

		schemaPath, ok := paths["/api/projects/{name}/actions/schema"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(schemaPath).To(HaveKey("get"))

		diffPath, ok := paths["/api/projects/{name}/diff"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(diffPath).To(HaveKey("get"))

		runPath, ok := paths["/api/project-runs/{runId}/api/tests/stream"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(runPath).To(HaveKey("get"))

		commitPath, ok := paths["/api/projects/{name}/commit-queue"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(commitPath).To(HaveKey("post"))
		Expect(paths).NotTo(HaveKey("/api/projects/{name}/commit-queue/{id}"))
	})

	It("serves project action schemas from the injected command provider", func() {
		server := &Server{}
		server.SetProjectActionOptionsProvider(fakeProjectActionOptionsProvider{
			schema: ProjectActionSchema{
				SchemaVersion: 1,
				Action:        "lint",
				Schema:        map[string]any{"type": "object"},
				Defaults:      map[string]any{"changed": false},
			},
		})

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/projects/gavel/actions/schema?action=lint", nil)
		request.SetPathValue("name", "gavel")
		server.handleProjectActionSchema(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var response ProjectActionSchema
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response).To(Equal(ProjectActionSchema{
			SchemaVersion: 1,
			Action:        "lint",
			Schema:        map[string]any{"type": "object"},
			Defaults:      map[string]any{"changed": false},
		}))
	})

	It("runs validated advanced options without changing the quick action contract", func() {
		provider := fakeProjectActionOptionsProvider{
			args:     []string{"--changed=true", "--linters=golangci-lint"},
			received: map[string]any{},
		}
		server := &Server{}
		server.SetProjectActionOptionsProvider(provider)

		commands := make(chan []string, 1)
		originalExecute := executeProjectAction
		executeProjectAction = func(_ context.Context, _ string, args []string, output io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
			return runFakeProjectAction(group, func() (int, error) {
				commands <- append([]string(nil), args...)
				_, err := io.WriteString(output, "ok")
				return 0, err
			}, opts...)
		}
		DeferCleanup(func() { executeProjectAction = originalExecute })

		body := bytes.NewBufferString(`{"action":"lint","options":{"changed":true,"linters":["golangci-lint"]}}`)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/actions", body)
		request.SetPathValue("name", "gavel")
		server.handleProjectAction(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusAccepted), recorder.Body.String())
		Eventually(commands).Should(Receive(Equal([]string{
			"lint", "--work-dir", "/work/gavel", "--changed=true", "--linters=golangci-lint",
		})))
		Expect(provider.received).To(Equal(map[string]any{"changed": true, "linters": []any{"golangci-lint"}}))
	})

	It("sends commits to the queue instead of the one-shot action endpoint", func() {
		body := bytes.NewBufferString(`{"action":"commit","files":["one.go"]}`)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/actions", body)
		request.SetPathValue("name", "gavel")
		(&Server{}).handleProjectAction(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("/api/projects/{name}/commit-queue"))
	})

	It("maps lint and changed-test actions to their gavel commands", func() {
		originalGather := gatherProjectStatus
		gatherProjectStatus = func(string, status.Options) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "one.go"}}}, nil
		}
		DeferCleanup(func() { gatherProjectStatus = originalGather })

		commands := make(chan []string, 2)
		originalExecute := executeProjectAction
		executeProjectAction = func(_ context.Context, _ string, args []string, output io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
			return runFakeProjectAction(group, func() (int, error) {
				commands <- append([]string(nil), args...)
				_, err := io.WriteString(output, "ok")
				return 0, err
			}, opts...)
		}
		DeferCleanup(func() { executeProjectAction = originalExecute })

		server := &Server{}
		for _, action := range []projectAction{projectActionLint, projectActionTest} {
			body := bytes.NewBufferString(`{"action":"` + string(action) + `"}`)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/actions", body)
			request.SetPathValue("name", "gavel")
			server.handleProjectAction(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusAccepted), recorder.Body.String())
			Eventually(func() bool { return !server.projectActionFor("gavel").Running }).Should(BeTrue())
		}

		Expect(<-commands).To(Equal([]string{"lint", "--work-dir", "/work/gavel"}))
		Expect(<-commands).To(Equal([]string{"test", "--work-dir", "/work/gavel", "--changed"}))
	})

	It("validates project action files without loading historical results", func() {
		originalGather := gatherProjectStatus
		gatherProjectStatus = func(_ string, opts status.Options) (*status.Result, error) {
			Expect(opts.NoResults).To(BeTrue())
			return &status.Result{Files: []status.FileStatus{{Path: "one.go"}}}, nil
		}
		DeferCleanup(func() { gatherProjectStatus = originalGather })

		Expect(validateProjectActionFiles("/work/gavel", []string{"one.go"})).To(Succeed())
	})

	It("serves test and lint action details from isolated runner SSE streams", func() {
		for _, action := range []projectAction{projectActionTest, projectActionLint} {
			release := make(chan struct{})
			originalExecute := executeProjectAction
			executeProjectAction = func(_ context.Context, workDir string, _ []string, _ io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
				return runFakeProjectAction(group, func() (int, error) {
					<-release
					completed := testui.Snapshot{
						Metadata: &testui.SnapshotMetadata{Started: time.Now().UTC(), Ended: time.Now().UTC()},
						Status:   testui.SnapshotStatus{LintRun: action == projectActionLint},
					}
					if action == projectActionTest {
						completed.Tests = []parsers.Test{{Name: "focused test", Passed: true}}
					} else {
						completed.Lint = []*linters.LinterResult{{Linter: "golangci-lint", Success: true}}
					}
					_, err := snapshots.SavePerRun(workDir, &completed, completed.Metadata.Started, "")
					Expect(err).NotTo(HaveOccurred())
					return 0, nil
				}, opts...)
			}

			server := &Server{}
			started, err := server.startProjectAction(Project{Name: "gavel", Dir: GinkgoT().TempDir()}, action, []string{string(action)})
			Expect(err).NotTo(HaveOccurred())
			Expect(started.RunID).NotTo(BeEmpty())

			httpServer := httptest.NewServer(server.Handler())
			response, err := http.Get(httpServer.URL + "/api/project-runs/" + started.RunID + "/api/tests/stream")
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Header.Get("Content-Type")).To(HavePrefix("text/event-stream"))
			line, err := bufio.NewReader(response.Body).ReadString('\n')
			Expect(err).NotTo(HaveOccurred())
			Expect(line).To(HavePrefix("data: "))
			var snapshot struct {
				Status struct {
					Running bool `json:"running"`
					LintRun bool `json:"lint_run"`
				} `json:"status"`
			}
			Expect(json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data: "))), &snapshot)).To(Succeed())
			Expect(snapshot.Status.Running).To(BeTrue())
			Expect(snapshot.Status.LintRun).To(Equal(action == projectActionLint))
			Expect(response.Body.Close()).To(Succeed())

			close(release)
			Eventually(func() bool { return !server.projectActionFor("gavel").Running }).Should(BeTrue())
			Eventually(func(g Gomega) {
				response, err := http.Get(httpServer.URL + "/api/project-runs/" + started.RunID + "/api/tests")
				g.Expect(err).NotTo(HaveOccurred())
				defer response.Body.Close()
				var terminal struct {
					Status struct {
						Running bool `json:"running"`
					} `json:"status"`
					Tests []parsers.Test          `json:"tests"`
					Lint  []*linters.LinterResult `json:"lint"`
				}
				g.Expect(json.NewDecoder(response.Body).Decode(&terminal)).To(Succeed())
				g.Expect(terminal.Status.Running).To(BeFalse())
				if action == projectActionTest {
					g.Expect(terminal.Tests).To(HaveLen(1))
				} else {
					g.Expect(terminal.Lint).To(HaveLen(1))
				}
			}).Should(Succeed())
			httpServer.Close()
			executeProjectAction = originalExecute
		}
	})

	It("publishes command output while the project action is still running", func() {
		release := make(chan struct{})
		wrote := make(chan struct{})
		originalExecute := executeProjectAction
		executeProjectAction = func(_ context.Context, _ string, _ []string, output io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
			return runFakeProjectAction(group, func() (int, error) {
				_, err := io.WriteString(output, "discovering tests\n")
				close(wrote)
				<-release
				return 0, err
			}, opts...)
		}
		DeferCleanup(func() { executeProjectAction = originalExecute })

		server := &Server{}
		_, err := server.startProjectAction(Project{Name: "gavel", Dir: "/work/gavel"}, projectActionTest, []string{"test"})
		Expect(err).NotTo(HaveOccurred())
		Eventually(wrote).Should(BeClosed())
		Expect(server.projectActionFor("gavel")).To(SatisfyAll(
			WithTransform(func(action projectActionStatus) bool { return action.Running }, BeTrue()),
			WithTransform(func(action projectActionStatus) string { return action.Output }, Equal("discovering tests\n")),
		))

		close(release)
		Eventually(func() bool { return !server.projectActionFor("gavel").Running }).Should(BeTrue())
	})

	It("returns a conflict while another project action is running", func() {
		originalGather := gatherProjectStatus
		gatherProjectStatus = func(string, status.Options) (*status.Result, error) {
			return &status.Result{}, nil
		}
		DeferCleanup(func() { gatherProjectStatus = originalGather })

		release := make(chan struct{})
		originalExecute := executeProjectAction
		executeProjectAction = func(_ context.Context, _ string, _ []string, _ io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
			return runFakeProjectAction(group, func() (int, error) {
				<-release
				return 0, nil
			}, opts...)
		}
		DeferCleanup(func() { executeProjectAction = originalExecute })

		server := &Server{}
		requestAction := func() *httptest.ResponseRecorder {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/actions", bytes.NewBufferString(`{"action":"lint"}`))
			request.SetPathValue("name", "gavel")
			server.handleProjectAction(recorder, request)
			return recorder
		}

		Expect(requestAction().Code).To(Equal(http.StatusAccepted))
		Eventually(func() bool { return server.projectActionFor("gavel").Running }).Should(BeTrue())
		Expect(requestAction().Code).To(Equal(http.StatusConflict))
		close(release)
		Eventually(func() bool { return !server.projectActionFor("gavel").Running }).WithTimeout(time.Second).Should(BeTrue())
	})
})

type fakeProjectActionOptionsProvider struct {
	schema   ProjectActionSchema
	args     []string
	received map[string]any
}

func runFakeProjectAction(group *clickytask.Group, execute func() (int, error), opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
	return clickytask.StartTask("Run fake project action", func(_ commonscontext.Context, _ *clickytask.Task) (cexec.ExecResult, error) {
		exitCode, err := execute()
		return cexec.ExecResult{ExitCode: exitCode}, err
	}, append([]clickytask.Option{clickytask.WithGroup(group)}, opts...)...)
}

func (p fakeProjectActionOptionsProvider) Schema(string) (ProjectActionSchema, error) {
	return p.schema, nil
}

func (p fakeProjectActionOptionsProvider) Args(_ string, options map[string]any) ([]string, error) {
	for key, value := range options {
		p.received[key] = value
	}
	return p.args, nil
}
