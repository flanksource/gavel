package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/repomap"
	"github.com/ghodss/yaml"
)

// GavelConfigFileName is the only file name the config loaders read. Every
// layer — home, git root, target directory — is this name inside a directory.
const GavelConfigFileName = ".gavel.yaml"

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

// LoadGavelConfig merges the built-in defaults with the user's home config, the
// git root's, then cwd's. A missing file is an absent layer; any other read,
// parse or validation error fails the load — discarding one used to run a
// project with a typo in its .gavel.yaml on built-in defaults. Results are cached per directory
// against the exact bytes of every consulted file (see gavelConfigCache), and
// each call returns its own deep copy, so callers may mutate what they get.
func LoadGavelConfig(cwd string) (GavelConfig, error) {
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return GavelConfig{}, fmt.Errorf("resolve config directory %q: %w", cwd, err)
	}
	layers, err := readGavelConfigLayers(GavelConfigDirs(absCwd))
	if err != nil {
		return GavelConfig{}, err
	}
	if cached, ok := cachedGavelConfig(absCwd, layers); ok {
		return cloneGavelConfig(cached)
	}
	configs, err := loadConfigLayers(absCwd, layers)
	if err != nil {
		return GavelConfig{}, err
	}
	cfg := DefaultGavelConfig()
	for _, parsed := range configs {
		cfg = MergeGavelConfig(cfg, parsed)
	}
	storeGavelConfig(absCwd, layers, cfg)
	return cloneGavelConfig(cfg)
}

const (
	originUserHome        = "user-home"
	originGitRootParent   = "git-root-parent"
	originGitRoot         = "git-root"
	originTargetDirectory = "target-directory"
	originParentDirectory = "parent-directory"
)

// GavelConfigDir is one directory whose .gavel.yaml joins the config chain,
// labelled with why it is consulted.
type GavelConfigDir struct {
	Origin string
	Dir    string
}

// GavelConfigDirs is the single source of truth for the .gavel.yaml chain of
// the absolute directory absTarget, lowest precedence first: $HOME, the git
// root's immediate parent (so one file can cover every repo checked out side
// by side), the git root, then absTarget. Only the one parent level is read,
// never further ancestors, and it is skipped outside a git repository or when
// it is "/". A directory reached twice keeps its first, lower-precedence
// origin. $HOME and the git root are resolved on every call, so a changed
// $HOME or a newly created .git moves the chain — and with it the cache
// fingerprint.
func GavelConfigDirs(absTarget string) []GavelConfigDir {
	var dirs []GavelConfigDir
	add := func(origin, dir string) {
		for _, existing := range dirs {
			if existing.Dir == dir {
				return
			}
		}
		dirs = append(dirs, GavelConfigDir{Origin: origin, Dir: dir})
	}
	if home, err := os.UserHomeDir(); err == nil {
		add(originUserHome, home)
	}
	if gitRoot := repomap.FindGitRoot(absTarget); gitRoot != "" {
		if parent := filepath.Dir(gitRoot); filepath.Dir(parent) != parent {
			add(originGitRootParent, parent)
		}
		add(originGitRoot, gitRoot)
		var nested []string
		for dir := filepath.Dir(absTarget); absTarget != gitRoot && dir != gitRoot && dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			nested = append(nested, dir)
		}
		for i := len(nested) - 1; i >= 0; i-- {
			add(originParentDirectory, nested[i])
		}
	}
	add(originTargetDirectory, absTarget)
	return dirs
}

