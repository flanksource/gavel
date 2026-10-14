package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// generatedDir is where inline `script:` bodies are written, relative to the
// Playwright package (cwd). It sits inside the Playwright testDir (e2e) so the
// spec is discovered and resolves @playwright/test and ../fixtures/cdp; its
// .gitignore keeps the transient files out of git.
const generatedDir = "e2e/.generated"

// PlaywrightConfig is a `kind: playwright` step. In the foreground it runs a
// Playwright spec with `pnpm exec playwright test`; with `background: true` it
// runs a node page watcher (e.g. e2e/watch-session.ts) until fixture teardown.
type PlaywrightConfig struct {
	File         string             `json:"file,omitempty" jsonschema:"description=Spec or watcher script, relative to the fixture file"`
	Script       string             `json:"script,omitempty" jsonschema:"description=Inline TypeScript spec/watcher; written under <cwd>/e2e/.generated and never templated"`
	Cwd          string             `json:"cwd,omitempty" jsonschema:"description=Package with @playwright/test (default: nearest ancestor of the spec or fixture that has it)"`
	URL          string             `json:"url,omitempty" jsonschema:"description=Exported to the script as TARGET_URL"`
	Args         []string           `json:"args,omitempty" jsonschema:"description=Extra 'playwright test' arguments"`
	ReadyTimeout *fixtures.Duration `json:"readyTimeout,omitempty" jsonschema:"description=Background: how long to wait for the watcher's ready event (default 60s)"`
	StopTimeout  *fixtures.Duration `json:"stopTimeout,omitempty" jsonschema:"description=Background: how long the watcher gets to exit after SIGTERM (default 10s)"`
}

// Playwright is the `playwright` kind.
type Playwright struct{}

func (Playwright) Name() string            { return "playwright" }
func (Playwright) Schema() json.RawMessage { return schemaOf(&PlaywrightConfig{}) }

// playwrightRun is a resolved playwright step.
type playwrightRun struct {
	cfg     PlaywrightConfig
	cwd     string
	script  string // absolute path of the spec/watcher
	env     map[string]string
	cleanup func() error // removes a generated inline script
}

func resolvePlaywright(step *fixtures.StepContext, ext string) (*playwrightRun, error) {
	cfg, err := fixtures.KindConfig[PlaywrightConfig](step)
	if err != nil {
		return nil, err
	}
	if (cfg.File == "") == (cfg.Script == "") {
		return nil, fmt.Errorf("playwright step %q: set exactly one of file or script", step.Spec.ID)
	}
	for _, field := range []struct {
		name  string
		value *string
	}{{"file", &cfg.File}, {"cwd", &cfg.Cwd}, {"url", &cfg.URL}} {
		if *field.value, err = render(step, field.name, *field.value); err != nil {
			return nil, err
		}
	}
	for i := range cfg.Args {
		if cfg.Args[i], err = render(step, "args", cfg.Args[i]); err != nil {
			return nil, err
		}
	}
	run := &playwrightRun{cfg: cfg, env: environ(step), cleanup: func() error { return nil }}
	if cfg.File != "" {
		run.script = resolvePath(step, cfg.File)
		if _, err := os.Stat(run.script); err != nil {
			return nil, fmt.Errorf("playwright step %q: file: %w", step.Spec.ID, err)
		}
	}
	if run.cwd, err = playwrightPackage(step, cfg.Cwd, run.script); err != nil {
		return nil, err
	}
	if cfg.Script != "" {
		if err := run.writeScript(step, ext); err != nil {
			return nil, err
		}
	}
	if cfg.URL != "" {
		run.env["TARGET_URL"] = cfg.URL
	}
	return run, nil
}

// playwrightPackage is the directory `playwright`/`node` run in: the
// configured cwd, else the nearest ancestor of the spec (or the fixture) with
// node_modules/@playwright/test.
func playwrightPackage(step *fixtures.StepContext, cwd, script string) (string, error) {
	if cwd != "" {
		cwd = resolvePath(step, cwd)
		if !hasPlaywright(cwd) {
			return "", fmt.Errorf("playwright step %q: cwd %s has no node_modules/@playwright/test (run pnpm install there)", step.Spec.ID, cwd)
		}
		return cwd, nil
	}
	start := fixtureDir(step)
	if script != "" {
		start = filepath.Dir(script)
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		if hasPlaywright(dir) {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("playwright step %q: no ancestor of %s has node_modules/@playwright/test; set cwd to the Playwright package", step.Spec.ID, start)
		}
	}
}

func hasPlaywright(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "node_modules", "@playwright", "test", "package.json"))
	return err == nil
}

