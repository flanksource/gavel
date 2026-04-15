package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/serve"
	"github.com/flanksource/gavel/verify"
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

// hookSpecFromConfig translates a resolved GavelConfig into the serve package's
// declarative HookSpec so serve/ stays free of a verify import.
func hookSpecFromConfig(cfg verify.GavelConfig) serve.HookSpec {
	toSteps := func(in []verify.HookStep) []serve.HookStep {
		if len(in) == 0 {
			return nil
		}
		out := make([]serve.HookStep, 0, len(in))
		for _, s := range in {
			out = append(out, serve.HookStep{Name: s.Name, Run: s.Run})
		}
		return out
	}
	return serve.HookSpec{
		Pre:  toSteps(cfg.Pre),
		Cmd:  cfg.SSH.Cmd,
		Post: toSteps(cfg.Post),
	}
}

func runServe(opts ServeOptions) (any, error) {
	// Load .gavel.yaml once at startup (not per-push): the SSH server is a
	// long-lived process and a config change should require a restart, same
	// as every other daemon setting.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	gavelCfg, err := verify.LoadGavelConfig(cwd)
	if err != nil {
		return nil, err
	}
	spec := hookSpecFromConfig(gavelCfg)
	logger.V(1).Infof("SSH post-receive hook: cmd=%q pre=%d post=%d",
		spec.Cmd, len(spec.Pre), len(spec.Post))

	srv, err := serve.NewServer(serve.Options{
		Host:        opts.Host,
		Port:        opts.Port,
		HostKeyPath: opts.HostKeyPath,
		RepoDir:     opts.RepoDir,
		HookWriter:  serve.NewHookWriter(spec),
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