// LoadGavelConfigTrace resolves the effective config for the provided file or
// directory path and records which .gavel.yaml files contributed to the merged
// result. Resolution order matches normal loading (GavelConfigDirs): built-in
// defaults, then the user's home config, the git root's parent, the git root,
// then the target directory (or the parent directory when the target path is
// a file).
func LoadGavelConfigTrace(path string) (GavelConfigTrace, error) {
	targetPath, targetDir, err := resolveGavelConfigTarget(path)
	if err != nil {
		return GavelConfigTrace{}, err
	}

	trace := GavelConfigTrace{
		TargetPath: targetPath,
		TargetDir:  targetDir,
		GitRoot:    repomap.FindGitRoot(targetDir),
		Merged:     DefaultGavelConfig(),
	}

	for _, dir := range GavelConfigDirs(targetDir) {
		candidate := GavelConfigSource{Origin: dir.Origin, Path: filepath.Join(dir.Dir, GavelConfigFileName)}
		if candidate.Origin == originTargetDirectory && targetPath != targetDir {
			candidate.Origin = originParentDirectory
		}
		data, err := readFileInDir(dir.Dir, GavelConfigFileName)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return GavelConfigTrace{}, err
		}

		cfg, err := parseGavelConfig(candidate.Path, data)
		if err != nil {
			return GavelConfigTrace{}, err
		}
		candidate.Raw = string(data)
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
	// Callers name a directory and this loader reads that directory's config.
	// Pinning the file name keeps a caller-supplied path from selecting some
	// other file to parse and echo back (the settings API resolves its
	// directory from a request parameter), and a path that names anything else
	// is a caller bug worth reporting rather than a missing config.
	if filepath.Base(path) != GavelConfigFileName {
		return GavelConfig{}, "", fmt.Errorf("load gavel config: %q is not a %s file", path, GavelConfigFileName)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return GavelConfig{}, "", err
	}
	gc, err := parseGavelConfig(path, data)
	if err != nil {
		return GavelConfig{}, "", err
	}
	return gc, string(data), nil
}

// parseGavelConfig decodes and validates the bytes of the .gavel.yaml at path,
// stamping prompt base directories with path's directory.
func parseGavelConfig(path string, data []byte) (GavelConfig, error) {
	gavelConfigParses.Add(1)
	document, err := decodeYAMLDocument(data)
	if err != nil {
		return GavelConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	var value any
	if err := document.Decode(&value); err != nil {
		return GavelConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	encoded, err := yaml.YAMLToJSON(data)
	if err != nil {
		return GavelConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return decodeGavelConfig(path, encoded)
}

func decodeGavelConfig(path string, encoded []byte) (GavelConfig, error) {
	var gc GavelConfig
	if err := DecodeJSON(encoded, &gc, DecodeOptions{Source: path, Renames: legacyFieldHints}); err != nil {
		return GavelConfig{}, err
	}
	if err := gc.Todos.Validate(); err != nil {
		return GavelConfig{}, fmt.Errorf("%s: %w", path, err)
	}
	if err := gc.PR.Lifecycle.Validate(); err != nil {
		return GavelConfig{}, fmt.Errorf("%s: pr.%w", path, err)
	}
	if err := gc.validateRuntime(); err != nil {
		return GavelConfig{}, fmt.Errorf("%s: %w", path, err)
	}
	if err := validateConfigSpecLayers(gc, path); err != nil {
		return GavelConfig{}, err
	}
	setPromptSpecBaseDirs(&gc, filepath.Dir(path))
	gc.resolveOptionPaths(filepath.Dir(path))
	for i := range gc.Lint.Ignore {
		if gc.Lint.Ignore[i].File != "" {
			gc.Lint.Ignore[i].baseDir = filepath.Dir(path)
		}
	}
	return gc, nil
}

func SaveGavelConfig(dir string, cfg GavelConfig) error {
	path := filepath.Join(dir, ".gavel.yaml")
	// Marshal through JSON rather than yaml.Marshal so the document can be pruned
	// before it is written: ghodss/yaml encodes via JSON anyway, and writing every
	// zero-valued struct as `{}` is what fills these files with dead keys.
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	pruned, err := pruneEmptyConfigNodes(encoded)
	if err != nil {
		return err
	}
	data, err := yaml.JSONToYAML(pruned)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
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
