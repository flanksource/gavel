package serve

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"crypto/x509"

	"github.com/flanksource/commons/logger"
	gossh "golang.org/x/crypto/ssh"

	"github.com/gliderlabs/ssh"
)

type Options struct {
	Host        string
	Port        int
	HostKeyPath string
	RepoDir     string
	HookWriter  func(bareRepo, gavelPath string) error
}

type Server struct {
	opts       Options
	srv        *ssh.Server
	hookWriter func(bareRepo, gavelPath string) error
}

func NewServer(opts Options) (*Server, error) {
	if opts.RepoDir == "" {
		home, _ := os.UserHomeDir()
		opts.RepoDir = filepath.Join(home, ".gavel", "repos")
	}
	if err := os.MkdirAll(opts.RepoDir, 0o755); err != nil {
		return nil, fmt.Errorf("create repo dir: %w", err)
	}

	hostKey, err := loadOrGenerateHostKey(opts.HostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("host key: %w", err)
	}

	s := &Server{opts: opts, hookWriter: opts.HookWriter}
	if s.hookWriter == nil {
		s.hookWriter = writePostReceiveHook
	}
	s.srv = &ssh.Server{
		Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Handler: s.handleSession,
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true // accept all keys for local dev
		},
	}
	s.srv.AddHostKey(hostKey)
	return s, nil
}

func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.opts.Host, s.opts.Port)
	logger.Infof("SSH server listening on %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return s.srv.Serve(ln)
}

func (s *Server) Serve(ln net.Listener) error {
	return s.srv.Serve(ln)
}

func (s *Server) Close() error {
	return s.srv.Close()
}

func (s *Server) handleSession(sess ssh.Session) {
	cmd := sess.Command()
	logger.V(1).Infof("SSH session from %s, command: %v", sess.RemoteAddr(), cmd)
	if len(cmd) < 2 || cmd[0] != "git-receive-pack" {
		logger.Infof("Rejected unsupported command from %s: %v", sess.RemoteAddr(), cmd)
		fmt.Fprintf(sess.Stderr(), "unsupported command: %v\n", cmd)
		sess.Exit(1) //nolint:errcheck
		return
	}

	repoPath := cleanRepoPath(cmd[1])
	logger.Infof("Receiving push for %s from %s", repoPath, sess.RemoteAddr())
	exitCode := HandleGitReceive(sess, repoPath, s.opts.RepoDir, s.hookWriter)
	logger.V(1).Infof("Push for %s completed with exit code %d", repoPath, exitCode)
	sess.Exit(exitCode) //nolint:errcheck
}

func loadOrGenerateHostKey(path string) (gossh.Signer, error) {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".gavel", "ssh_host_key")
	}

	if data, err := os.ReadFile(path); err == nil {
		logger.V(1).Infof("Loaded SSH host key from %s", path)
		return gossh.ParsePrivateKey(data)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	if err := os.WriteFile(path, pemBlock, 0o600); err != nil {
		return nil, err
	}

	logger.Infof("Generated SSH host key at %s", path)
	return gossh.ParsePrivateKey(pemBlock)
}

func cleanRepoPath(raw string) string {
	// git-receive-pack sends the path with surrounding quotes
	path := raw
	if len(path) >= 2 && path[0] == '\'' && path[len(path)-1] == '\'' {
		path = path[1 : len(path)-1]
	}
	// Strip leading slash
	for len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if path == "" {
		path = "default"
	}
	return path
}
