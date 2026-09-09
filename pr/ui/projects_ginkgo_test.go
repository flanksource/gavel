package ui

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project catalog resolution", func() {
	var originalPath string

	BeforeEach(func() {
		originalPath = projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
	})

	AfterEach(func() {
		projectsPath = originalPath
	})

	It("names the catalog file when a project cannot be found", func() {
		_, err := GetProject("config-db")

		Expect(err).To(MatchError(And(
			ContainSubstring(`unknown project: "config-db"`),
			ContainSubstring(projectsPath),
		)))
		Expect(errors.Is(err, ErrProjectNotFound)).To(BeTrue())
	})

	It("returns malformed catalog errors with the catalog path", func() {
		Expect(os.WriteFile(projectsPath, []byte("{"), 0o600)).To(Succeed())

		_, err := LoadProjects()

		Expect(err).To(MatchError(ContainSubstring(projectsPath)))
	})

	It("selects the most-specific configured project for a nested directory", func() {
		root := GinkgoT().TempDir()
		parent := filepath.Join(root, "workspace")
		child := filepath.Join(parent, "services", "config-db")
		Expect(os.MkdirAll(filepath.Join(child, "internal"), 0o755)).To(Succeed())
		Expect(SaveProjects([]Project{
			{Name: "workspace", Dir: parent},
			{Name: "config-db", Dir: child, Repos: []string{"flanksource/config-db"}},
		})).To(Succeed())

		project, err := ProjectForDir(filepath.Join(child, "internal"))

		Expect(err).NotTo(HaveOccurred())
		Expect(project.Name).To(Equal("config-db"))
	})

	It("names the catalog file when no configured project contains a directory", func() {
		Expect(SaveProjects([]Project{{Name: "other", Dir: GinkgoT().TempDir()}})).To(Succeed())

		_, err := ProjectForDir(GinkgoT().TempDir())

		Expect(err).To(MatchError(And(
			ContainSubstring("project for directory"),
			ContainSubstring(projectsPath),
		)))
		Expect(errors.Is(err, ErrProjectNotFound)).To(BeTrue())
	})
})

// A project registered without repos must stay an empty list all the way to the
// browser: Go marshals a nil slice as JSON null, and the dashboard's Project
// type declares repos as a plain array, so a null crashes the projects sidebar
// on `project.repos[0]`. These specs assert the raw bytes on both sides of the
// catalog, because unmarshalling erases the null/[] distinction.
var _ = Describe("project repos are never null", func() {
	var originalPath string

	BeforeEach(func() {
		originalPath = projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		originalCounts := projectTodoCounts
		projectTodoCounts = func(context.Context, Project) (todoCounts, error) { return todoCounts{}, nil }
		DeferCleanup(func() {
			projectsPath = originalPath
			projectTodoCounts = originalCounts
		})
	})

	It("persists an empty list rather than null for a project with no repos", func() {
		Expect(SaveProjects([]Project{{Name: "time-kiosk", Dir: GinkgoT().TempDir()}})).To(Succeed())

		stored, err := os.ReadFile(projectsPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(stored)).To(ContainSubstring(`"repos": []`))
		Expect(string(stored)).NotTo(ContainSubstring("null"))
	})

	It("heals a catalog written by an older gavel that stored repos as null", func() {
		Expect(os.WriteFile(projectsPath, []byte(`[{"name":"time-kiosk","dir":"/tmp/time-kiosk","repos":null}]`), 0o600)).To(Succeed())

		projects, err := LoadProjects()

		Expect(err).NotTo(HaveOccurred())
		Expect(projects).To(HaveLen(1))
		Expect(projects[0].Repos).NotTo(BeNil())
		Expect(projects[0].Repos).To(BeEmpty())
	})

	It("serves an empty repos list from the projects collection", func() {
		Expect(os.WriteFile(projectsPath, []byte(`[{"name":"time-kiosk","dir":"/tmp/time-kiosk","repos":null}]`), 0o600)).To(Succeed())

		recorder := httptest.NewRecorder()
		(&Server{}).handleProjects(recorder, httptest.NewRequest(http.MethodGet, "/api/projects", nil))

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.String()).To(ContainSubstring(`"repos":[]`))
	})

	It("echoes an empty repos list when a project is created without one", func() {
		recorder := httptest.NewRecorder()
		body := strings.NewReader(`{"name":"time-kiosk","dir":"/tmp/time-kiosk"}`)
		(&Server{}).handleProjects(recorder, httptest.NewRequest(http.MethodPost, "/api/projects", body))

		Expect(recorder.Code).To(Equal(http.StatusCreated))
		Expect(recorder.Body.String()).To(ContainSubstring(`"repos":[]`))
	})
})