func (r *playwrightRun) writeScript(step *fixtures.StepContext, ext string) error {
	dir := filepath.Join(r.cwd, generatedDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("playwright step %q: %w", step.Spec.ID, err)
	}
	r.script = filepath.Join(dir, stepName(step)+"-"+randomHex(4)+ext)
	if err := os.WriteFile(r.script, []byte(r.cfg.Script), 0o644); err != nil {
		return fmt.Errorf("playwright step %q: write inline script: %w", step.Spec.ID, err)
	}
	r.cleanup = func() error { return os.Remove(r.script) }
	return nil
}

func (Playwright) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	start := time.Now()
	run, err := resolvePlaywright(step, ".spec.ts")
	if err != nil {
		return fixtures.StepOutcome{}, err
	}
	spec, err := filepath.Rel(run.cwd, run.script)
	if err != nil {
		return fixtures.StepOutcome{}, errors.Join(err, run.cleanup())
	}
	report := filepath.Join(step.TmpDir, stepName(step)+"-playwright.json")
	run.env["PLAYWRIGHT_JSON_OUTPUT_FILE"] = report
	args := append([]string{"exec", "playwright", "test", spec, "--reporter=json"}, run.cfg.Args...)
	process := cexec.NewExec("pnpm", args...).WithCwd(run.cwd).WithExactEnv(run.env).WithProcessGroup().WithContext(ctx)
	execution := process.Run().Result()
	if err := run.cleanup(); err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("playwright step %q: remove inline script: %w", step.Spec.ID, err)
	}

	result := newResult(step, start)
	result.Command, result.CWD = "pnpm "+strings.Join(args, " "), run.cwd
	result.Stdout, result.Stderr, result.ExitCode = execution.Stdout, execution.Stderr, execution.ExitCode
	tests, reportErr := parsePlaywrightReport(report, run.cwd)
	outputs := playwrightOutputs(tests, execution)
	result.Children = playwrightChildren(tests)

	switch {
	case reportErr != nil:
		result = result.Failf("playwright produced no usable report: %v", reportErr)
	case execution.ExitCode != 0 || outputs["failed"].(int) > 0:
		result = result.Failf("playwright test exited %d: %d of %d tests failed", execution.ExitCode, outputs["failed"], outputs["total"])
	case outputs["total"].(int) == 0:
		result = result.Failf("playwright ran no tests from %s", spec)
	}
	return fixtures.StepOutcome{Result: finish(result), Outputs: outputs}, nil
}

// parsePlaywrightReport reuses gavel's Playwright JSON parser; a report with
// top-level errors (a spec that fails to compile) is an error.
func parsePlaywrightReport(path, workDir string) ([]parsers.Test, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var header struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(header.Errors) > 0 {
		messages := make([]string, 0, len(header.Errors))
		for _, e := range header.Errors {
			messages = append(messages, parsers.StripANSI(e.Message))
		}
		return nil, errors.New(strings.Join(messages, "\n"))
	}
	return parsers.NewPlaywrightJSON(workDir).Parse(strings.NewReader(string(raw)))
}

func playwrightTestName(t parsers.Test) string {
	if len(t.Suite) <= 1 {
		return t.Name
	}
	// Suite[0] is the spec file; the rest are describe() titles.
	return strings.Join(append(append([]string{}, t.Suite[1:]...), t.Name), " > ")
}

func playwrightStatus(t parsers.Test) task.Status {
	switch {
	case t.Failed:
		return task.StatusFAIL
	case t.Skipped:
		return task.StatusSKIP
	default:
		return task.StatusPASS
	}
}

func playwrightChildren(tests []parsers.Test) []*fixtures.FixtureNode {
	children := make([]*fixtures.FixtureNode, 0, len(tests))
	for _, t := range tests {
		name := playwrightTestName(t)
		children = append(children, &fixtures.FixtureNode{
			Name: name,
			Type: fixtures.TestNode,
			Results: &fixtures.FixtureResult{
				Name: name, Type: "playwright", Status: playwrightStatus(t), Duration: t.Duration,
				Error: t.Message, Stderr: t.Stderr, Test: fixtures.FixtureTest{Name: name},
			},
		})
	}
	return children
}

func playwrightOutputs(tests []parsers.Test, execution *cexec.ExecResult) map[string]any {
	counts := map[task.Status]int{}
	list := make([]map[string]any, 0, len(tests))
	for _, t := range tests {
		status := playwrightStatus(t)
		counts[status]++
		list = append(list, map[string]any{"name": playwrightTestName(t), "status": string(status), "error": t.Message})
	}
	return map[string]any{
		"exitCode": execution.ExitCode,
		"total":    len(tests),
		"passed":   counts[task.StatusPASS],
		"failed":   counts[task.StatusFAIL],
		"skipped":  counts[task.StatusSKIP],
		"tests":    list,
		"stdout":   execution.Stdout,
		"stderr":   execution.Stderr,
	}
}
