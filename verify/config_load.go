package verify

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	// Provenance maps each configured leaf (a JSON pointer) to the path of the source that set it.
	Provenance map[string]string `json:"provenance,omitempty" yaml:"provenance,omitempty"`
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

// LoadGlobalGavelConfig merges the built-in defaults with the user's home config only. Settings the
// serve host alone may choose, such as storage, read it, so pushed code can't change them through
// a repository's .gavel.yaml. A missing home config is the defaults.
func LoadGlobalGavelConfig() (GavelConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return GavelConfig{}, fmt.Errorf("find the home directory for %s: %w", GavelConfigFileName, err)
	}
	cfg := DefaultGavelConfig()
	parsed, err := LoadSingleGavelConfig(filepath.Join(home, GavelConfigFileName))
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return GavelConfig{}, err
	}
	return MergeGavelConfig(cfg, parsed), nil
}

// LoadRepositoryGavelConfig is the defaults overlaid with root's own .gavel.yaml and nothing
// else: a hosted repository's settings, which neither the serve host's home config nor any
// parent directory may change.
func LoadRepositoryGavelConfig(root string) (GavelConfig, error) {
	cfg := DefaultGavelConfig()
	parsed, err := LoadSingleGavelConfig(filepath.Join(root, GavelConfigFileName))
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return GavelConfig{}, err
	}
	return MergeGavelConfig(cfg, parsed), nil
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
	}

	layers, err := readGavelConfigLayers(GavelConfigDirs(targetDir))
	if err != nil {
		return GavelConfigTrace{}, err
	}
	for _, layer := range layers {
		if !layer.exists {
			continue
		}
		candidate := GavelConfigSource{Origin: layer.origin, Path: layer.path}
		if candidate.Origin == originTargetDirectory && targetPath != targetDir {
			candidate.Origin = originParentDirectory
		}
		cfg, err := parseGavelConfig(candidate.Path, layer.data)
		if err != nil {
			return GavelConfigTrace{}, err
		}
		candidate.Raw = string(layer.data)
		candidate.Config = cfg
		trace.Sources = append(trace.Sources, candidate)
	}
	trace.Merged, trace.Provenance, err = ComposeGavelConfig(trace.Sources)
	return trace, err
}

// ComposeGavelConfig layers sources over the built-in defaults in order, the
// way captain's ComposeSpecLayers composes spec layers: each source's set
// fields win field by field, and Provenance names, per configured leaf (a JSON
// pointer such as /comments/agents/default/prompt), the source that last set it.
func ComposeGavelConfig(sources []GavelConfigSource) (GavelConfig, map[string]string, error) {
	merged := DefaultGavelConfig()
	provenance := map[string]string{}
	for _, source := range sources {
		encoded, err := json.Marshal(source.Config)
		if err != nil {
			return GavelConfig{}, nil, fmt.Errorf("encode %s: %w", source.Path, err)
		}
		pruned, err := pruneEmptyConfigNodes(encoded)
		if err != nil {
			return GavelConfig{}, nil, fmt.Errorf("encode %s: %w", source.Path, err)
		}
		var tree any
		if err := json.Unmarshal(pruned, &tree); err != nil {
			return GavelConfig{}, nil, fmt.Errorf("encode %s: %w", source.Path, err)
		}
		for _, leaf := range configLeaves(tree, "") {
			provenance[leaf] = source.Path
		}
		merged = MergeGavelConfig(merged, source.Config)
	}
	return merged, provenance, nil
}

// configLeaves lists the JSON pointers of node's scalar and list values.
func configLeaves(node any, pointer string) []string {
	object, ok := node.(map[string]any)
	if !ok {
		return []string{pointer}
	}
	var leaves []string
	for key, child := range object {
		escaped := strings.NewReplacer("~", "~0", "/", "~1").Replace(key)
		leaves = append(leaves, configLeaves(child, pointer+"/"+escaped)...)
	}
	return leaves
}

// IsGavelConfigFile is whether name is a file a config layer reads: the
// layer's .gavel.yaml, or a snippet .gavel.<name>.yaml splitting the layer's
// config across several files.
func IsGavelConfigFile(name string) bool {
	if name == GavelConfigFileName {
		return true
	}
	middle, ok := strings.CutPrefix(name, ".gavel.")
	if !ok {
		return false
	}
	middle, ok = strings.CutSuffix(middle, ".yaml")
	return ok && middle != "" && !strings.ContainsAny(middle, `/\`)
}

// snippetFiles lists the snippets in dir, by name: they apply after the
// directory's .gavel.yaml, in this order. A missing directory has none.
func snippetFiles(dir string) ([]string, error) {
	root, err := os.OpenRoot(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer root.Close() //nolint:errcheck
	listing, err := root.Open(".")
	if err != nil {
		return nil, err
	}
	defer listing.Close() //nolint:errcheck
	entries, err := listing.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", dir, err)
	}
	var names []string
	for _, entry := range entries {
		if entry.Type().IsRegular() && entry.Name() != GavelConfigFileName && IsGavelConfigFile(entry.Name()) {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// LoadGavelConfigDir reads dir's layer: its .gavel.yaml, then its snippets by
// name. A missing directory or file contributes nothing.
func LoadGavelConfigDir(dir string) ([]GavelConfigSource, error) {
	snippets, err := snippetFiles(dir)
	if err != nil {
		return nil, err
	}
	var sources []GavelConfigSource
	for _, name := range append([]string{GavelConfigFileName}, snippets...) {
		path := filepath.Join(dir, name)
		cfg, err := LoadSingleGavelConfig(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		sources = append(sources, GavelConfigSource{Path: path, Config: cfg})
	}
	return sources, nil
}

// LoadSingleGavelConfig reads one .gavel.yaml or snippet file from the given
// absolute path without layering with home/gitRoot/cwd siblings. Returns a zero-value
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
	if !IsGavelConfigFile(filepath.Base(path)) {
		return GavelConfig{}, "", fmt.Errorf("load gavel config: %q is not a %s or .gavel.<name>.yaml file", path, GavelConfigFileName)
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
	if err := gc.Comments.Agents.Validate(); err != nil {
		return GavelConfig{}, fmt.Errorf("%s: %w", path, err)
	}
	if err := gc.PR.Lifecycle.Validate(); err != nil {
		return GavelConfig{}, fmt.Errorf("%s: pr.%w", path, err)
	}
	edit, err := gc.Edit.Normalize()
	if err != nil {
		return GavelConfig{}, fmt.Errorf("%s: %w", path, err)
	}
	gc.Edit = edit
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

// SaveGavelConfig writes cfg to path, a directory's .gavel.yaml or one of its
// .gavel.<name>.yaml snippets.
func SaveGavelConfig(path string, cfg GavelConfig) error {
	if !IsGavelConfigFile(filepath.Base(path)) {
		return fmt.Errorf("save gavel config: %q is not a %s or .gavel.<name>.yaml file", path, GavelConfigFileName)
	}
	// ghodss/yaml encodes via JSON anyway, so the pruned JSON document is written as YAML.
	encoded, err := MarshalGavelConfig(cfg)
	if err != nil {
		return err
	}
	data, err := yaml.JSONToYAML(encoded)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// MarshalGavelConfig is cfg as the JSON document a config file or settings row
// holds: only the configured keys, with every zero-valued struct's `{}` pruned.
func MarshalGavelConfig(cfg GavelConfig) ([]byte, error) {
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return pruneEmptyConfigNodes(encoded)
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
