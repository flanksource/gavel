package repomap

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/git/rules"
	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/commons/logger"
	"github.com/ghodss/yaml"
	"github.com/samber/lo"
	"github.com/samber/oops"
)

//go:embed defaults.yaml
var defaultArchYAML string

func (conf *ArchConf) GetFileMap(path string, commit string) (*FileMap, error) {
	f := FileMap{
		Path:     path,
		Language: detectLanguage(path),
	}

	f.Scopes = conf.Scopes.GetScopesByPath(path)
	f.Tech = conf.Tech.GetTechByPath(path)

	if IsYaml(path) {
		// Check if file exists at commit before attempting to read
		// This prevents spurious errors for deleted files
		if !conf.FileExistsAtCommit(path, commit) {
			return &f, nil
		}

		content, err := conf.ReadFile(path, commit)
		if err != nil {
			logger.Errorf("Error reading %s:%s %w", path, commit, err)
			return &f, nil
		}

		f.KubernetesRefs, err = ExtractKubernetesRefsFromContent(content)
		if err != nil {
			logger.Errorf("Error extracting k8s refs from %s:%s %w", path, commit, err)
		}
	}

	return &f, nil
}

type DirMap struct {
	Path string
	// Default scope for files in this directory
	Scope    ScopeType         `yaml:"scope,omitempty"`
	Language string            `yaml:"language,omitempty"`
	Size     int64             `yaml:"size,omitempty"`
	Tech     []ScopeTechnology `yaml:"tech,omitempty"`
	Children []FileMap         `yaml:"children,omitempty"`
}

type ArchConf struct {
	Git      GitConfig            `yaml:"git,omitempty"`
	Build    BuildConfig          `yaml:"build,omitempty"`
	Golang   GolangConfig         `yaml:"golang,omitempty"`
	Scopes   ScopesConfig         `yaml:"scopes,omitempty"`
	Tech     TechnologyConfig     `yaml:"tech,omitempty"`
	Severity rules.SeverityConfig `yaml:"severity,omitempty"`
	repoPath string               `yaml:"-"` // git repository path, not serialized
}

// IsGitRoot checks if the given path is the root of a git repository

func IsGitRoot(path string) bool {
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		return false
	}
	return true
}

// FindGitRoot walks up from path to find .git directory
func FindGitRoot(path string) string {
	dir := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	}

	for {
		if IsGitRoot(dir) {
			dir, err := filepath.Abs(dir)
			if err != nil {
				panic(err)
			}
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func GetFileMap(path string, commit string) (*FileMap, error) {
	userConf, err := GetConfForFile(path)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to get arch.yaml for %s", path)
	}

	// Load embedded defaults
	defaultConf, err := loadDefaultArchConf()
	if err != nil {
		return nil, oops.Wrapf(err, "failed to load embedded defaults")
	}

	// Merge user config with defaults (user rules checked first)
	conf := defaultConf.Merge(userConf)

	// Set repository path by finding git root
	repoPath := FindGitRoot(path)
	if repoPath == "" {
		return nil, fmt.Errorf("failed to find git repository root for path: %s", path)
	}
	conf.repoPath = repoPath

	f, err := conf.GetFileMap(path, commit)

	if err != nil {
		return nil, oops.Wrapf(err, "failed to get file map for %s", path)
	}
	logger.Tracef("%s => %s", path, f.Pretty().ANSI())
	return f, err
}

// GetConf returns an ArchConf instance with git repository path set
func GetConf(path string) (*ArchConf, error) {
	userConf, err := GetConfForFile(path)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to get arch.yaml for %s", path)
	}

	// Load embedded defaults
	defaultConf, err := loadDefaultArchConf()
	if err != nil {
		return nil, oops.Wrapf(err, "failed to load embedded defaults")
	}

	// Merge user config with defaults (user rules checked first)
	conf := defaultConf.Merge(userConf)

	// Set repository path by finding git root
	repoPath := FindGitRoot(path)
	if repoPath == "" {
		return nil, fmt.Errorf("failed to find git repository root for path: %s", path)
	}
	conf.repoPath = repoPath

	return &conf, nil
}

