// Runs sandbox jobs on agent-substrate through captain's substrate sandbox, as runs of the bash
// model: the job's command runs in a checkout of its commit, reaching serve's run API over the run's
// channel, with egress limited to an allowlist and model keys injected outside.
package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/ai"
	// The bash model a job runs on, and every provider an agent run's model may name.
	_ "github.com/flanksource/captain/pkg/ai/provider"
	"github.com/flanksource/captain/pkg/ai/provider/bash"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	// The substrate sandbox a job's run executes in.
	_ "github.com/flanksource/captain/pkg/sandbox/substraterun"
	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sandbox/runclient"
)

// DefaultSubstrateEgress are the hosts a job or agent run reaches besides its model APIs: the Go,
// npm and Python registries, and GitHub. They are passed through without decryption.
var DefaultSubstrateEgress = []string{
	"proxy.golang.org", "sum.golang.org",
	"registry.npmjs.org",
	"pypi.org", "files.pythonhosted.org",
	"github.com", "codeload.github.com", "objects.githubusercontent.com",
}

// SubstrateOptions configures the agent-substrate executor: the substrate it reaches, what jobs may
// reach besides DefaultSubstrateEgress and their model APIs, and the run API that answers them.
type SubstrateOptions struct {
	substrate.Options
	// Egress are hosts jobs may reach besides DefaultSubstrateEgress.
	Egress []string
	// RunAPI is serve's run API, which answers what a job's program asks of its run, its spec and
	// its results, as the run, over the run's channel.
	RunAPI http.Handler
	// Program is the environment's program, which runs the job's run there; empty is gavel's.
	Program []string
}

// substrateProgram is gavel's program in a substrate environment, which runs a whole run there.
var substrateProgram = []string{"gavel", "runner", "run-task"}

type substrateExecutor struct {
	options SubstrateOptions
	client  *substrate.Client
}

// NewSubstrate returns an executor that runs jobs on agent-substrate through captain's substrate
// sandbox.
func NewSubstrate(options SubstrateOptions) (Executor, error) {
	if options.RunAPI == nil {
		return nil, errors.New("the substrate executor answers its jobs' run API requests over each run's channel: SubstrateOptions.RunAPI is required")
	}
	if options.Atespace == "" {
		options.Atespace = "ate-env"
	}
	if len(options.Program) == 0 {
		options.Program = substrateProgram
	}
	client, err := substrate.Dial(options.Options)
	if err != nil {
		return nil, err
	}
	return &substrateExecutor{options: options, client: client}, nil
}

func (e *substrateExecutor) Name() string { return "substrate" }

