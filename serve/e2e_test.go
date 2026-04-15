package serve

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gossh "golang.org/x/crypto/ssh"
)

var _ = Describe("SSH Git Serve E2E", func() {
	var (
		server       *Server
		listener     net.Listener
		clientRepo   string
		artifactsDir string
		sshCmd       string
	)

	BeforeEach(func() {
		repoDir := GinkgoT().TempDir()
		clientRepo = GinkgoT().TempDir()
		artifactsDir = GinkgoT().TempDir()

		// Generate ephemeral SSH key in OpenSSH format — the OpenSSH client
		// (used by `git push` via GIT_SSH_COMMAND) rejects PKCS#8 PEM blocks.
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		Expect(err).NotTo(HaveOccurred())
		pemBlock, err := gossh.MarshalPrivateKey(priv, "")
		Expect(err).NotTo(HaveOccurred())
		sshKeyPath := filepath.Join(GinkgoT().TempDir(), "id_ed25519")
		Expect(os.WriteFile(sshKeyPath, pem.EncodeToMemory(pemBlock), 0o600)).To(Succeed())

		// Custom hook that records worktree contents instead of running gavel
		hookWriter := func(bareRepo, _ string) error {
			hooksDir := filepath.Join(bareRepo, "hooks")
			if err := os.MkdirAll(hooksDir, 0o755); err != nil {
				return err
			}
			ad := artifactsDir
			script := fmt.Sprintf(`#!/bin/bash
set -euo pipefail
ARTIFACTS="%s"
BARE="%s"
while read oldrev newrev refname; do
  WORKDIR=$(mktemp -d)
  trap "rm -rf $WORKDIR" EXIT
  unset GIT_DIR
  git clone "$BARE" "$WORKDIR" 2>/dev/null
  git -C "$WORKDIR" checkout "$newrev" 2>/dev/null
  cd "$WORKDIR"
  find . -not -path './.git/*' -type f | sort > "$ARTIFACTS/files.txt"
  echo "LINT_RAN=true" > "$ARTIFACTS/lint_marker.txt"
  echo "FILE_COUNT=$(find . -not -path './.git/*' -type f | wc -l | tr -d ' ')" >> "$ARTIFACTS/lint_marker.txt"
  test -d "$WORKDIR/.git" && echo "HAS_GIT=true" >> "$ARTIFACTS/lint_marker.txt"
  rm -rf "$WORKDIR"
  trap - EXIT
done
`, ad, bareRepo)
			return os.WriteFile(filepath.Join(hooksDir, "post-receive"), []byte(script), 0o755)
		}

		server, err = NewServer(Options{
			Host:       "127.0.0.1",
			RepoDir:    repoDir,
			HookWriter: hookWriter,
		})
		Expect(err).NotTo(HaveOccurred())

		listener, err = net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		port := listener.Addr().(*net.TCPAddr).Port

		go func() {
			defer GinkgoRecover()
			_ = server.Serve(listener)
		}()

		// Init client repo with known files
		git := func(args ...string) {
			cmd := exec.Command("git", args...)
			cmd.Dir = clientRepo
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
		}
		git("init", "-b", "main")
		git("config", "user.email", "test@test.com")
		git("config", "user.name", "Test")
		git("config", "commit.gpgsign", "false")

		Expect(os.WriteFile(filepath.Join(clientRepo, "README.md"), []byte("# Test\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(clientRepo, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(clientRepo, "subdir"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(clientRepo, "subdir", "config.yaml"), []byte("key: value\n"), 0o644)).To(Succeed())

		git("add", ".")
		git("commit", "-m", "initial")

		sshCmd = fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p %d", sshKeyPath, port)
		git("remote", "add", "gavel", "ssh://127.0.0.1/test-project")
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	push := func() {
		cmd := exec.Command("git", "push", "gavel", "main")
		cmd.Dir = clientRepo
		cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "git push failed: %s", out)
	}

	It("receives all git tree contents into a worktree", func() {
		push()

		content, err := os.ReadFile(filepath.Join(artifactsDir, "files.txt"))
		Expect(err).NotTo(HaveOccurred())
		files := strings.TrimSpace(string(content))
		Expect(files).To(ContainSubstring("./README.md"))
		Expect(files).To(ContainSubstring("./main.go"))
		Expect(files).To(ContainSubstring("./subdir/config.yaml"))
	})

	It("runs linting on the pushed code", func() {
		push()

		content, err := os.ReadFile(filepath.Join(artifactsDir, "lint_marker.txt"))
		Expect(err).NotTo(HaveOccurred())
		marker := string(content)
		Expect(marker).To(ContainSubstring("LINT_RAN=true"))
		Expect(marker).To(ContainSubstring("FILE_COUNT=3"))
		Expect(marker).To(ContainSubstring("HAS_GIT=true"))
	})
})
