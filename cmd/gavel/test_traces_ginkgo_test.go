// Tests that a gavel test run's trace store is copied, once the run closes it, to the storage
// ~/.gavel.yaml configures, under the workspace's prefix, and stays only on disk without one.
package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/workspace"
)

var _ = Describe("a gavel test run's trace store", func() {
	It("is copied to the configured storage when the run closes it", func() {
		home, bucket, workDir := GinkgoT().TempDir(), GinkgoT().TempDir(), GinkgoT().TempDir()
		invocation := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("storage:\n  path: "+bucket+"\n"), 0o644)).To(Succeed())

		trace, err := startRunTrace(testrunner.RunOptions{WorkDir: invocation, Dirs: workspace.Dirs{Invocation: invocation, StateRoot: workDir}, Fixtures: true}, time.Now())
		Expect(err).ToNot(HaveOccurred())
		Expect(trace.close(workDir)).To(Succeed())

		prefix, err := blob.WorkspacePrefix(workDir)
		Expect(err).ToNot(HaveOccurred())
		objects, err := blob.Local(bucket).List(context.Background(), prefix+"traces/"+trace.runID()+"/")
		Expect(err).ToNot(HaveOccurred())
		Expect(objects).ToNot(BeEmpty(), "the store's files are under traces/<runID>/ in the workspace's prefix")
		Expect(trace.host.Dir()).To(BeADirectory(), "the local store stays for the run's snapshot")
	})

	It("stays only on disk without configured storage", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		workDir := GinkgoT().TempDir()

		trace, err := startRunTrace(testrunner.RunOptions{WorkDir: workDir, Dirs: workspace.Dirs{Invocation: workDir, StateRoot: workDir}, Fixtures: true}, time.Now())
		Expect(err).ToNot(HaveOccurred())
		Expect(trace.archive).To(BeNil())
		Expect(trace.close(workDir)).To(Succeed())
		Expect(trace.host.Dir()).To(BeADirectory())
	})
})
