package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/types"
)

// TestLoadTodosDefaultsResolvesModelFallbacks pins the config→effective-model
// wiring: a .gavel.yaml todos.run with a compact model and a fallbacks list
// resolves into todosDefaults whose Expand()+Candidates() put the primary first
// and include the configured fallback. Guards the fallback-drop bug where only
// the model name flowed end to end and the fallback chain was silently discarded.
func TestLoadTodosDefaultsResolvesModelFallbacks(t *testing.T) {
	saved := todosDef
	t.Cleanup(func() { todosDef = saved })

	dir := t.TempDir()
	// Isolate HOME so a real ~/.gavel.yaml cannot contribute todos config.
	t.Setenv("HOME", t.TempDir())

	const cfg = "todos:\n  run:\n    model: \"opus:high\"\n    fallbacks:\n      - sonnet\n"
	if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write .gavel.yaml: %v", err)
	}

	if err := loadTodosDefaults(dir); err != nil {
		t.Fatalf("loadTodosDefaults: %v", err)
	}
	if todosDef.Spec.Name != "opus:high" {
		t.Errorf("todosDef.Spec.Name = %q, want opus:high", todosDef.Spec.Name)
	}
	if len(todosDef.Spec.Fallbacks) != 1 || todosDef.Spec.Fallbacks[0].Name != "sonnet" {
		t.Fatalf("todosDef.Spec.Fallbacks = %+v, want [sonnet]", todosDef.Spec.Fallbacks)
	}

	eff, err := todosDef.Spec.Model.Expand()
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if eff.Name != "opus" || eff.Effort != api.EffortHigh {
		t.Fatalf("expanded primary = %+v, want opus/high", eff)
	}
	cands := eff.Candidates()
	if len(cands) != 2 || cands[0].Name != "opus" || cands[1].Name != "sonnet" {
		t.Fatalf("Candidates() = %+v, want [opus sonnet]", cands)
	}
}

func TestNewAgentRunConfigPreservesConfiguredSpec(t *testing.T) {
	savedDefaults, savedMode, savedCommit, savedDirty := todosDef, todosRunMode, commitAfter, dirty
	t.Cleanup(func() {
		todosDef, todosRunMode, commitAfter, dirty = savedDefaults, savedMode, savedCommit, savedDirty
	})
	todosRunMode = types.ModeRun
	commitAfter = true
	dirty = false

	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	const cfg = `todos:
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
`
	if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write .gavel.yaml: %v", err)
	}
	if err := loadTodosDefaults(dir); err != nil {
		t.Fatalf("loadTodosDefaults: %v", err)
	}

	runConfig, err := newAgentRunConfig(t.Context(), drivers.Cli, dir, nil, nil)
	if err != nil {
		t.Fatalf("newAgentRunConfig: %v", err)
	}
	spec := runConfig.Spec
	if spec.Prompt.User != "Implement the reviewed change." || spec.Prompt.System != "Keep changes surgical." {
		t.Fatalf("prompt = %+v", spec.Prompt)
	}
	if len(spec.Memory.Skills) != 1 || spec.Memory.Skills[0] != "gavel-todos" {
		t.Fatalf("memory = %+v", spec.Memory)
	}
	if spec.Permissions.Tools.Policies()["Bash"] != api.ToolPolicyAsk || len(spec.Permissions.MCP.Modes) != 1 {
		t.Fatalf("permissions = %+v", spec.Permissions)
	}
	if spec.Setup == nil || spec.Setup.Cwd != filepath.Join(dir, "workspace") || len(spec.Setup.DotEnv) != 1 || spec.Setup.DotEnv[0] != filepath.Join(dir, ".env") {
		t.Fatalf("setup = %+v", spec.Setup)
	}
	if spec.Workflow == nil || spec.Workflow.Verify == nil || spec.Workflow.Verify.MaxIterations != 4 ||
		len(spec.Workflow.Commits) != 1 || spec.Workflow.Commits[0].Gates != api.CommitGatesCheap {
		t.Fatalf("workflow = %+v", spec.Workflow)
	}
	if spec.CLIArgs["fullAuto"] != true {
		t.Fatalf("cliArgs = %+v", spec.CLIArgs)
	}
}
