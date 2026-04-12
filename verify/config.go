package verify

import (
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/repomap"
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

type LintIgnoreRule struct {
	Rule   string `yaml:"rule,omitempty" json:"rule,omitempty"`
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	File   string `yaml:"file,omitempty" json:"file,omitempty"`
}

func (r LintIgnoreRule) MatchesViolation(v models.Violation) bool {
	if r.Source != "" && r.Source != v.Source {
		return false
	}
	if r.Rule != "" {
		if v.Rule == nil || v.Rule.Method != r.Rule {
			return false
		}
	}
	if r.File != "" {
		matched, _ := doublestar.Match(r.File, v.File)
		if !matched {
			return false
		}
	}
	return r.Rule != "" || r.Source != ""
}

type LintConfig struct {
	Ignore []LintIgnoreRule `yaml:"ignore,omitempty" json:"ignore,omitempty"`
}

type GavelConfig struct {
	Verify VerifyConfig `yaml:"verify" json:"verify"`
	Lint   LintConfig   `yaml:"lint,omitempty" json:"lint,omitempty"`
}

func DefaultVerifyConfig() VerifyConfig {
	return VerifyConfig{
		Model: "claude",
	}
}

func LoadConfig(cwd string) (VerifyConfig, error) {
	gc, err := LoadGavelConfig(cwd)
	return gc.Verify, err
}

func LoadGavelConfig(cwd string) (GavelConfig, error) {
	cfg := GavelConfig{Verify: DefaultVerifyConfig()}

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

func SaveGavelConfig(dir string, cfg GavelConfig) error {
	path := filepath.Join(dir, ".gavel.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func mergeFromFile(base GavelConfig, path string) GavelConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return base
	}
	var gc GavelConfig
	if err := yaml.Unmarshal(data, &gc); err != nil {
		return base
	}
	base.Verify = MergeVerifyConfig(base.Verify, gc.Verify)
	base.Lint = MergeLintConfig(base.Lint, gc.Lint)
	return base
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

func MergeLintConfig(base, override LintConfig) LintConfig {
	if len(override.Ignore) > 0 {
		base.Ignore = append(base.Ignore, override.Ignore...)
	}
	return base
}
