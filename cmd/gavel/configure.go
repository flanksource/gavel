package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	clickyapi "github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

// ConfigureOptions drives `gavel configure`, the write counterpart to the
// read-only `gavel config`.
type ConfigureOptions struct {
	Args []string `json:"-" args:"true"`
	// Model is the compact selector written to ai.model — "api:haiku",
	// "agent:claude-sonnet-5", or a bare catalog slug when ~/.captain.yaml
	// already pins a mode for that provider.
	Model string `json:"model,omitempty" flag:"model" help:"Model for every AI operation, as a compact selector (e.g. agent:claude-sonnet-5)"`
	// GroupModel writes commit.grouping.model. Grouping reasons over the whole
	// change set, so it usually wants a more capable tier than message writing.
	GroupModel string `json:"groupModel,omitempty" flag:"group-model" help:"Model for AI commit grouping (commit.grouping.model)"`
	// VerifyModel writes todos.verify.model, the grader for a definition of done.
	VerifyModel string `json:"verifyModel,omitempty" flag:"verify-model" help:"Model for the definition-of-done grader (todos.verify.model)"`
	// Global writes ~/.gavel.yaml instead of the repo's, for defaults that should
	// follow the user across checkouts.
	Global bool `json:"global,omitempty" flag:"global" help:"Write ~/.gavel.yaml instead of the repository's .gavel.yaml"`
}

// ConfigureResult reports what was written where.
type ConfigureResult struct {
	Path    string            `json:"path" yaml:"path"`
	Written map[string]string `json:"written" yaml:"written"`
}

func (r ConfigureResult) Pretty() clickyapi.Text {
	t := clicky.Text("").Append("wrote ", "text-muted").Append(r.Path, "font-mono").NewLine()
	for _, key := range slices.Sorted(maps.Keys(r.Written)) {
		t = t.Append("  "+key+": ", "text-muted").Append(r.Written[key], "font-mono").NewLine()
	}
	return t
}

func init() {
	cmd := clicky.AddNamedCommand("configure", rootCmd, ConfigureOptions{}, runConfigure)
	cmd.Use = "configure [folder]"
	cmd.Short = "Write AI model selections into .gavel.yaml"
	cmd.Long = strings.TrimSpace(`
Write the model selections gavel's AI commands run on.

gavel has no built-in default model: every AI command reads .gavel.yaml, then
~/.captain.yaml, and stops with this command's name if neither pins one. This is
how you pin one.

  gavel configure --model agent:claude-sonnet-5
  gavel configure --model api:haiku --group-model agent:claude-sonnet-5
  gavel configure --global --model agent:claude-sonnet-5

Use 'gavel config' to see the merged result and which file each value came from.`)
	cmd.Args = cobra.MaximumNArgs(1)
}

func runConfigure(opts ConfigureOptions) (any, error) {
	if opts.Model == "" && opts.GroupModel == "" && opts.VerifyModel == "" {
		return nil, fmt.Errorf("nothing to configure: pass at least one of --model, --group-model or --verify-model")
	}

	dir, err := configureTargetDir(opts)
	if err != nil {
		return nil, err
	}

	// Load only the file being edited, not the merged ladder: writing the merged
	// result back would bake every inherited value into this repo and silently
	// pin things the user never chose here.
	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(dir, ".gavel.yaml"))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read %s: %w", filepath.Join(dir, ".gavel.yaml"), err)
	}

	written := map[string]string{}
	assign := func(key, selector string, apply func(api.Model)) error {
		if selector == "" {
			return nil
		}
		// Validate now, against the real catalog, so a typo fails here rather
		// than on the next AI run.
		model, err := (api.Model{Name: selector}).Expand()
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", key, selector, err)
		}
		if _, err := captainai.Resolve(model); err != nil {
			return fmt.Errorf("invalid %s %q: %w", key, selector, err)
		}
		apply(api.Model{Name: selector})
		written[key] = selector
		return nil
	}

	if err := assign("ai.model", opts.Model, func(m api.Model) { cfg.AI.Model = m }); err != nil {
		return nil, err
	}
	if err := assign("commit.grouping.model", opts.GroupModel, func(m api.Model) {
		cfg.Commit.Grouping.Spec.Model = m
	}); err != nil {
		return nil, err
	}
	if err := assign("todos.verify.model", opts.VerifyModel, func(m api.Model) {
		cfg.Todos.Verify.Model = m
	}); err != nil {
		return nil, err
	}

	if err := verify.SaveGavelConfig(dir, cfg); err != nil {
		return nil, fmt.Errorf("write %s: %w", filepath.Join(dir, ".gavel.yaml"), err)
	}
	return ConfigureResult{Path: filepath.Join(dir, ".gavel.yaml"), Written: written}, nil
}

// configureTargetDir picks the file to edit: the home config with --global, the
// positional argument when given, else the repository root — the same place
// LoadGavelConfig reads a repo's config from, so `gavel config` shows the write
// immediately.
func configureTargetDir(opts ConfigureOptions) (string, error) {
	if opts.Global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return home, nil
	}
	if len(opts.Args) > 0 && opts.Args[0] != "" {
		return filepath.Abs(opts.Args[0])
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return "", err
	}
	return gitRepoRoot(workDir)
}
