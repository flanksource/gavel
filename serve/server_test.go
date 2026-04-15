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
		Expect(string(content)).To(ContainSubstring("/usr/local/bin/gavel test --lint"))
		Expect(string(content)).To(ContainSubstring(`--cwd "$WORKDIR"`))
	})
})

var _ = Describe("renderHookScript", func() {
	It("injects pre-hooks before the main command", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel", HookSpec{
			Pre: []HookStep{{Name: "deps", Run: "make deps"}},
		})
		preIdx := strings.Index(script, "make deps")
		mainIdx := strings.Index(script, "/bin/gavel test --lint")
		Expect(preIdx).NotTo(Equal(-1))
		Expect(mainIdx).NotTo(Equal(-1))
		Expect(preIdx).To(BeNumerically("<", mainIdx))
		Expect(script).To(ContainSubstring(" pre: deps"))
	})

	It("replaces the main command when ssh.cmd is set to a non-gavel command", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel", HookSpec{
			Cmd: "make ci",
		})
		Expect(script).To(ContainSubstring(`(cd "$WORKDIR" && make ci)`))
		Expect(script).NotTo(ContainSubstring("/bin/gavel test --lint"))
	})

	It("rewrites a gavel-prefixed ssh.cmd to the real binary path", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel", HookSpec{
			Cmd: "gavel test --ui",
		})
		Expect(script).To(ContainSubstring(`/bin/gavel test --ui --cwd "$WORKDIR" --no-progress`))
	})

	It("runs post-hooks even when the main command fails", func() {
		script := renderHookScript("/repos/bare", "/bin/gavel", HookSpec{
			Post: []HookStep{{Name: "notify", Run: "echo done"}},
		})
		// post-hooks must come after MAIN_EXIT is captured and wrap in
		// `set +e` so a failing notifier doesn't mask the real failure.
		mainExit := strings.Index(script, "MAIN_EXIT=$?")
		postRun := strings.Index(script, "echo done")
		failExit := strings.Index(script, "exit $MAIN_EXIT")
		Expect(mainExit).NotTo(Equal(-1))
		Expect(postRun).NotTo(Equal(-1))
		Expect(failExit).NotTo(Equal(-1))
		Expect(mainExit).To(BeNumerically("<", postRun))
		Expect(postRun).To(BeNumerically("<", failExit))
		// post-hooks isolated from `set -e`.
		postSetPlus := strings.Index(script[postRun-120:postRun], "set +e")
		Expect(postSetPlus).NotTo(Equal(-1))
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
