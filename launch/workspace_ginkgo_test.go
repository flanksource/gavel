package launch

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/commons-db/shell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("workspace authorization", func() {
	It("disables local access by default and rejects a symlink outside enabled roots", func() {
		root := GinkgoT().TempDir()
		disabled, err := NewWorkspaces(WorkspacePolicy{Root: root})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(disabled.Close)
		_, err = disabled.Prepare(context.Background(), shell.Setup{Cwd: root}, "caller")
		Expect(err).To(MatchError(ContainSubstring("disabled")))
		outside := GinkgoT().TempDir()
		Expect(os.Symlink(outside, filepath.Join(root, "escape"))).To(Succeed())
		enabled, err := NewWorkspaces(WorkspacePolicy{Root: root, AllowLocal: true, LocalRoots: []string{root}})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(enabled.Close)
		_, err = enabled.Prepare(context.Background(), shell.Setup{Cwd: filepath.Join(root, "escape")}, "caller")
		Expect(err).To(MatchError(ContainSubstring("outside enabled roots")))
	})
	It("expires idle preparations while leaving the original directory in place", func() {
		root := GinkgoT().TempDir()
		workspaces, err := NewWorkspaces(WorkspacePolicy{Root: root, AllowLocal: true, LocalRoots: []string{root}})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(workspaces.Close)
		prepared, err := workspaces.Prepare(context.Background(), shell.Setup{Cwd: root}, "caller")
		Expect(err).NotTo(HaveOccurred())
		workspaces.mu.Lock()
		workspaces.entries[prepared.ID].Expires = time.Now().Add(-time.Second)
		workspaces.mu.Unlock()
		Expect(workspaces.expire()).To(Succeed())
		_, err = workspaces.Get(prepared.ID, "caller", false)
		Expect(err).To(HaveOccurred())
		Expect(root).To(BeADirectory())
	})
})
