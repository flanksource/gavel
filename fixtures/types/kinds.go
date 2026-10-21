package types

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
)

// The kinds this package owns: every markdown fence/row lowers to one of them,
// and a YAML/JSON step selects one with `kind:`.
func init() {
	fixtures.RegisterKind(execKind{})
	fixtures.RegisterKind(runnerKind{name: fixtures.RunnerKindTest, run: RunTestStep})
	fixtures.RegisterKind(runnerKind{name: fixtures.RunnerKindLint, run: RunLintStep})
	fixtures.RegisterKind(NewAIKind(RunAIStep))
}

// execKind runs a command: a markdown fence or row as parsed, or a YAML/JSON
// step whose config names a script or an executable.
type execKind struct{}

func (execKind) Name() string { return "exec" }

func (execKind) Schema() json.RawMessage { return execSchema }

// RawConfigKeys keeps goProfiles args unrendered: their `{{.path}}` is the
// profile destination, known only once the command is about to run.
func (execKind) RawConfigKeys() []string { return []string{"goProfiles"} }

func (k execKind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	result, err := k.run(ctx, step)
	if err != nil {
		return fixtures.StepOutcome{}, err
	}
	return fixtures.StepOutcome{Result: result, Outputs: execOutputs(result)}, nil
}

// Observe re-runs the command, so `wait.until` polls over fresh
// stdout/stderr/exitCode. A non-zero exit is an observation; a command that
// could not be launched is an observation error.
func (k execKind) Observe(ctx context.Context, step *fixtures.StepContext, _ any) (map[string]any, error) {
	result, err := k.run(ctx, step)
	if err != nil {
		return nil, err
	}
	if result.Status == task.StatusERR {
		return execOutputs(result), fmt.Errorf("%s", result.Error)
	}
	return execOutputs(result), nil
}

func (execKind) run(ctx context.Context, step *fixtures.StepContext) (fixtures.FixtureResult, error) {
	test := step.Test
	if test.Step != nil {
		cfg, err := fixtures.KindConfig[execStepConfig](step)
		if err != nil {
			return fixtures.FixtureResult{}, err
		}
		if test, err = cfg.apply(test); err != nil {
			return fixtures.FixtureResult{}, fmt.Errorf("exec step %q: %w", step.Spec.ID, err)
		}
	}
	return (&ExecFixture{}).Run(ctx, test, step.Opts), nil
}

// execStepConfig is a YAML/JSON exec step. `script` runs under bash -c (or
// the named `exec`); otherwise `exec` + `args` run directly. exitCode, stdout
// and stderr are the same expectations a markdown fence declares; exitCode
// defaults to 0.
type execStepConfig struct {
	Script   string              `json:"script,omitempty"`
	Exec     string              `json:"exec,omitempty"`
	Args     []string            `json:"args,omitempty"`
	Env      map[string]any      `json:"env,omitempty"`
	CWD      string              `json:"cwd,omitempty"`
	Terminal string              `json:"terminal,omitempty"`
	Input    []fixtures.PTYInput `json:"input,omitempty"`
	ExitCode *int                `json:"exitCode,omitempty"`
	Stdout   string              `json:"stdout,omitempty"`
	Stderr   string              `json:"stderr,omitempty"`

	GoProfiles map[string]fixtures.GoProfileOutput `json:"goProfiles,omitempty"`
}

func (c execStepConfig) apply(test fixtures.FixtureTest) (fixtures.FixtureTest, error) {
	base := fixtures.ExecFixtureBase{Exec: c.Exec, Args: c.Args, Env: c.Env, CWD: c.CWD, Terminal: c.Terminal, Input: c.Input, GoProfiles: c.GoProfiles}
	if len(c.Input) > 0 && c.Terminal != "pty" {
		return test, fmt.Errorf("input needs terminal: pty")
	}
	if err := fixtures.ValidatePTYInput(c.Input); err != nil {
		return test, err
	}
	switch {
	case c.Script != "":
		if base.Exec == "" {
			base.Exec = "bash"
		}
		base.Args = append([]string{"-c", c.Script}, c.Args...)
	case c.Exec == "":
		return test, fmt.Errorf("declares neither script nor exec")
	}
	test.ExecFixtureBase = base
	test.Expected.ExitCode = c.ExitCode
	test.Expected.Stdout = c.Stdout
	test.Expected.Stderr = c.Stderr
	return test, nil
}

