package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/serve"
)

// gavelVersionBanner returns the same version string printed by
// `gavel version`. Shared by `gavel ssh serve` startup and the rendered
// post-receive hook so both show the same build identifiers.
func gavelVersionBanner() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)",
		version, commit, date, runtime.Version())
}

type ServeOptions struct {
	Port        int    `flag:"port" help:"SSH server port" default:"2222"`
	Host        string `flag:"host" help:"Listen address" default:"0.0.0.0"`
	HostKeyPath string `flag:"host-key" help:"Path to SSH host key (default: ~/.gavel/ssh_host_key)"`
	RepoDir     string `flag:"repo-dir" help:"Directory for cached bare repos (default: ~/.gavel/repos)"`
}

func (o ServeOptions) Help() string {
	return `Start an SSH server that accepts git push and runs gavel test --lint.

Developers add this as a git remote and push to trigger linting and tests:

  gavel ssh serve --port 2222

Then from any project:

  git remote add gavel ssh://localhost:2222/myproject
  git push gavel HEAD:main

Results stream back in real-time. Push is rejected on failure.
Repos are cached for fast incremental pushes.`
}

func init() {
	clicky.AddNamedCommand("serve", sshCmd, ServeOptions{}, runServe)
}

func runServe(opts ServeOptions) (any, error) {
	logger.Infof("gavel %s", gavelVersionBanner())

	srv, err := serve.NewServer(serve.Options{
		Host:        opts.Host,
		Port:        opts.Port,
		HostKeyPath: opts.HostKeyPath,
		RepoDir:     opts.RepoDir,
	})
	if err != nil {
		return nil, err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		logger.Infof("Shutting down SSH server...")
		srv.Close()
	}()

	return nil, srv.ListenAndServe()
}
