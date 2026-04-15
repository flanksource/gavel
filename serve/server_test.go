package serve

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Serve Suite")
}

var _ = Describe("cleanRepoPath", func() {
	DescribeTable("strips quotes and leading slashes",
		func(input, expected string) {
			Expect(cleanRepoPath(input)).To(Equal(expected))
		},
		Entry("bare path", "myproject", "myproject"),
		Entry("quoted path", "'/myproject'", "myproject"),
		Entry("leading slash", "/myproject", "myproject"),
		Entry("multiple slashes", "///myproject", "myproject"),
		Entry("nested path", "'/org/repo'", "org/repo"),
		Entry("empty becomes default", "", "default"),
		Entry("just slash becomes default", "/", "default"),
		Entry("quoted empty becomes default", "'/'", "default"),
	)
})

var _ = Describe("writePostReceiveHook", func() {
	It("creates an executable hook script with the default command", func() {
		tmpDir := GinkgoT().TempDir()

		err := writePostReceiveHook(tmpDir, "/usr/local/bin/gavel")
		Expect(err).NotTo(HaveOccurred())

		hookPath := filepath.Join(tmpDir, "hooks", "post-receive")
		info, err := os.Stat(hookPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm() & 0o111).NotTo(BeZero())

		content, err := os.ReadFile(hookPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("#!/bin/bash"))
		Expect(string(content)).To(ContainSubstring("/usr/local/bin/gavel test --lint --ui"))
		Expect(string(content)).To(ContainSubstring(`--cwd "$WORKDIR"`))
	})
})

// The rendered script no longer templates pre/main/post. Pre/post hooks
// live in the pushed repo's .gavel.yaml and are executed by `gavel test`
// itself (gated on $CI / --skip-hooks). See cmd/gavel/test_hooks_test.go
// for that coverage.
var _ = Describe("renderHookScript", func() {
	It("runs gavel test when no ssh.cmd is set in the pushed repo", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel")
		Expect(script).To(ContainSubstring(`/bin/gavel test --lint --ui --no-progress --cwd "$WORKDIR"`))
	})

	It("extracts ssh.cmd from $WORKDIR/.gavel.yaml via yq", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel")
		Expect(script).To(ContainSubstring(`yq -r '.ssh.cmd // ""' "$WORKDIR/.gavel.yaml"`))
	})

	It("fails loud if yq is not on $PATH", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel")
		Expect(script).To(ContainSubstring(`command -v yq`))
		Expect(script).To(ContainSubstring(`yq not found`))
	})

	It("evals ssh.cmd in place of gavel test when set", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel")
		Expect(script).To(ContainSubstring(`if [ -n "$SSH_CMD" ]; then`))
		Expect(script).To(ContainSubstring(`(cd "$WORKDIR" && eval "$SSH_CMD")`))
	})

	It("exports CI=1 so gavel test picks up hooks automatically", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel")
		// CI=1 must come before EITHER branch runs so both gavel test and
		// ssh.cmd-override paths see it.
		ciIdx := strings.Index(script, "export CI=1")
		branchIdx := strings.Index(script, `if [ -n "$SSH_CMD" ]; then`)
		Expect(ciIdx).NotTo(Equal(-1))
		Expect(branchIdx).NotTo(Equal(-1))
		Expect(ciIdx).To(BeNumerically("<", branchIdx))
	})

	It("propagates the main command's non-zero exit via FAILED", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel")
		Expect(script).To(ContainSubstring(`if [ $EXIT -ne 0 ]; then`))
		Expect(script).To(ContainSubstring(`exit $EXIT`))
	})
})

var _ = Describe("ensureBareRepo", func() {
	It("initializes a bare git repo", func() {
		tmpDir, err := os.MkdirTemp("", "gavel-bare-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		repoPath := filepath.Join(tmpDir, "test.git")
		err = ensureBareRepo(repoPath)
		Expect(err).NotTo(HaveOccurred())

		// HEAD file should exist in a bare repo
		_, err = os.Stat(filepath.Join(repoPath, "HEAD"))
		Expect(err).NotTo(HaveOccurred())
	})

	It("is idempotent", func() {
		tmpDir, err := os.MkdirTemp("", "gavel-bare-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		repoPath := filepath.Join(tmpDir, "test.git")
		Expect(ensureBareRepo(repoPath)).To(Succeed())
		Expect(ensureBareRepo(repoPath)).To(Succeed())
	})
})

var _ = Describe("loadOrGenerateHostKey", func() {
	It("generates and persists a host key", func() {
		tmpDir, err := os.MkdirTemp("", "gavel-key-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		keyPath := filepath.Join(tmpDir, "host_key")
		signer, err := loadOrGenerateHostKey(keyPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(signer).NotTo(BeNil())

		// Key file should exist
		_, err = os.Stat(keyPath)
		Expect(err).NotTo(HaveOccurred())

		// Loading again should return the same key type
		signer2, err := loadOrGenerateHostKey(keyPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(signer2.PublicKey().Type()).To(Equal(signer.PublicKey().Type()))
	})
})
