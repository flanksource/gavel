package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	shlex "github.com/anmitsu/go-shlex"
	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/fixtures"
)

// AgentBrowserConfig is a `kind: agent-browser` step: agent-browser commands
// run in order, attached over CDP to the fixture's Chrome (never launching
// their own), in the BROWSER_SESSION session.
type AgentBrowserConfig struct {
	Commands []string `json:"commands" jsonschema:"required,minItems=1,description=agent-browser commands without the binary (e.g. 'open $TODO_URL' or 'snapshot'); $X expands from the fixture env"`
	Observe  string   `json:"observe,omitempty" jsonschema:"description=Command re-run on each wait tick (e.g. snapshot)"`
	Session  string   `json:"session,omitempty" jsonschema:"description=agent-browser session (default $BROWSER_SESSION)"`
	CDP      string   `json:"cdp,omitempty" jsonschema:"description=CDP endpoint to attach to (default $BROWSER_CDP_URL)"`
	Binary   string   `json:"binary,omitempty" jsonschema:"description=agent-browser executable (default: PATH, then ~/Library/pnpm/bin)"`
}

// AgentBrowser is the `agent-browser` kind.
type AgentBrowser struct{}

func (AgentBrowser) Name() string            { return "agent-browser" }
func (AgentBrowser) Schema() json.RawMessage { return schemaOf(&AgentBrowserConfig{}) }

// agentBrowserSession is how every command of a step invokes agent-browser;
// it is also the Observe handle.
type agentBrowserSession struct {
	binary, session, cdp string
	env                  map[string]string
	observe              string
}

// commandRun is one executed agent-browser command.
type commandRun struct {
	Cmd      string
	ExitCode int
	Stdout   string
	Stderr   string
}

func (r commandRun) asMap() map[string]any {
	return map[string]any{"cmd": r.Cmd, "exitCode": r.ExitCode, "stdout": r.Stdout, "stderr": r.Stderr}
}

func (AgentBrowser) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	start := time.Now()
	cfg, session, err := agentBrowserSetup(step)
	if err != nil {
		return fixtures.StepOutcome{}, err
	}
	result := newResult(step, start)
	var stdout strings.Builder
	var snapshot string
	results := make([]map[string]any, 0, len(cfg.Commands))
	for _, command := range cfg.Commands {
		run, err := session.run(ctx, step, command)
		if err != nil {
			return fixtures.StepOutcome{}, err
		}
		results = append(results, run.asMap())
		stdout.WriteString(run.Stdout)
		if isSnapshot(run.Cmd) {
			snapshot = run.Stdout
		}
		result.Command = session.binary + " " + run.Cmd
		if run.ExitCode != 0 {
			result = result.Failf("agent-browser %s exited %d: %s", run.Cmd, run.ExitCode, strings.TrimSpace(run.Stderr+"\n"+run.Stdout))
			break
		}
	}
	result.Stdout = stdout.String()
	outcome := fixtures.StepOutcome{
		Result:  finish(result),
		Outputs: map[string]any{"stdout": stdout.String(), "snapshot": snapshot, "results": results},
		Handle:  session,
	}
	// The first step to use a session owns its daemon's teardown; later steps
	// in the same session reuse it.
	if _, loaded := openSessions.LoadOrStore(session.session, true); !loaded {
		outcome.Cleanup = session.close
	}
	return outcome, nil
}

// openSessions are agent-browser sessions with a teardown registered.
var openSessions sync.Map

// close ends the agent-browser session daemon. Attached over CDP, agent-browser
// does not own the Chrome, so this disconnects without closing the browser.
func (s *agentBrowserSession) close(ctx context.Context) error {
	defer openSessions.Delete(s.session)
	args := []string{"--engine", "chrome", "--session", s.session, "close"}
	execution := cexec.NewExec(s.binary, args...).WithExactEnv(s.env).WithContext(ctx).Run().Result()
	if execution.ExitCode != 0 {
		return fmt.Errorf("agent-browser %s exited %d: %s", strings.Join(args, " "), execution.ExitCode, strings.TrimSpace(execution.Stderr+"\n"+execution.Stdout))
	}
	return nil
}

