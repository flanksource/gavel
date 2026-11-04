// Runs sandbox jobs on agent-substrate: each job is an environment created from the runner image,
// whose one process is the job, with egress limited to an allowlist and model keys injected outside.
package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/commons/logger"
)

// DefaultSubstrateEgress are the hosts a job reaches besides gavel serve and the model APIs: the
// Go, npm and Python registries, and GitHub. They are passed through without decryption.
var DefaultSubstrateEgress = []string{
	"proxy.golang.org", "sum.golang.org",
	"registry.npmjs.org",
	"pypi.org", "files.pythonhosted.org",
	"github.com", "codeload.github.com", "objects.githubusercontent.com",
}

// SubstrateOptions configures the agent-substrate executor: the substrate it reaches, and what
// jobs may reach besides serve, DefaultSubstrateEgress and their model APIs.
type SubstrateOptions struct {
	substrate.Options
	// Egress are hosts jobs may reach besides DefaultSubstrateEgress.
	Egress []string
}

type substrateExecutor struct {
	options SubstrateOptions
	client  *substrate.Client
}

// NewSubstrate returns an executor that runs jobs as agent-substrate environments.
func NewSubstrate(options SubstrateOptions) (Executor, error) {
	if options.Atespace == "" {
		options.Atespace = "ate-env"
	}
	client, err := substrate.Dial(options.Options)
	if err != nil {
		return nil, err
	}
	return &substrateExecutor{options: options, client: client}, nil
}

func (e *substrateExecutor) Name() string { return "substrate" }

func (e *substrateExecutor) Submit(ctx context.Context, job Job) (Handle, error) {
	process, err := substrateProcess(job)
	if err != nil {
		return nil, err
	}
	egress, err := substrateEgress(e.options.Egress, job.Env["GAVEL_SERVER"], job.Credentials)
	if err != nil {
		return nil, err
	}
	env, err := e.client.Create(ctx, substrate.Environment{ID: JobName(job.RunID), Egress: egress})
	if err != nil {
		return nil, err
	}
	logger.Infof("sandbox run %s in substrate environment %s/%s: the egress gateway injects %s", job.RunID, e.options.Atespace, env.ID(), injections(job.Credentials))
	proc, err := env.Start(ctx, process)
	if err != nil {
		return nil, errors.Join(err, env.Delete(context.WithoutCancel(ctx)))
	}
	return &substrateHandle{env: env, proc: proc}, nil
}

// injections names each credential the egress gateway puts into a job's requests by its host,
// header and URI, none of them secret, so a run missing its key shows whether gavel asked for it.
func injections(credentials []substrate.Credential) string {
	if len(credentials) == 0 {
		return "no credentials"
	}
	named := make([]string, len(credentials))
	for i, credential := range credentials {
		named[i] = credential.Host + " " + credential.Header + " from " + credential.URI
	}
	return strings.Join(named, ", ")
}

// substrateProcess is the job's process: its command and settings, and its run token. Any other
// secret would enter the environment, so it is refused; the client gives each injected credential's
// variable a placeholder.
func substrateProcess(job Job) (substrate.Process, error) {
	env := map[string]string{}
	for key, value := range job.Env {
		env[key] = value
	}
	for key, value := range job.SecretEnv {
		if key != "GAVEL_RUN_TOKEN" {
			return substrate.Process{}, fmt.Errorf("secret %s would enter the substrate environment; give it to substrate's egress gateway as an injected credential instead", key)
		}
		env[key] = value
	}
	return substrate.Process{Command: job.Command, Env: env, Timeout: job.Limits.withDefaults().Timeout}, nil
}

// substrateEgress allows a job to reach gavel serve at server, the hosts it needs, and the model
// APIs of its credentials, which the gateway injects. Everything else is denied.
// Sweep deletes the environments of runs, gavel-run-<run>, created more than lifetime ago: what
// jobs and agent runs left behind when serve died before their end.
func (e *substrateExecutor) Sweep(ctx context.Context, lifetime time.Duration) ([]string, error) {
	return e.client.Sweep(ctx, jobPrefix, time.Now().Add(-lifetime))
}

// SubstrateAgentEgress is what an agent run may reach besides its model APIs, spelled as captain's
// substrate sandbox takes egress: DefaultSubstrateEgress and hosts, over TLS on 443, passed
// through. Unlike a job, an agent run doesn't reach serve: it asks serve over its own channel.
func SubstrateAgentEgress(hosts []string) []string {
	egress := make([]string, 0, len(DefaultSubstrateEgress)+len(hosts))
	for _, host := range append(append([]string{}, DefaultSubstrateEgress...), hosts...) {
		egress = append(egress, net.JoinHostPort(host, "443"))
	}
	return egress
}

func substrateEgress(hosts []string, server string, credentials []substrate.Credential) (substrate.Egress, error) {
	serve, err := url.Parse(server)
	if err != nil || serve.Hostname() == "" {
		return substrate.Egress{}, fmt.Errorf("substrate egress names hosts: GAVEL_SERVER %q isn't a URL", server)
	}
	if net.ParseIP(serve.Hostname()) != nil {
		return substrate.Egress{}, fmt.Errorf("substrate egress names hosts: GAVEL_SERVER %q must name serve by DNS name, not an IP address", server)
	}
	port := uint64(80)
	if serve.Scheme == "https" {
		port = 443
	}
	if serve.Port() != "" {
		// url.Parse checks a port is digits, not that it is a port.
		port, err = strconv.ParseUint(serve.Port(), 10, 16)
		if err != nil || port == 0 {
			return substrate.Egress{}, fmt.Errorf("substrate egress: GAVEL_SERVER %q has port %s, outside 1-65535", server, serve.Port())
		}
	}
	egress := substrate.Egress{Credentials: credentials}
	serveHost := substrate.HostPort{Host: serve.Hostname(), Port: int(port)}
	if serve.Scheme == "https" {
		egress.Passthrough = append(egress.Passthrough, serveHost)
	} else {
		egress.HTTP = append(egress.HTTP, serveHost)
	}
	for _, host := range append(append([]string{}, DefaultSubstrateEgress...), hosts...) {
		egress.Passthrough = append(egress.Passthrough, substrate.HostPort{Host: host, Port: 443})
	}
	return egress, nil
}

type substrateHandle struct {
	env  *substrate.Env
	proc *substrate.Proc
}

func (h *substrateHandle) Ref() string { return h.env.ID() }

// Logs follows the job's stdout and stderr until it exits.
func (h *substrateHandle) Logs(ctx context.Context) (io.ReadCloser, error) {
	return h.proc.Output(ctx), nil
}

// Wait returns the job's exit code once it exits, then removes its environment. The job's run
// stands when the environment can't be removed, so that is a warning, not Wait's error.
func (h *substrateHandle) Wait(ctx context.Context) (Result, error) {
	code, err := h.proc.Wait(ctx)
	if err != nil {
		return Result{}, err
	}
	if err := h.env.Delete(context.WithoutCancel(ctx)); err != nil {
		logger.Warnf("the job in substrate environment %s exited %d, but deleting the environment failed, so it is left behind until it is deleted by hand: %v", h.env.ID(), code, err)
	}
	return Result{ExitCode: code}, nil
}

// Cancel kills the job and removes its environment.
func (h *substrateHandle) Cancel(ctx context.Context) error {
	if err := h.proc.Kill(ctx); err != nil {
		return err
	}
	return h.env.Delete(ctx)
}
