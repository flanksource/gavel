package verify

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/repomap"
	"github.com/ghodss/yaml"
)

type ChecksConfig struct {
	Disabled           []string `yaml:"disabled" json:"disabled"`
	DisabledCategories []string `yaml:"disabledCategories" json:"disabledCategories"`
}

type VerifyConfig struct {
	Model  string       `yaml:"model" json:"model"`
	Prompt string       `yaml:"prompt" json:"prompt"`
	Checks ChecksConfig `yaml:"checks" json:"checks"`
}

type GavelConfig struct {
	Verify VerifyConfig `yaml:"verify" json:"verify"`
}

func DefaultVerifyConfig() VerifyConfig {
	return VerifyConfig{
		Model: "claude",
	}
}

func LoadConfig(cwd string) (VerifyConfig, error) {
	cfg := DefaultVerifyConfig()

	home, err := os.UserHomeDir()
	if err == nil {
		cfg = mergeFromFile(cfg, filepath.Join(home, ".gavel.yaml"))
	}

	gitRoot := repomap.FindGitRoot(cwd)
	if gitRoot != "" {
		cfg = mergeFromFile(cfg, filepath.Join(gitRoot, ".gavel.yaml"))
	}

	absCwd, _ := filepath.Abs(cwd)
	if absCwd != gitRoot {
		cfg = mergeFromFile(cfg, filepath.Join(absCwd, ".gavel.yaml"))
	}

	return cfg, nil
}

func mergeFromFile(base VerifyConfig, path string) VerifyConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return base
	}
	var gc GavelConfig
	if err := yaml.Unmarshal(data, &gc); err != nil {
		return base
	}
	return MergeVerifyConfig(base, gc.Verify)
}

func MergeVerifyConfig(base, override VerifyConfig) VerifyConfig {
	if override.Model != "" {
		base.Model = override.Model
	}
	if override.Prompt != "" {
		base.Prompt = override.Prompt
	}
	if len(override.Checks.Disabled) > 0 {
		base.Checks.Disabled = append(base.Checks.Disabled, override.Checks.Disabled...)
	}
	if len(override.Checks.DisabledCategories) > 0 {
		base.Checks.DisabledCategories = append(base.Checks.DisabledCategories, override.Checks.DisabledCategories...)
	}
	return base
}
