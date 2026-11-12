package ui

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/flanksource/gavel/projects"
	. "github.com/onsi/ginkgo/v2"
)

// projectsPath is the projects.json the running test pointed the process
// catalog at.
var projectsPath string

// useProjectsFile points the process catalog at the projects.json at path for
// the running spec.
func useProjectsFile(path string) {
	projectsPath = path
	DeferCleanup(projects.UseFileForTest(path))
}

// useProjectsFileT is useProjectsFile for a plain go test.
func useProjectsFileT(t *testing.T, path string) {
	projectsPath = path
	t.Cleanup(projects.UseFileForTest(path))
}

// seedProjects replaces the test catalog with ps.
func seedProjects(ps []projects.Project) error {
	if err := os.Remove(projectsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	catalog := projects.NewFile(projectsPath)
	for _, p := range ps {
		if _, err := catalog.Create(context.Background(), p); err != nil {
			return err
		}
	}
	return nil
}
