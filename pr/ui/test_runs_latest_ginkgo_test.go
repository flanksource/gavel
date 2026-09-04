package ui

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("project tree test runs", func() {
	var currentPath string
	var workDir string

	BeforeEach(func() {
		originalProjectsPath := projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		DeferCleanup(func() { projectsPath = originalProjectsPath })

		workDir = GinkgoT().TempDir()
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: workDir}})).To(Succeed())

		gavelDir := filepath.Join(workDir, snapshots.Dir)
		Expect(os.MkdirAll(gavelDir, 0o755)).To(Succeed())
		currentPath = filepath.Join(gavelDir, "sha-current.json")
		current := testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Started: time.Date(2026, 7, 22, 8, 0, 0, 0, time.UTC)},
			Tests:    []parsers.Test{{Name: "current", Passed: true}},
		}
		data, err := json.Marshal(current)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(currentPath, data, 0o644)).To(Succeed())

		pointer, err := json.Marshal(snapshots.Pointer{
			Path: filepath.Join(snapshots.Dir, filepath.Base(currentPath)),
			SHA:  "current",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(gavelDir, snapshots.PointerLast+".json"), pointer, 0o644)).To(Succeed())

		newer := testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Started: time.Date(2026, 7, 23, 8, 0, 0, 0, time.UTC)},
			Tests:    []parsers.Test{{Name: "newer-history", Failed: true}},
		}
		_, err = snapshots.SavePerRun(workDir, &newer, newer.Metadata.Started, "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("shows only the snapshot referenced by .gavel/last.json", func() {
		projects, err := collectTestRuns(context.Background())

		Expect(err).NotTo(HaveOccurred())
		Expect(projects).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("gavel"),
			"Dir":  Equal(workDir),
			"Runs": ConsistOf(MatchFields(IgnoreExtras, Fields{
				"RunID":  Equal(snapshots.PointerLast),
				"Passed": Equal(1),
				"Failed": Equal(0),
				"Total":  Equal(1),
			})),
		})))
	})

	It("resolves the last run route through the pointer", func() {
		path, err := resolveRunPath(workDir, snapshots.PointerLast)

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(currentPath))
	})
})