func GetConfForFile(path string) (*ArchConf, error) {

	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		path = filepath.Dir(path)
	}

	path, _ = filepath.Abs(path)
	file := filepath.Join(path, "arch.yaml")
	if stat, err := os.Stat(file); os.IsNotExist(err) {
		if IsGitRoot(path) {
			return nil, nil
		}
		// Stop at root directory to prevent infinite recursion
		parent := filepath.Dir(path)
		if parent == path {
			return nil, nil
		}
		return GetConfForFile(parent)
	} else if err == nil && !stat.IsDir() {
		return LoadArchConf(file)
	}
	return nil, nil
}

func LoadArchConf(path string) (*ArchConf, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load arch.yaml from %s: %w", path, err)
	}

	var conf ArchConf
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	// Validate scope configuration
	if err := conf.Scopes.Validate(); err != nil {
		return nil, err
	}

	if repoPath := FindGitRoot(path); repoPath != "" {
		conf.repoPath = repoPath
	}

	return &conf, nil
}

// LoadDefaultArchConf loads the embedded defaults.yaml configuration
func LoadDefaultArchConf() (*ArchConf, error) {
	var conf ArchConf
	if err := yaml.Unmarshal([]byte(defaultArchYAML), &conf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedded defaults.yaml: %w", err)
	}
	return &conf, nil
}

// loadDefaultArchConf is kept for backwards compatibility
func loadDefaultArchConf() (*ArchConf, error) {
	return LoadDefaultArchConf()
}

// mergeArchConf merges user config with defaults
// User rules are checked first, defaults used as fallback
func (defaultConf ArchConf) Merge(userConf *ArchConf) ArchConf {
	if userConf == nil {
		return defaultConf
	}

	merged := ArchConf{
		Git:    userConf.Git,
		Build:  userConf.Build,
		Golang: userConf.Golang,
		Scopes: ScopesConfig{
			AllowedScopes: userConf.Scopes.AllowedScopes,
			Rules:         make(PathRules),
		},
		Tech: TechnologyConfig{
			Rules: make(PathRules),
		},
		repoPath: lo.CoalesceOrEmpty(userConf.repoPath, defaultConf.repoPath),
	}

	// Merge scope rules: user rules first, then defaults
	for scope, rules := range userConf.Scopes.Rules {
		merged.Scopes.Rules[scope] = rules
	}
	for scope, rules := range defaultConf.Scopes.Rules {
		if _, exists := merged.Scopes.Rules[scope]; !exists {
			merged.Scopes.Rules[scope] = rules
		} else {
			// Append default rules after user rules for same scope
			merged.Scopes.Rules[scope] = append(merged.Scopes.Rules[scope], rules...)
		}
	}

	// Merge tech rules: user rules first, then defaults
	for tech, rules := range userConf.Tech.Rules {
		merged.Tech.Rules[tech] = rules
	}
	for tech, rules := range defaultConf.Tech.Rules {
		if _, exists := merged.Tech.Rules[tech]; !exists {
			merged.Tech.Rules[tech] = rules
		} else {
			// Append default rules after user rules for same tech
			merged.Tech.Rules[tech] = append(merged.Tech.Rules[tech], rules...)
		}
	}

	return merged
}

// RepoPath returns the git repository path
func (conf *ArchConf) RepoPath() string {
	return conf.repoPath
}

// Exec returns a clicky wrapper for executing git commands in this repository
func (conf *ArchConf) Exec() exec.WrapperFunc {
	return clicky.Exec("git").WithCwd(conf.repoPath).AsWrapper()
}

// FileExistsAtCommit checks if a file exists at a specific commit
func (conf *ArchConf) FileExistsAtCommit(path string, commit string) bool {
	if conf.repoPath == "" || commit == "" {
		return false
	}

	git := conf.Exec()
	_, err := git("cat-file", "-e", fmt.Sprintf("%s:%s", commit, path))
	return err == nil
}

// ReadFile reads file content from disk if commit is empty/HEAD, otherwise from git
func (conf *ArchConf) ReadFile(path string, commit string) (string, error) {
	if conf.repoPath == "" {
		return "", fmt.Errorf("repository path not set")
	}

	if commit == "" {
		return "", fmt.Errorf("must specify a commit to read at")
	}

	// Read from specific commit using git show
	git := conf.Exec()
	result, err := git("show", fmt.Sprintf("%s:%s", commit, path))
	if err != nil {
		return "", fmt.Errorf("failed to read %s at commit %s: %w", path, commit, err)
	}

	return result.Stdout, nil
}
