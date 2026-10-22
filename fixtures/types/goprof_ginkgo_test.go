package types

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture goProfiles", func() {
	It("finds no profile in a directory the command never wrote, and refuses an escaping destination", func() {
		root := GinkgoT().TempDir()
		output := fixtures.GoProfileOutput{Directory: "samples", Env: "PROFILE_DIRECTORY"}
		prepared, err := prepareGoProfiles(root, map[string]fixtures.GoProfileOutput{"samples": output})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Remove(prepared.Env["PROFILE_DIRECTORY"])).To(Succeed())
		Expect(goProfileFiles(prepared.Root, "samples", output)).To(BeEmpty())

		_, err = prepareGoProfiles(root, map[string]fixtures.GoProfileOutput{"cpu": {File: "../outside.pprof", Env: "PROFILE_FILE"}})
		Expect(err).To(MatchError(ContainSubstring("must be relative")))
	})

	It("refuses a profile written through a symlinked parent", func() {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		output := fixtures.GoProfileOutput{File: "output/cpu.pprof", Env: "PROFILE_FILE"}
		prepared, err := prepareGoProfiles(root, map[string]fixtures.GoProfileOutput{"cpu": output})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Remove(filepath.Join(prepared.Root, "output"))).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(prepared.Root, "output"))).To(Succeed())
		Expect(os.WriteFile(filepath.Join(outside, "cpu.pprof"), []byte("profile"), 0o600)).To(Succeed())
		_, err = goProfileFiles(prepared.Root, "cpu", output)
		Expect(err).To(MatchError(ContainSubstring("outside the fixture profile directory")))
	})
})
