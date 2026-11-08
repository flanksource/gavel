package editor

import (
	"context"
	"encoding/json"
	"path/filepath"

	ui "github.com/flanksource/gavel/pr/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fixtureProject is the project useProject registers.
const fixtureProject = "demo"

// useProject registers dir as the only project in a spec-local catalog.
func useProject(dir string) {
	DeferCleanup(ui.UseProjectsFileForTest(filepath.Join(GinkgoT().TempDir(), "projects.json")))
	Expect(ui.SaveProjects([]ui.Project{{Name: fixtureProject, Dir: dir}})).To(Succeed())
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
