package verify

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/repomap"
	"github.com/ghodss/yaml"
)

type VerifyConfig struct {
	Model    string             `yaml:"model" json:"model"`
	Prompt   string             `yaml:"prompt" json:"prompt"`
	Sections []string           `yaml:"sections" json:"sections"`
	Weights  map[string]float64 `yaml:"weights" json:"weights"`
}

type GavelConfig struct {
	Verify VerifyConfig `yaml:"verify" json:"verify"`
}

var defaultSections = []string{
	"security",
	"performance",
	"duplication",
	"accuracy",
	"regression",
	"testing",
}

var defaultWeights = map[string]float64{
	"security":    2.0,
	"performance": 1.5,
	"duplication": 1.0,
	"accuracy":    1.0,
	"regression":  1.0,
	"testing":     1.0,
}

func DefaultVerifyConfig() VerifyConfig {
	return VerifyConfig{
		Model:    "claude",
		Sections: defaultSections,
		Weights:  defaultWeights,
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
	if len(override.Sections) > 0 {
		base.Sections = override.Sections
	}
	if len(override.Weights) > 0 {
		for k, v := range override.Weights {
			base.Weights[k] = v
		}
	}
	return base
}
