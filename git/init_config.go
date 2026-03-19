package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/repomap"
)

type InitConfigOptions struct {
	Path  string `json:"path" flag:"path" help:"Path to git repository" default:"."`
	Model string `json:"model" flag:"model" help:"AI CLI to use for recommendations: claude, gemini, codex (or skip AI with --model=none)" default:"claude"`
	Debug bool   `json:"debug" flag:"debug" help:"Enable debug logging"`
}

func (o InitConfigOptions) GetName() string { return "init-config" }

func InitConfig(opts InitConfigOptions) (string, error) {
	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	root := repomap.FindGitRoot(absPath)
	if root == "" {
		return "", fmt.Errorf("not a git repository: %s", absPath)
	}

	configPath := filepath.Join(root, "repomap.yaml")

	if _, err := os.Stat(configPath); err == nil {
		logger.Infof("Found existing %s", configPath)
	} else {
		defaultCfg := `extends:
  - preset:bots
  - preset:noise
  - preset:merges

exclude:
  files: []
  authors: []
`
		if err := os.WriteFile(configPath, []byte(defaultCfg), 0o644); err != nil {
			return "", fmt.Errorf("failed to write %s: %w", configPath, err)
		}
		clicky.Infof("Created %s with default rules", configPath)
	}

	if opts.Model == "none" || opts.Model == "" {
		return configPath, nil
	}

	clicky.Infof("Asking %s to recommend config updates...", opts.Model)
	if err := runAIRecommendation(opts.Model, root, configPath, opts.Debug); err != nil {
		logger.Warnf("AI recommendation failed: %v", err)
		clicky.Infof("You can edit %s manually", configPath)
		return configPath, nil
	}

	return configPath, nil
}

func runAIRecommendation(model, repoRoot, configPath string, debug bool) error {
	prompt := buildInitPrompt(configPath)

	name, args := buildAIArgs(model, prompt, repoRoot, debug)

	logger.V(1).Infof("exec: %s %s", name, strings.Join(args, " "))

	proc := clicky.Exec(name, args...).
		WithCwd(repoRoot).
		WithTimeout(3*time.Minute).
		Stream(os.Stderr, os.Stderr)

	if debug {
		proc = proc.Debug()
	}

	result := proc.Run().Result()

	if result.Error != nil || result.ExitCode != 0 {
		msg := result.Stderr
		if msg == "" {
			msg = result.Stdout
		}
		return fmt.Errorf("%s failed (exit %d): %s", name, result.ExitCode, msg)
	}

	return nil
}

func buildAIArgs(model, prompt, _ string, debug bool) (string, []string) {
	name, modelFlag := resolveAICLI(model)

	switch name {
	case "claude":
		args := []string{"-p", prompt, "--allowedTools", "Read,Glob,Grep,Edit,Write,Bash"}
		if modelFlag != "" {
			args = append(args, "--model", modelFlag)
		}
		if debug {
			args = append(args, "--verbose")
		}
		return name, args

	case "gemini":
		args := []string{"-p", prompt}
		if modelFlag != "" {
			args = append(args, "-m", modelFlag)
		}
		return name, args

	case "codex":
		args := []string{"-q", prompt, "--full-auto"}
		if modelFlag != "" {
			args = append(args, "-m", modelFlag)
		}
		return name, args

	default:
		return "claude", []string{"-p", prompt, "--allowedTools", "Read,Glob,Grep,Edit,Write,Bash"}
	}
}

func resolveAICLI(model string) (string, string) {
	known := map[string]bool{"claude": true, "gemini": true, "codex": true}
	if known[model] {
		return model, ""
	}
	prefix := model
	if idx := strings.IndexByte(model, '-'); idx > 0 {
		prefix = model[:idx]
	}
	if known[prefix] {
		return prefix, model
	}
	return "claude", model
}

func buildInitPrompt(configPath string) string {
	return fmt.Sprintf(`You are configuring .gitanalyze.yaml for a git repository.

Read the existing config at %s, then analyze this repository to recommend updates:

1. Look at the git log (last 50 commits) to identify:
   - Bot authors that should be filtered (CI bots, dependabot variants, etc.)
   - Common noise commit patterns (fixup!, squash!, auto-generated, etc.)
   - Commit types that produce noise for this specific repo

2. Look at the repository structure to identify:
   - Generated files that should be ignored (vendor/, node_modules/, dist/, build/, etc.)
   - Lock files specific to this project's package managers
   - Binary or asset files (images, fonts, compiled output)
   - IDE and editor config files (.vscode/, .idea/, etc.)

3. If this repo contains Kubernetes manifests, identify:
   - Resource types that are auto-generated or low-signal
   - Namespaces that are noise (kube-system, etc.)

Edit %s directly with your recommendations. Keep the existing filter_sets structure.
Add new filter sets if appropriate. Update the includes list.
Add YAML comments explaining each addition.

Do NOT remove existing defaults - only add to them.
Do NOT add entries that duplicate what's already there.
Keep the file concise and well-organized.`, configPath, configPath)
}
