package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/serve"
)

type ServeOptions struct {
	Port        int    `flag:"port" help:"SSH server port" default:"2222"`
	Host        string `flag:"host" help:"Listen address" default:"0.0.0.0"`
	HostKeyPath string `flag:"host-key" help:"Path to SSH host key (default: ~/.gavel/ssh_host_key)"`
	RepoDir     string `flag:"repo-dir" help:"Directory for cached bare repos (default: ~/.gavel/repos)"`
}

func (o ServeOptions) Help() string {
	return `Start an SSH server that accepts git push and runs gavel test --lint.

Developers add this as a git remote and push to trigger linting and tests:

  gavel serve --port 2222

Then from any project:

  git remote add gavel ssh://localhost:2222/myproject
  git push gavel HEAD:main

Results stream back in real-time. Push is rejected on failure.
Repos are cached for fast incremental pushes.`
}

func init() {
	clicky.AddNamedCommand("serve", rootCmd, ServeOptions{}, runServe)
}

func runServe(opts ServeOptions) (any, error) {
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