// Submit starts job's run on the bash model in captain's substrate sandbox: its command runs in
// environment JobName(run), in a checkout of its commit, with what its program asks of its run
// answered by the executor's run API, as the run, over the run's channel. The environment holds
// neither the run's token nor a route to serve.
func (e *substrateExecutor) Submit(ctx context.Context, job Job) (Handle, error) {
	if job.Source == nil {
		return nil, fmt.Errorf("sandbox run %s names no source: a substrate job runs in a checkout of its commit, which its dispatcher's Sources finds", job.RunID)
	}
	env, token, err := substrateJobEnv(job)
	if err != nil {
		return nil, err
	}
	timeout := job.Limits.withDefaults().Timeout
	options := SubstrateBackend(e.options.Options, e.options.Egress)
	options["environment"] = JobName(job.RunID)
	options["ref"] = job.Source.Commit
	options["command"] = e.options.Program
	options["env"] = env
	options["credentials"] = SubstrateCredentialOptions(job.Credentials)
	options["maxLifetime"] = (timeout + reportTime).String()
	logger.Infof("sandbox run %s in substrate environment %s/%s: the egress gateway injects %s", job.RunID, e.options.Atespace, JobName(job.RunID), injections(job.Credentials))

	handle := newSubstrateHandle(JobName(job.RunID))
	spec := api.Spec{Model: api.Model{Name: "bash"}, Prompt: api.Prompt{User: shellCommand(job.Command)}, Sandbox: &api.SandboxRef{Mode: api.SandboxSubstrate}}
	// The command runs at the top of the environment's checkout of the repository.
	spec.SetCwd(job.Source.GitDir)
	in := promptrun.Input{
		Resolved:          api.ResolvedSpec{Spec: spec},
		Config:            ai.Config{SandboxSelection: &api.SandboxConfig{Kind: api.SandboxSubstrate, Name: "gavel-jobs", Options: options}},
		Repo:              job.Source.GitDir,
		EnvironmentOutput: handle.log,
		OnEvent: func(_ int, event ai.Event) {
			if event.Kind == ai.EventText {
				_, _ = handle.log.Write([]byte(event.Text))
			}
		},
		// A job's run leaves no transcripts: a command is no agent.
		OnTranscript: func(string, string, []byte) {},
		HostCalls: map[string]func(context.Context, json.RawMessage) (json.RawMessage, error){
			runclient.HostCallName: runclient.AnswerHostCall(e.options.RunAPI, job.RunID.String(), token, runclient.JobReach),
		},
		Timeout: timeout,
	}
	runCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	handle.cancel = cancel
	go handle.run(runCtx, in)
	return handle, nil
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

// substrateJobEnv is the job's settings for its environment, and the run token its run's channel
// is answered with. The job reaches serve over that channel, never the network, so serve's address
// and the token stay out of the environment; any other secret would enter it, so it is refused.
func substrateJobEnv(job Job) (map[string]string, string, error) {
	env := maps.Clone(job.Env)
	delete(env, "GAVEL_SERVER")
	for key := range job.SecretEnv {
		if key != "GAVEL_RUN_TOKEN" {
			return nil, "", fmt.Errorf("secret %s would enter the substrate environment; give it to substrate's egress gateway as an injected credential instead", key)
		}
	}
	return env, job.SecretEnv["GAVEL_RUN_TOKEN"], nil
}

// SubstrateBackend is the substrate backend runs execute on through captain's substrate sandbox, as
// its options spell it: how to reach the substrate, the image template, and what a run may reach
// besides its model APIs, DefaultSubstrateEgress and hosts, over TLS on 443, passed through. A run
// doesn't reach serve: it asks serve over its own channel. Each run adds its own options.
func SubstrateBackend(options substrate.Options, hosts []string) map[string]any {
	egress := make([]string, 0, len(DefaultSubstrateEgress)+len(hosts))
	for _, host := range append(append([]string{}, DefaultSubstrateEgress...), hosts...) {
		egress = append(egress, net.JoinHostPort(host, "443"))
	}
	backend := map[string]any{"egress": egress}
	for key, value := range map[string]string{
		"envAPI": options.EnvAPI, "control": options.Control, "tokenFile": options.ControlTokenFile,
		"caFile": options.CAFile, "clientCert": options.ClientCertFile, "clientKey": options.ClientKeyFile,
		"atespace": options.Atespace, "template": options.Template, "trustBundle": options.TrustBundle,
	} {
		if value != "" {
			backend[key] = value
		}
	}
	return backend
}

// SubstrateCredentialOptions are credentials as captain's substrate sandbox takes a run's: each
// named in full, with the URI its key is resolved from.
func SubstrateCredentialOptions(credentials []substrate.Credential) []any {
	named := make([]any, len(credentials))
	for i, credential := range credentials {
		named[i] = map[string]any{
			"host": credential.Host, "header": credential.Header, "prefix": credential.Prefix, "env": credential.Env, "uri": credential.URI,
		}
	}
	return named
}

// shellWord is an argument the shell takes as it is.
var shellWord = regexp.MustCompile(`^[A-Za-z0-9_./:=@%+,-]+$`)

// shellCommand is command as one bash command line, each argument quoted unless the shell takes
// it as it is.
func shellCommand(command []string) string {
	words := make([]string, len(command))
	for i, arg := range command {
		if shellWord.MatchString(arg) {
			words[i] = arg
			continue
		}
		words[i] = "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
	}
	return strings.Join(words, " ")
}

// Sweep deletes the environments of runs, gavel-run-<run>, created more than lifetime ago: what
// jobs and agent runs left behind when serve died before their end.
func (e *substrateExecutor) Sweep(ctx context.Context, lifetime time.Duration) ([]string, error) {
	return e.client.Sweep(ctx, jobPrefix, time.Now().Add(-lifetime))
}

// substrateHandle is a job's run: its log as the run writes it, and how it ended.
type substrateHandle struct {
	ref    string
	log    *runLog
	logs   *io.PipeReader
	taken  sync.Once
	done   chan struct{}
	cancel context.CancelFunc
	result Result
	err    error
}

func newSubstrateHandle(ref string) *substrateHandle {
	reader, writer := io.Pipe()
	return &substrateHandle{ref: ref, log: &runLog{to: writer}, logs: reader, done: make(chan struct{})}
}

// run runs in, then ends the log and records how the job ended.
func (h *substrateHandle) run(ctx context.Context, in promptrun.Input) {
	defer close(h.done)
	defer h.cancel()
	result, err := promptrun.Run(ctx, in)
	_ = h.log.to.Close()
	h.result, h.err = jobResult(h.ref, result, err)
}

// jobResult is how the job's run ended: the exit code its bash command reported.
func jobResult(env string, result promptrun.Result, err error) (Result, error) {
	if err != nil {
		return Result{}, fmt.Errorf("the job in substrate environment %s: %w", env, err)
	}
	if result.Response == nil || result.Response.StructuredData == nil {
		return Result{}, fmt.Errorf("the job in substrate environment %s ended without reporting its exit code", env)
	}
	data, err := json.Marshal(result.Response.StructuredData)
	if err != nil {
		return Result{}, fmt.Errorf("read the outcome of the job in substrate environment %s: %w", env, err)
	}
	var outcome bash.Outcome
	if err := json.Unmarshal(data, &outcome); err != nil {
		return Result{}, fmt.Errorf("read the outcome of the job in substrate environment %s: %w", env, err)
	}
	return Result{ExitCode: outcome.ExitCode}, nil
}

func (h *substrateHandle) Ref() string { return h.ref }

// Logs follows the job's output, and its environment's, until the job ends. It can be followed
// once.
func (h *substrateHandle) Logs(context.Context) (io.ReadCloser, error) {
	taken := false
	h.taken.Do(func() { taken = true })
	if !taken {
		return nil, fmt.Errorf("the log of the job in substrate environment %s is already followed", h.ref)
	}
	return h.logs, nil
}

// discardUnfollowedLog reads the job's log to the end when nothing follows it, so the run's writes
// to it never block.
func (h *substrateHandle) discardUnfollowedLog() {
	h.taken.Do(func() { go func() { _, _ = io.Copy(io.Discard, h.logs) }() })
}

// Wait returns the job's exit code once its run ends, which deletes its environment. The job's run
// stands when the environment can't be deleted, which captain warns of.
func (h *substrateHandle) Wait(ctx context.Context) (Result, error) {
	h.discardUnfollowedLog()
	select {
	case <-h.done:
		return h.result, h.err
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

// Cancel stops the job's run, which kills the job and deletes its environment, and waits for it
// to end.
func (h *substrateHandle) Cancel(ctx context.Context) error {
	h.discardUnfollowedLog()
	h.cancel()
	select {
	case <-h.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// runLog is a job's log, which its command's output and its environment's are written to at once.
type runLog struct {
	mu sync.Mutex
	to *io.PipeWriter
}

func (l *runLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.to.Write(p)
}