// Observe re-runs the step's `observe:` command; its output is `stdout`, and
// also `snapshot` when it is a snapshot.
func (AgentBrowser) Observe(ctx context.Context, step *fixtures.StepContext, handle any) (map[string]any, error) {
	session, ok := handle.(*agentBrowserSession)
	if !ok {
		return nil, fmt.Errorf("agent-browser observe: unexpected handle %T", handle)
	}
	if session.observe == "" {
		return nil, fmt.Errorf("agent-browser step %q: wait: needs an observe: command to re-run", step.Spec.ID)
	}
	run, err := session.run(ctx, step, session.observe)
	if err != nil {
		return nil, err
	}
	if run.ExitCode != 0 {
		return nil, fmt.Errorf("agent-browser %s exited %d: %s", run.Cmd, run.ExitCode, strings.TrimSpace(run.Stderr))
	}
	outputs := map[string]any{"stdout": run.Stdout, "exitCode": run.ExitCode}
	if isSnapshot(run.Cmd) {
		outputs["snapshot"] = run.Stdout
	}
	return outputs, nil
}

func agentBrowserSetup(step *fixtures.StepContext) (AgentBrowserConfig, *agentBrowserSession, error) {
	cfg, err := fixtures.KindConfig[AgentBrowserConfig](step)
	if err != nil {
		return cfg, nil, err
	}
	if len(cfg.Commands) == 0 {
		return cfg, nil, fmt.Errorf("agent-browser step %q: commands: is empty", step.Spec.ID)
	}
	for _, field := range []struct {
		name  string
		value *string
	}{{"session", &cfg.Session}, {"cdp", &cfg.CDP}, {"binary", &cfg.Binary}} {
		if *field.value, err = render(step, field.name, *field.value); err != nil {
			return cfg, nil, err
		}
	}
	const hint = "run a `kind: browser` step first (agent-browser only attaches to the fixture's Chrome)"
	if cfg.CDP == "" {
		if cfg.CDP, err = requireEnv(step, "BROWSER_CDP_URL", hint); err != nil {
			return cfg, nil, err
		}
	}
	if cfg.Session == "" {
		if cfg.Session, err = requireEnv(step, "BROWSER_SESSION", hint); err != nil {
			return cfg, nil, err
		}
	}
	binary, err := resolveAgentBrowser(cfg.Binary)
	if err != nil {
		return cfg, nil, fmt.Errorf("agent-browser step %q: %w", step.Spec.ID, err)
	}
	return cfg, &agentBrowserSession{binary: binary, session: cfg.Session, cdp: cfg.CDP, env: environ(step), observe: cfg.Observe}, nil
}

func (s *agentBrowserSession) run(ctx context.Context, step *fixtures.StepContext, command string) (commandRun, error) {
	rendered, err := render(step, "commands", command)
	if err != nil {
		return commandRun{}, err
	}
	words, err := shlex.Split(rendered, true)
	if err != nil {
		return commandRun{}, fmt.Errorf("agent-browser step %q: parse command %q: %w", step.Spec.ID, rendered, err)
	}
	if len(words) == 0 {
		return commandRun{}, fmt.Errorf("agent-browser step %q: empty command", step.Spec.ID)
	}
	// --engine is always explicit: the CLI default is chrome, and a session
	// must keep one engine for its whole life.
	args := append([]string{"--engine", "chrome", "--session", s.session, "--cdp", s.cdp}, words...)
	execution := cexec.NewExec(s.binary, args...).WithExactEnv(s.env).WithContext(ctx).Run().Result()
	if execution.ExitCode < 0 {
		return commandRun{}, fmt.Errorf("agent-browser step %q: run %s: %w", step.Spec.ID, rendered, execution.Error)
	}
	return commandRun{Cmd: rendered, ExitCode: execution.ExitCode, Stdout: execution.Stdout, Stderr: execution.Stderr}, nil
}

func isSnapshot(command string) bool {
	return command == "snapshot" || strings.HasPrefix(command, "snapshot ")
}

var cdpSupport sync.Map // binary → error (nil when --cdp is supported)

// resolveAgentBrowser finds agent-browser and checks it can attach over CDP.
func resolveAgentBrowser(binary string) (string, error) {
	candidates := []string{binary}
	if binary == "" {
		candidates = []string{"agent-browser"}
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, filepath.Join(home, "Library", "pnpm", "bin", "agent-browser"))
		}
	}
	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}
		supported, _ := cdpSupport.LoadOrStore(path, checkCDPFlag(path))
		if supported != nil {
			return "", supported.(error)
		}
		return path, nil
	}
	return "", fmt.Errorf("agent-browser not found (tried %s); install it with `pnpm add -g agent-browser`", strings.Join(candidates, ", "))
}

func checkCDPFlag(path string) error {
	help := cexec.NewExec(path, "--help").Run().Result()
	if !strings.Contains(help.Stdout+help.Stderr, "--cdp <") {
		return fmt.Errorf("%s does not support --cdp (CDP attach); upgrade agent-browser", path)
	}
	return nil
}
