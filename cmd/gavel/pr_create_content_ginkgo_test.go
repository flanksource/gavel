package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR content configuration snapshot", func() {
	It("keeps the status prompt in the same commit invocation snapshot", func() {
		cfg := verify.GavelConfig{Status: verify.StatusConfig{Summary: verify.PromptSpec{File: "summary.prompt"}}}
		got := buildCommitOptions(CommitOptions{Summary: true}, "/work/review", cfg, nil)
		Expect(got.Status).To(Equal(cfg.Status))
	})

	It("reports malformed commit configuration before entering the commit pipeline", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		output, err := exec.Command("git", "init", dir).CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("commit: [invalid"), 0o600)).To(Succeed())
		_, err = runCommit(CommitOptions{WorkDir: dir, Message: "fix: explicit"})
		Expect(err).To(MatchError(ContainSubstring("load commit configuration")))
		Expect(err.Error()).To(ContainSubstring("yaml"))
	})
})
