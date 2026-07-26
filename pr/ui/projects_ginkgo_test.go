package ui

import (
	"errors"
	"os"
	"path/filepath"

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
