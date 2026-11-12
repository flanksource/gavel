package editor

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/flanksource/gavel/projects"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fixtureProject is the project useProject registers.
const fixtureProject = "demo"

// useProject registers dir as the only project in a spec-local catalog.
func useProject(dir string) {
	path := filepath.Join(GinkgoT().TempDir(), "projects.json")
	DeferCleanup(projects.UseFileForTest(path))
	_, err := projects.NewFile(path).Create(context.Background(), projects.Project{Name: fixtureProject, Dir: dir})
	Expect(err).NotTo(HaveOccurred())
}

func jsonString(value string) string {
	data, err := json.Marshal(value)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

// mustTarget is where doc's threads live under the registered project.
func mustTarget(doc Document) fileTarget {
	target, err := targetOf(context.Background(), doc.Path)
	Expect(err).NotTo(HaveOccurred())
	return target
}