// execOutputs are an exec step's CEL roots and export sources: stdout,
// stderr, exitCode, and json when stdout parses as a JSON document.
func execOutputs(result fixtures.FixtureResult) map[string]any {
	outputs := map[string]any{"stdout": result.Stdout, "stderr": result.Stderr, "exitCode": result.ExitCode}
	trimmed := strings.TrimSpace(result.Stdout)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var parsed any
		if json.Unmarshal([]byte(trimmed), &parsed) == nil {
			outputs["json"] = parsed
		}
	}
	return outputs
}

var execSchema = json.RawMessage(`{
  "type": "object",
  "description": "Run a command. ` + "`script`" + ` runs under bash -c (or the named exec); otherwise exec + args run directly.",
  "additionalProperties": false,
  "anyOf": [{"required": ["script"]}, {"required": ["exec"]}],
  "properties": {
    "script": {"type": "string", "description": "Shell script run with bash -c (or exec -c)."},
    "exec": {"type": "string", "description": "Executable to run."},
    "args": {"type": "array", "items": {"type": "string"}},
    "env": {"type": "object", "additionalProperties": {"type": ["string", "number", "boolean"]}},
    "cwd": {"type": "string"},
    "terminal": {"type": "string", "enum": ["", "pty"]},
    "input": {
      "type": "array",
      "description": "Script a terminal: pty session. Each item waits until expect appears in the output (ANSI stripped, after the previous match), then types send followed by keys.",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["expect"],
        "properties": {
          "expect": {"type": "string"},
          "send": {"type": "string", "description": "Literal text to type."},
          "keys": {"type": "array", "items": {"type": "string", "enum": ["enter", "tab", "space", "esc", "backspace", "up", "down", "right", "left", "ctrl-c", "ctrl-d"]}}
        }
      }
    },
    "exitCode": {"type": "integer", "description": "Expected exit code (default 0)."},
    "stdout": {"type": "string", "description": "Expected stdout, or @file for a golden file."},
    "stderr": {"type": "string", "description": "Expected stderr, or @file for a golden file."},
    "goProfiles": {
      "type": "object",
      "description": "Named Go pprof outputs the command writes, ingested into the run's goprof trace under --profile. The name (cpu, heap, allocs, mutex, block, goroutine) says which profile it is.",
      "additionalProperties": {
        "type": "object",
        "additionalProperties": false,
        "properties": {
          "file": {"type": "string", "description": "Relative pprof filename under the step's artifact directory."},
          "directory": {"type": "string", "description": "Relative directory whose emitted files are Go pprof profiles."},
          "args": {"type": "array", "items": {"type": "string"}, "description": "Arguments appended to the command only with --profile; {{.path}} is the absolute destination."},
          "env": {"type": "string", "description": "Variable set to the absolute destination only with --profile."}
        }
      }
    }
  }
}`)

// runnerKind runs the test or lint engine. A markdown fence carries its YAML
// body; a YAML/JSON step's config is the same option set (JSON is YAML).
type runnerKind struct {
	name string
	run  func(fixtures.FixtureTest, fixtures.RunOptions) fixtures.FixtureResult
}

func (k runnerKind) Name() string { return k.name }

func (k runnerKind) Schema() json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"type": "object", "description": "Options of `+"`gavel %s`"+`, as the YAML keys of a `+"`yaml %s`"+` fence; see gavel fixtures --schema for the full option set."}`, k.name, k.name))
}

func (k runnerKind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	test := step.Test
	if test.RunnerStep == nil {
		test.RunnerStep = &fixtures.RunnerStepSpec{Kind: k.name, Config: string(step.Spec.Config)}
	}
	opts := step.Opts
	opts.Context = ctx
	result := k.run(test, opts)
	outputs := map[string]any{}
	if result.Run != nil {
		outputs = map[string]any{"total": result.Run.Total, "passed": result.Run.Passed, "failed": result.Run.Failed,
			"warned": result.Run.Warned, "skipped": result.Run.Skipped, "lintViolations": result.Run.LintViolations}
	}
	return fixtures.StepOutcome{Result: result, Outputs: outputs}, nil
}
