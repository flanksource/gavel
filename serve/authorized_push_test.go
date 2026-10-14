// Tests the push server's authorized mode with real git pushes over SSH: keys and repos are
// checked by the Authorizer, hooks never run, and OnPushed receives the updated refs.
package serve

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gossh "golang.org/x/crypto/ssh"
)

// recordingAuthorizer stands in for the database-backed Authorizer: these specs cover
// the server's handling of its answers, not how it decides.
type recordingAuthorizer struct {
	user    uuid.UUID
	key     gossh.PublicKey
	allowed string
	target  Target
	asked   []string
}

func (a *recordingAuthorizer) UserForKey(_ context.Context, key ssh.PublicKey) (uuid.UUID, bool) {
	if a.key == nil || !ssh.KeysEqual(key, a.key) {
		return uuid.Nil, false
	}
	return a.user, true
}

func (a *recordingAuthorizer) AuthorizePush(_ context.Context, user uuid.UUID, repoPath string) (Target, error) {
	a.asked = append(a.asked, repoPath)
	if user != a.user || repoPath != a.allowed {
		return Target{}, errors.New("not this one")
	}
	return a.target, nil
}

var _ = Describe("authorized push", func() {
	var (
		authorizer *recordingAuthorizer
		clientRepo string
		bareDir    string
		sshCmd     string
		pushes     []Push
		pushesMu   sync.Mutex
	)

	BeforeEach(func() {
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		Expect(err).NotTo(HaveOccurred())
		block, err := gossh.MarshalPrivateKey(priv, "")
		Expect(err).NotTo(HaveOccurred())
		keyPath := filepath.Join(GinkgoT().TempDir(), "id_ed25519")
		Expect(os.WriteFile(keyPath, pem.EncodeToMemory(block), 0o600)).To(Succeed())
		signer, err := gossh.NewSignerFromKey(priv)
		Expect(err).NotTo(HaveOccurred())

		bareDir = filepath.Join(GinkgoT().TempDir(), "repos", "org", "repo.git")
		authorizer = &recordingAuthorizer{
			user: uuid.New(), key: signer.PublicKey(), allowed: "acme/app",
			target: Target{OrgID: uuid.New(), RepoID: uuid.New(), BareDir: bareDir},
		}
		pushes = nil
		server, err := NewServer(Options{
			Host:        "127.0.0.1",
			RepoDir:     GinkgoT().TempDir(),
			HostKeyPath: filepath.Join(GinkgoT().TempDir(), "ssh_host_key"),
			Authorizer:  authorizer,
			OnPushed: func(_ context.Context, push Push) {
				pushesMu.Lock()
				defer pushesMu.Unlock()
				fmt.Fprintln(push.Out, "push received")
				push.Out = nil
				pushes = append(pushes, push)
			},
		})
		Expect(err).NotTo(HaveOccurred())
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		go func() {
			defer GinkgoRecover()
			_ = server.Serve(listener)
		}()
		DeferCleanup(server.Close)

		clientRepo = GinkgoT().TempDir()
		for _, args := range [][]string{
			{"init", "-b", "main"}, {"config", "user.email", "t@t.com"}, {"config", "user.name", "T"},
			{"config", "commit.gpgsign", "false"}, {"commit", "--allow-empty", "-m", "initial"},
		} {
			gitIn(clientRepo, args...)
		}
		sshCmd = fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o IdentitiesOnly=yes -p %d",
			keyPath, listener.Addr().(*net.TCPAddr).Port)
	})

	push := func(path string) (string, error) {
		cmd := exec.Command("git", "push", "ssh://127.0.0.1/"+path, "main")
		cmd.Dir = clientRepo
		cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	It("refuses an unknown key", func() {
		authorizer.key = nil
		out, err := push("acme/app")
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("Permission denied"))
		Expect(authorizer.asked).To(BeEmpty())
	})

	It("refuses a repository the user may not push to, without saying which reason", func() {
		out, err := push("acme/other")
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("repository not found or access denied"))
		Expect(authorizer.asked).To(Equal([]string{"acme/other"}))
		Expect(bareDir).NotTo(BeADirectory())
	})

	It("receives into the authorized repository without running its hooks and reports the refs", func() {
		gitIn("", "init", "--bare", bareDir)
		marker := filepath.Join(GinkgoT().TempDir(), "hook-ran")
		hook := filepath.Join(bareDir, "hooks", "pre-receive")
		Expect(os.WriteFile(hook, []byte("#!/bin/sh\ntouch "+marker+"\n"), 0o755)).To(Succeed())

		out, err := push("acme/app")
		Expect(err).NotTo(HaveOccurred(), out)
		Expect(out).To(ContainSubstring("push received"), "OnPushed writes to the pusher's terminal")
		Expect(marker).NotTo(BeAnExistingFile(), "hooks never run on the host")

		head := strings.TrimSpace(gitIn(clientRepo, "rev-parse", "HEAD"))
		Expect(strings.TrimSpace(gitIn("", "--git-dir", bareDir, "rev-parse", "refs/heads/main"))).To(Equal(head))
		pushesMu.Lock()
		defer pushesMu.Unlock()
		Expect(pushes).To(HaveLen(1))
		Expect(pushes[0].OrgID).To(Equal(authorizer.target.OrgID))
		Expect(pushes[0].RepoID).To(Equal(authorizer.target.RepoID))
		Expect(pushes[0].UserID).To(Equal(authorizer.user))
		Expect(pushes[0].Refs).To(Equal([]RefUpdate{{Name: "refs/heads/main", Old: "", New: head}}))
	})

	It("creates the authorized repository on first push", func() {
		out, err := push("acme/app")
		Expect(err).NotTo(HaveOccurred(), out)
		Expect(filepath.Join(bareDir, "HEAD")).To(BeAnExistingFile())
	})
})

func gitIn(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
	return string(out)
}
