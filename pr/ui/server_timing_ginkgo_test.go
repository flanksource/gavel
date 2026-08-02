package ui

import (
	"context"
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
	var originalProjectTodoCounts func(ctx context.Context, project Project) (todoCounts, error)

	BeforeEach(func() {
		originalProjectsPath = projectsPath
		originalProjectTodoCounts = projectTodoCounts
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
	})

	AfterEach(func() {
		projectsPath = originalProjectsPath
		projectTodoCounts = originalProjectTodoCounts
	})

	It("reports total, file, and database time while browsing projects", func() {
		projectDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(projectDir, "Procfile"), []byte("web: make serve\n"), 0o600)).To(Succeed())
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: projectDir}})).To(Succeed())
		projectTodoCounts = func(_ context.Context, _ Project) (todoCounts, error) {
			return todoCounts{Total: 3, Open: 2, Completed: 1}, nil
		}

		response := requestProjectEndpoint("/api/projects")

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(serverTimingMetricNames(response.Header().Get("Server-Timing"))).To(ConsistOf("total", "file", "db"))
		Expect(response.Body.String()).To(ContainSubstring(`"todoCounts":{"total":3`))
	})

	It("reports total, file, and git time for project status", func() {
		dir := createDiffRepository()
		writeDiffFile(dir, "src/unstaged.go", "package src\n\nvar Unstaged = true\n")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())

		response := requestProjectEndpoint("/api/projects/gavel/status")

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(serverTimingMetricNames(response.Header().Get("Server-Timing"))).To(ConsistOf("total", "file", "git"))
		Expect(response.Body.String()).To(ContainSubstring(`"branch":`))
	})

	It("reports total, file, and git time for successful and invalid diffs", func() {
		dir := createDiffRepository()
		writeDiffFile(dir, "src/unstaged.go", "package src\n\nvar Unstaged = true\n")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())

		valid := requestProjectEndpoint("/api/projects/gavel/diff?path=src%2Funstaged.go")
		invalid := requestProjectEndpoint("/api/projects/gavel/diff?path=missing.go")

		Expect(valid.Code).To(Equal(http.StatusOK), valid.Body.String())
		Expect(serverTimingMetricNames(valid.Header().Get("Server-Timing"))).To(ConsistOf("total", "file", "git"))
		Expect(valid.Body.String()).To(ContainSubstring(`"path":"src/unstaged.go"`))
		Expect(invalid.Code).To(Equal(http.StatusBadRequest), invalid.Body.String())
		Expect(serverTimingMetricNames(invalid.Header().Get("Server-Timing"))).To(ConsistOf("total", "file", "git"))
		Expect(invalid.Body.String()).To(ContainSubstring(`"error":`))
	})
})

func requestProjectEndpoint(path string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
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
