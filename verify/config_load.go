package verify

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/repomap"
	"github.com/ghodss/yaml"
	yamlv3 "gopkg.in/yaml.v3"
)

type GavelConfigSource struct {
	Origin string      `json:"origin" yaml:"origin"`
	Path   string      `json:"path" yaml:"path"`
	Raw    string      `json:"-" yaml:"-"`
	Config GavelConfig `json:"config" yaml:"config"`
}

type GavelConfigTrace struct {
	TargetPath string              `json:"targetPath" yaml:"targetPath"`
	TargetDir  string              `json:"targetDir" yaml:"targetDir"`
	GitRoot    string              `json:"gitRoot,omitempty" yaml:"gitRoot,omitempty"`
	Sources    []GavelConfigSource `json:"sources,omitempty" yaml:"sources,omitempty"`
	Merged     GavelConfig         `json:"merged" yaml:"merged"`
}

func LoadGavelConfig(cwd string) (GavelConfig, error) {
	cfg := DefaultGavelConfig()

	home, err := os.UserHomeDir()
	if err == nil {
		if cfg, err = mergeFromFile(cfg, filepath.Join(home, ".gavel.yaml")); err != nil {
			return GavelConfig{}, err
		}
	}

	gitRoot := repomap.FindGitRoot(cwd)
	if gitRoot != "" {
		if cfg, err = mergeFromFile(cfg, filepath.Join(gitRoot, ".gavel.yaml")); err != nil {
			return GavelConfig{}, err
		}
	}

	absCwd, _ := filepath.Abs(cwd)
	if absCwd != gitRoot {
		if cfg, err = mergeFromFile(cfg, filepath.Join(absCwd, ".gavel.yaml")); err != nil {
			return GavelConfig{}, err
		}
	}

	return cfg, nil
}

// LoadGavelConfigTrace resolves the effective config for the provided file or
// directory path and records which .gavel.yaml files contributed to the merged
// result. Resolution order matches normal loading: built-in defaults, then the
// user's home config, then the git-root config, then the target directory (or
// the parent directory when the target path is a file).
func LoadGavelConfigTrace(path string) (GavelConfigTrace, error) {
	targetPath, targetDir, err := resolveGavelConfigTarget(path)
	if err != nil {
		return GavelConfigTrace{}, err
	}

	trace := GavelConfigTrace{
		TargetPath: targetPath,
		TargetDir:  targetDir,
		Merged:     DefaultGavelConfig(),
	}

	var candidates []GavelConfigSource
	seen := make(map[string]struct{})
	addCandidate := func(origin, candidatePath string) {
		if candidatePath == "" {
			return
		}
		if _, ok := seen[candidatePath]; ok {
			return
		}
		seen[candidatePath] = struct{}{}
		candidates = append(candidates, GavelConfigSource{
			Origin: origin,
			Path:   candidatePath,
		})
	}

	if home, err := os.UserHomeDir(); err == nil {
		addCandidate("user-home", filepath.Join(home, ".gavel.yaml"))
	}

	trace.GitRoot = repomap.FindGitRoot(targetDir)
	if trace.GitRoot != "" {
		addCandidate("git-root", filepath.Join(trace.GitRoot, ".gavel.yaml"))
	}

	origin := "target-directory"
	if targetPath != targetDir {
		origin = "parent-directory"
	}
	addCandidate(origin, filepath.Join(targetDir, ".gavel.yaml"))

	for _, candidate := range candidates {
		cfg, raw, err := loadSingleGavelConfig(candidate.Path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return GavelConfigTrace{}, err
		}

		candidate.Raw = raw
		candidate.Config = cfg
		trace.Sources = append(trace.Sources, candidate)
		trace.Merged = MergeGavelConfig(trace.Merged, cfg)
	}

	return trace, nil
}

// LoadSingleGavelConfig reads one .gavel.yaml file from the given absolute
// path without layering with home/gitRoot/cwd siblings. Returns a zero-value
// config with os.ErrNotExist when the file is missing so callers can detect
// "need to create" vs. a real read/parse error.
func LoadSingleGavelConfig(path string) (GavelConfig, error) {
	cfg, _, err := loadSingleGavelConfig(path)
	return cfg, err
}

func loadSingleGavelConfig(path string) (GavelConfig, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return GavelConfig{}, "", err
	}
	var doc yamlv3.Node
	if err := yamlv3.Unmarshal(data, &doc); err != nil {
		return GavelConfig{}, "", fmt.Errorf("parse %s: %w", path, err)
	}
	if err := checkRemovedKeys(path, &doc); err != nil {
		return GavelConfig{}, "", err
	}
	var gc GavelConfig
	if err := yaml.Unmarshal(data, &gc); err != nil {
		return GavelConfig{}, "", fmt.Errorf("parse %s: %w", path, err)
	}
	if err := gc.Todos.Validate(); err != nil {
		return GavelConfig{}, "", fmt.Errorf("%s: %w", path, err)
	}
	setPromptSpecBaseDirs(&gc, filepath.Dir(path))
	return gc, string(data), nil
}

func SaveGavelConfig(dir string, cfg GavelConfig) error {
	path := filepath.Join(dir, ".gavel.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// mergeFromFile layers one .gavel.yaml onto base. A missing file is the ordinary
// case — that layer simply does not exist — but every other error is fatal: an
// unreadable or unparseable config, or one carrying an invalid enum, used to be
// discarded here, so a typo anywhere in .gavel.yaml silently ran the whole
// project on built-in defaults instead of the settings it declared.
func mergeFromFile(base GavelConfig, path string) (GavelConfig, error) {
	cfg, err := LoadSingleGavelConfig(path)
	if err != nil {
		if os.IsNotExist(err) {
			return base, nil
		}
		return base, err
	}
	return MergeGavelConfig(base, cfg), nil
}

func resolveGavelConfigTarget(path string) (string, string, error) {
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("resolve absolute path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", "", fmt.Errorf("stat %s: %w", absPath, err)
	}

	if info.IsDir() {
		return absPath, absPath, nil
	}

	return absPath, filepath.Dir(absPath), nil
}
