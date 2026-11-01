package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBranchValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sandbox branch validation")
}

var _ = Describe("Sandbox branch validation", func() {
	DescribeTable("rejects invalid input before invoking Git", func(branch string) {
		bin := GinkgoT().TempDir()
		probe := filepath.Join(bin, "invoked")
		Expect(os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\nprintf called >\"$GAVEL_GIT_PROBE\"\n"), 0700)).To(Succeed())
		GinkgoT().Setenv("PATH", bin)
		GinkgoT().Setenv("GAVEL_GIT_PROBE", probe)
		_, err := (&Sources{}).BranchCommit(context.Background(), Repo{}, branch)
		Expect(err).To(MatchError(ContainSubstring("not a valid branch name")))
		_, err = os.Stat(probe)
		Expect(os.IsNotExist(err)).To(BeTrue())
	},
		Entry("option", "--upload-pack=outside"),
		Entry("empty", ""),
		Entry("newline", "main\n"),
		Entry("NUL", "main\x00"),
	)
})
