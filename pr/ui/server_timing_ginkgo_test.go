package ui

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project server timing", func() {
	var originalProjectsPath string
	var originalProjectsTodoCounts func(ctx context.Context, projects []Project) []todoCountsResult

	BeforeEach(func() {
		originalProjectsPath = projectsPath
		originalProjectsTodoCounts = projectsTodoCounts
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
	})

	AfterEach(func() {
		projectsPath = originalProjectsPath
		projectsTodoCounts = originalProjectsTodoCounts
	})

	It("reports total, file, and database time while browsing projects", func() {
		projectDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(projectDir, "Procfile"), []byte("web: make serve\n"), 0o600)).To(Succeed())
		projects := make([]Project, 8)
		for i := range projects {
			projects[i] = Project{Name: fmt.Sprintf("project-%d", i), Dir: projectDir}
		}
		Expect(SaveProjects(projects)).To(Succeed())
		projectsTodoCounts = perProjectTodoCounts(func(Project) (todoCounts, error) {
			return todoCounts{Total: 3, Open: 2, Completed: 1}, nil
		})

		response := requestProjectEndpoint("/api/projects")

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(serverTimingMetricNames(response.Header().Get("Server-Timing"))).To(ConsistOf("total", "file", "db"))
		Expect(response.Body.String()).To(ContainSubstring(`"todoCounts":{"total":3`))
	})

	It("reports total, file, and sql time, and no git time, for a project status read from the stored state", func() {
		server := newTrackedGitServer().server
		dir := createDiffRepository()
		writeDiffFile(dir, "src/unstaged.go", "package src\n\nvar Unstaged = true\n")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())
		Expect(requestProjectEndpointOn(server, "/api/projects/gavel/status").Code).To(Equal(http.StatusOK), "the first read tracks the repository")

		response := requestProjectEndpointOn(server, "/api/projects/gavel/status")

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(serverTimingMetricNames(response.Header().Get("Server-Timing"))).To(ConsistOf("total", "file", "sql"))
		Expect(response.Body.String()).To(ContainSubstring(`"branch":"main"`))
	})

	It("reports git time only for the live patch of a diff", func() {
		server := newTrackedGitServer().server
		dir := createDiffRepository()
		writeDiffFile(dir, "src/unstaged.go", "package src\n\nvar Unstaged = true\n")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())
		Expect(requestProjectEndpointOn(server, "/api/projects/gavel/status").Code).To(Equal(http.StatusOK), "the first read tracks the repository")

		valid := requestProjectEndpointOn(server, "/api/projects/gavel/diff?path=src%2Funstaged.go")
		invalid := requestProjectEndpointOn(server, "/api/projects/gavel/diff?path=missing.go")

		Expect(valid.Code).To(Equal(http.StatusOK), valid.Body.String())
		Expect(serverTimingMetricNames(valid.Header().Get("Server-Timing"))).To(ConsistOf("total", "file", "sql", "git"))
		Expect(valid.Body.String()).To(ContainSubstring(`"path":"src/unstaged.go"`))
		Expect(invalid.Code).To(Equal(http.StatusBadRequest), invalid.Body.String())
		Expect(serverTimingMetricNames(invalid.Header().Get("Server-Timing"))).To(ConsistOf("total", "file", "sql"))
		Expect(invalid.Body.String()).To(ContainSubstring(`"error":`))
	})
})

func requestProjectEndpoint(path string) *httptest.ResponseRecorder {
	return requestProjectEndpointOn(&Server{}, path)
}

func requestProjectEndpointOn(server *Server, path string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	return response
}

func serverTimingMetricNames(header string) []string {
	if header == "" {
		return nil
	}
	parts := strings.Split(header, ",")
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		name, _, _ := strings.Cut(strings.TrimSpace(part), ";")
		names = append(names, name)
	}
	return names
}
