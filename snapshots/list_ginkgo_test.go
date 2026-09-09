package snapshots

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestSnapshotsGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Snapshots")
}

var _ = Describe("LastRun", func() {
	It("summarizes only the snapshot referenced by .gavel/last.json", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(workDir, Dir), 0o755)).To(Succeed())
		currentPath := filepath.Join(workDir, Dir, "sha-current.json")
		current := testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Started: time.Date(2026, 7, 22, 8, 0, 0, 0, time.UTC)},
			Tests:    []parsers.Test{{Name: "current", Passed: true}},
		}
		Expect(writeJSON(currentPath, current)).To(Succeed())
		Expect(writePointer(workDir, PointerLast, &Pointer{
			Path: filepath.Join(Dir, filepath.Base(currentPath)),
			SHA:  "current",
		})).To(Succeed())

		newer := testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Started: time.Date(2026, 7, 23, 8, 0, 0, 0, time.UTC)},
			Tests:    []parsers.Test{{Name: "newer-history", Failed: true}},
		}
		_, err := SavePerRun(workDir, &newer, newer.Metadata.Started, "")
		Expect(err).NotTo(HaveOccurred())

		run, err := LastRun(workDir)

		Expect(err).NotTo(HaveOccurred())
		Expect(run).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"RunID":  Equal(PointerLast),
			"Path":   Equal(currentPath),
			"Passed": Equal(1),
			"Failed": Equal(0),
			"Total":  Equal(1),
		})))
	})

	It("refuses a pointer whose path escapes the work dir", func() {
		// Save always records a workDir-relative `.gavel/<file>.json`. A pointer
		// naming anything else is a corrupt or hand-edited last.json, and
		// following it would stat and decode a file outside the project.
		workDir := GinkgoT().TempDir()
		const escaping = "../../../etc/passwd"
		Expect(os.Mkdir(filepath.Join(workDir, Dir), 0o755)).To(Succeed())
		Expect(writePointer(workDir, PointerLast, &Pointer{Path: escaping, SHA: "current"})).To(Succeed())

		run, err := LastRun(workDir)

		Expect(err).To(MatchError(ContainSubstring(escaping)))
		Expect(err).To(MatchError(ContainSubstring(workDir)))
		Expect(run).To(BeNil())
	})
})
