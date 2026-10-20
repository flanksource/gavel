package git_test

import (
	"context"
	"os"
	"path/filepath"

	gitpkg "github.com/flanksource/gavel/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("clone cleanup guard", func() {
	It("refuses to delete a directory that was not registered as a clone", func() {
		dir := filepath.Join(GinkgoT().TempDir(), "unmanaged")
		Expect(os.Mkdir(dir, 0o755)).To(Succeed())
		Expect(gitpkg.NewCloneManager().RemoveClone(context.Background(), dir)).To(MatchError(ContainSubstring("not managed")))
		Expect(dir).To(BeADirectory())
	})
})
