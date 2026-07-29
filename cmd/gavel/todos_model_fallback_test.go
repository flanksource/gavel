package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
)

// isolatedTodosRun writes cfg as the workspace's .gavel.yaml and neutralises
// every other input the resolution seam reads: HOME, because LoadGavelConfig
// merges ~/.gavel.yaml and would otherwise make whoever runs the suite a hidden
// layer, and the run flags, which are the highest layer — a value another test
// left behind would outrank the config under test.
func isolatedTodosRun(t *testing.T, cfg string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	savedModel, savedEffort, savedDriver := todoModel, todoEffort, todosDriver
	savedBudget, savedTurns := maxBudget, maxTurns
	savedMode, savedCommit, savedDirty, savedResume := todosRunMode, commitAfter, dirty, resumeSession
	t.Cleanup(func() {
		todoModel, todoEffort, todosDriver = savedModel, savedEffort, savedDriver
		maxBudget, maxTurns = savedBudget, savedTurns
		todosRunMode, commitAfter, dirty, resumeSession = savedMode, savedCommit, savedDirty, savedResume
	})
	todoModel, todoEffort, todosDriver = "", "", ""
	maxBudget, maxTurns = 0, 0
	todosRunMode, commitAfter, dirty, resumeSession = types.ModeRun, true, false, false

	dir := t.TempDir()
	if strings.TrimSpace(cfg) != "" {
		if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(cfg), 0o644); err != nil {
			t.Fatalf("write .gavel.yaml: %v", err)
		}
	}
	return dir
}

// TestTodosSpecResolvesModelFallbacks pins the config→effective-model wiring: a
// .gavel.yaml todos.run with a compact model and a fallbacks list resolves into
// an expanded primary whose Candidates() put it first and include the configured
// fallback. Guards the fallback-drop bug where only the model name flowed end to
// end and the fallback chain was silently discarded.
func TestTodosSpecResolvesModelFallbacks(t *testing.T) {
	dir := isolatedTodosRun(t, "todos:\n  run:\n    model: \"opus:high\"\n    fallbacks:\n      - sonnet\n")

	resolved, err := todosSpec(dir, nil)
	if err != nil {
		t.Fatalf("todosSpec: %v", err)
	}
	if resolved.Spec.Name != "opus" || resolved.Spec.Effort != api.EffortHigh {
		t.Fatalf("resolved model = %+v, want opus/high", resolved.Spec.Model)
	}
	cands := resolved.Spec.Model.Candidates()
	if len(cands) != 2 || cands[0].Name != "opus" || cands[1].Name != "sonnet" {
		t.Fatalf("Candidates() = %+v, want [opus sonnet]", cands)
	}
}

// TestTodosSpecMergesAIBaseIntoTodosRun proves the documented `ai:` base reaches
// todo runs — previously it reached every AI operation except the single most
// expensive one — and that todos.run overrides it field-wise rather than
// replacing the block.
//
// The model deliberately does NOT come from `ai:` here: todos-run.prompt declares
// one, and a mode's .prompt frontmatter outranks the base in every gavel AI
// operation (verify.PromptSpec.Resolve is base.Merge(defaultSpec).Merge(opSpec)).
// The base supplies what the frontmatter leaves unsaid.
func TestTodosSpecMergesAIBaseIntoTodosRun(t *testing.T) {
	dir := isolatedTodosRun(t, `ai:
  budget:
    cost: 4.5
    maxTurns: 3
  permissions:
    mode: acceptEdits
todos:
  run:
    budget:
      maxTurns: 12
`)

	resolved, err := todosSpec(dir, nil)
	if err != nil {
		t.Fatalf("todosSpec: %v", err)
	}
	if resolved.Spec.Budget.Cost != 4.5 {
		t.Errorf("budget.cost = %v, want the ai: base 4.5", resolved.Spec.Budget.Cost)
	}
	if resolved.Spec.Permissions.Mode != api.PermissionAcceptEdits {
		t.Errorf("permissions.mode = %q, want the ai: base acceptEdits", resolved.Spec.Permissions.Mode)
	}
	if resolved.Spec.Budget.MaxTurns != 12 {
		t.Errorf("budget.maxTurns = %d, want the todos.run override 12", resolved.Spec.Budget.MaxTurns)
	}
	if got := resolved.Provenance["budget.cost"]; got != ".gavel.yaml ai" {
		t.Errorf("budget.cost provenance = %q, want the ai layer", got)
	}
}

