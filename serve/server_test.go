package serve

import (
	"os"
	"path/filepath"
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
	It("creates an executable hook script", func() {
		tmpDir, err := os.MkdirTemp("", "gavel-test-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		err = writePostReceiveHook(tmpDir, "/usr/local/bin/gavel")
		Expect(err).NotTo(HaveOccurred())

		hookPath := filepath.Join(tmpDir, "hooks", "post-receive")
		info, err := os.Stat(hookPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm() & 0o111).NotTo(BeZero()) // executable

		content, err := os.ReadFile(hookPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("#!/bin/bash"))
		Expect(string(content)).To(ContainSubstring("gavel test --lint"))
		Expect(string(content)).To(ContainSubstring("/usr/local/bin/gavel"))
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
