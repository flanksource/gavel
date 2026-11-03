package git

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

//go:embed git-init-config.prompt
var initConfigPrompt string

type InitConfigOptions struct {
	Path  string `json:"path" flag:"path" help:"Path to git repository" default:"."`
	Model string `json:"model" flag:"model" help:"Captain model selector for recommendations; none skips AI"`
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

	configPath := filepath.Join(root, ".gitanalyze.yaml")

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

	if opts.Model == "none" {
		return configPath, nil
	}

	clicky.Infof("Requesting config recommendations...")
	if err := runAIRecommendation(opts.Model, root, configPath); err != nil {
		return configPath, fmt.Errorf("AI recommendation: %w", err)
	}

	return configPath, nil
}

func runAIRecommendation(model, repoRoot, configPath string) error {
	cfg, err := verify.LoadGavelConfig(repoRoot)
	if err != nil {
		return err
	}
	saved, _, err := captainconfig.Load()
	if err != nil {
		return err
	}
	resolved, err := cfg.Git.InitConfig.Resolve(verify.PromptResolveOptions{
		Base: cfg.AI, DefaultPrompt: initConfigPrompt, Name: "git.initConfig", Dir: repoRoot,
		Data: map[string]any{"configPath": configPath, "workDir": repoRoot}, Saved: &saved.AI, RequireModel: true,
		Request: api.Spec{Model: api.Model{Name: model}},
	})
	if err != nil {
		return err
	}
	prompting.Prepare()
	_, err = promptrun.Run(context.Background(), promptrun.Input{Resolved: resolved, Config: ai.AgentConfig{Model: resolved.Spec.Model, Budget: resolved.Spec.Budget}, Repo: repoRoot})
	return err
}