// TestTodosTimeoutReachesEveryMode pins .gavel.yaml todos.timeout as the one
// place a project caps a todo run, whichever entrypoint starts it. `todos plan
// revise` (which goes through newExecutor, not runTodosRun) and `todos check`
// both used to miss it entirely: revise ran with the zero-value defaults and
// check was pinned to its own flag, so the same todo got three different
// ceilings depending on how it was invoked.
func TestTodosTimeoutReachesEveryMode(t *testing.T) {
	dir := isolatedTodosRun(t, "todos:\n  timeout: 45m\n")

	for _, mode := range []types.RunMode{types.ModeRun, types.ModePlan, types.ModeVerify} {
		t.Run(string(mode), func(t *testing.T) {
			// timeout 0 is the unset flag: `todos check --timeout` and the run
			// flags all default to zero so the config can be seen.
			resolved, err := todosSpecForMode(dir, mode, nil, 0)
			if err != nil {
				t.Fatalf("todosSpecForMode: %v", err)
			}
			if resolved.Timeout != 45*time.Minute {
				t.Fatalf("timeout = %s, want the configured 45m", resolved.Timeout)
			}
		})
	}

	// The flag is the request layer, so an explicitly passed --timeout still wins.
	resolved, err := todosSpecForMode(dir, types.ModeVerify, nil, 90*time.Second)
	if err != nil {
		t.Fatalf("todosSpecForMode with flag: %v", err)
	}
	if resolved.Timeout != 90*time.Second {
		t.Fatalf("timeout = %s, want the explicit flag 90s", resolved.Timeout)
	}
}

func TestNewAgentRunConfigPreservesConfiguredSpec(t *testing.T) {
	dir := isolatedTodosRun(t, `todos:
  run:
    model: claude-sonnet-5
    prompt:
      user: Implement the reviewed change.
      system: Keep changes surgical.
    memory:
      skills: [gavel-todos]
    permissions:
      mode: acceptEdits
      tools:
        Bash: ask
      mcp:
        postgres: enabled
    setup:
      cwd: workspace
      dotenv: [.env]
    workflow:
      verify:
        commands: [go test ./todos]
        scope: changed
        maxIterations: 4
      commits:
        - on: run
          gates: cheap
    cliArgs:
      fullAuto: true
`)

	runConfig, _, err := newAgentRunConfig(t.Context(), dir, nil, nil)
	if err != nil {
		t.Fatalf("newAgentRunConfig: %v", err)
	}
	spec := runConfig.Spec
	// An inline `prompt.user` is the run's TEMPLATE, not a raw body override: it
	// reaches the agent once, rendered by todos/prompt. Carrying it on the spec as
	// well would inject the unrendered source alongside the rendered prompt.
	if runConfig.Template != "Implement the reviewed change." {
		t.Fatalf("template = %q, want the configured inline prompt", runConfig.Template)
	}
	if spec.Prompt.User != "" {
		t.Fatalf("spec prompt body = %q, want it carried only as the template", spec.Prompt.User)
	}
	if spec.Prompt.System != "Keep changes surgical." {
		t.Fatalf("prompt system = %q", spec.Prompt.System)
	}
	if len(spec.Memory.Skills) != 1 || spec.Memory.Skills[0] != "gavel-todos" {
		t.Fatalf("memory = %+v", spec.Memory)
	}
	if spec.Permissions.Tools.Policies()["Bash"] != api.ToolPolicyAsk || len(spec.Permissions.MCP.Modes) != 1 {
		t.Fatalf("permissions = %+v", spec.Permissions)
	}
	// Setup paths stay exactly as configured: anchoring them to the group's
	// working directory is the setup hook's job, performed once inside the run.
	// Resolving here too would pin them to the invocation directory.
	if spec.Setup == nil || spec.Setup.Cwd != "workspace" || len(spec.Setup.DotEnv) != 1 || spec.Setup.DotEnv[0] != ".env" {
		t.Fatalf("setup = %+v", spec.Setup)
	}
	if spec.Workflow == nil || spec.Workflow.Verify == nil || spec.Workflow.Verify.MaxIterations != 4 ||
		len(spec.Workflow.Commits) != 1 || spec.Workflow.Commits[0].Gates != api.CommitGatesCheap {
		t.Fatalf("workflow = %+v", spec.Workflow)
	}
	if spec.CLIArgs["fullAuto"] != true {
		t.Fatalf("cliArgs = %+v", spec.CLIArgs)
	}
	// WorkDir stays the discovery root: the executor joins the todo's own CWD
	// onto it. Pre-joining here would make that a double join.
	if runConfig.WorkDir != dir {
		t.Fatalf("workDir = %q, want %q", runConfig.WorkDir, dir)
	}
}
