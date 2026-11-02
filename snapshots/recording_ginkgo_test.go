package snapshots

import (
	"os"
	"path/filepath"

	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Snapshot recording durability", func() {
	It("publishes a new file instead of truncating readers of the old snapshot", func() {
		path := filepath.Join(GinkgoT().TempDir(), "snapshot.json")
		Expect(writeJSON(path, map[string]string{"state": "old"})).To(Succeed())
		Expect(os.Chmod(path, 0o600)).To(Succeed())
		reader, err := os.Open(path)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(reader.Close)
		old, err := reader.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(writeJSON(path, map[string]string{"state": "new"})).To(Succeed())
		current, err := os.Stat(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.SameFile(old, current)).To(BeFalse())
		Expect(current.Mode().Perm()).To(Equal(os.FileMode(0o600)))
	})

	It("returns periodic write errors at Stop even if a later retry succeeds", func() {
		workDir := GinkgoT().TempDir()
		blocked := filepath.Join(workDir, Dir)
		Expect(os.WriteFile(blocked, []byte("block directory"), 0o644)).To(Succeed())
		recorder := NewRecorder(workDir)
		recorder.Start()
		recorder.Update(&testui.Snapshot{})
		recorder.flush()
		Expect(os.Remove(blocked)).To(Succeed())
		Expect(recorder.Stop(false)).To(HaveOccurred())
		Expect(recorder.Stop(false)).To(HaveOccurred())
	})

	It("preserves an existing snapshot when encoding fails", func() {
		path := filepath.Join(GinkgoT().TempDir(), "snapshot.json")
		Expect(os.WriteFile(path, []byte(`{"state":"old"}`), 0o644)).To(Succeed())
		Expect(writeJSON(path, make(chan int))).To(HaveOccurred())
		payload, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(payload)).To(Equal(`{"state":"old"}`))
	})
})
