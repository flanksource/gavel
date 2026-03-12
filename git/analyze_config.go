package git

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/collections"
	"github.com/flanksource/commons/logger"
	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/repomap"
	"github.com/ghodss/yaml"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

//go:embed defaults.gitanalyze.yaml
var defaultGitAnalyzeYAML string

type GitAnalyzeConfig struct {
	FilterSets        map[string]FilterSet `yaml:"filter_sets,omitempty" json:"filter_sets,omitempty"`
	Includes          []string             `yaml:"includes,omitempty" json:"includes,omitempty"`
	IgnoreCommits     []string             `yaml:"ignore_commits,omitempty" json:"ignore_commits,omitempty"`
	IgnoreFiles       []string             `yaml:"ignore_files,omitempty" json:"ignore_files,omitempty"`
	IgnoreAuthors     []string             `yaml:"ignore_authors,omitempty" json:"ignore_authors,omitempty"`
	IgnoreCommitTypes []string             `yaml:"ignore_commit_types,omitempty" json:"ignore_commit_types,omitempty"`
	IgnoreCommitRules []CommitRule         `yaml:"ignore_commit_rules,omitempty" json:"ignore_commit_rules,omitempty"`
	IgnoreResources   []ResourceFilter     `yaml:"ignore_resources,omitempty" json:"ignore_resources,omitempty"`

	// compiled CEL programs, populated by Compile()
	compiledCommitRules []cel.Program `yaml:"-" json:"-"`
}

type FilterSet struct {
	IgnoreCommits     []string         `yaml:"ignore_commits,omitempty" json:"ignore_commits,omitempty"`
	IgnoreFiles       []string         `yaml:"ignore_files,omitempty" json:"ignore_files,omitempty"`
	IgnoreAuthors     []string         `yaml:"ignore_authors,omitempty" json:"ignore_authors,omitempty"`
	IgnoreCommitTypes []string         `yaml:"ignore_commit_types,omitempty" json:"ignore_commit_types,omitempty"`
	IgnoreCommitRules []CommitRule     `yaml:"ignore_commit_rules,omitempty" json:"ignore_commit_rules,omitempty"`
	IgnoreResources   []ResourceFilter `yaml:"ignore_resources,omitempty" json:"ignore_resources,omitempty"`
}

type ResourceFilter struct {
	Kind      string `yaml:"kind,omitempty" json:"kind,omitempty"`
	Name      string `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	CEL       string `yaml:"cel,omitempty" json:"cel,omitempty"`
}

type CommitRule struct {
	CEL string `yaml:"cel" json:"cel"`
}

func GetAnalyzeConfig(path string) (*GitAnalyzeConfig, error) {
	userConf, err := findAnalyzeConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to find .gitanalyze.yaml: %w", err)
	}

	defaultConf, err := loadDefaultAnalyzeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load default analyze config: %w", err)
	}

	merged := defaultConf.Merge(userConf)
	if err := merged.Compile(); err != nil {
		return nil, fmt.Errorf("failed to compile analyze config: %w", err)
	}
	return merged, nil
}

func findAnalyzeConfig(path string) (*GitAnalyzeConfig, error) {
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		path = filepath.Dir(path)
	}
	path, _ = filepath.Abs(path)

	for {
		file := filepath.Join(path, ".gitanalyze.yaml")
		if stat, err := os.Stat(file); err == nil && !stat.IsDir() {
			return loadAnalyzeConfig(file)
		}
		if repomap.IsGitRoot(path) {
			return nil, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			return nil, nil
		}
		path = parent
	}
}

func loadAnalyzeConfig(path string) (*GitAnalyzeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var conf GitAnalyzeConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return &conf, nil
}

func loadDefaultAnalyzeConfig() (*GitAnalyzeConfig, error) {
	var conf GitAnalyzeConfig
	if err := yaml.Unmarshal([]byte(defaultGitAnalyzeYAML), &conf); err != nil {
		return nil, fmt.Errorf("failed to parse embedded defaults.gitanalyze.yaml: %w", err)
	}
	return &conf, nil
}

func (c *GitAnalyzeConfig) Merge(user *GitAnalyzeConfig) *GitAnalyzeConfig {
	if user == nil {
		return c
	}

	merged := &GitAnalyzeConfig{
		FilterSets:        make(map[string]FilterSet),
		Includes:          user.Includes,
		IgnoreCommits:     append(sliceCopy(c.IgnoreCommits), user.IgnoreCommits...),
		IgnoreFiles:       append(sliceCopy(c.IgnoreFiles), user.IgnoreFiles...),
		IgnoreAuthors:     append(sliceCopy(c.IgnoreAuthors), user.IgnoreAuthors...),
		IgnoreCommitTypes: append(sliceCopy(c.IgnoreCommitTypes), user.IgnoreCommitTypes...),
		IgnoreCommitRules: append(sliceCopy(c.IgnoreCommitRules), user.IgnoreCommitRules...),
		IgnoreResources:   append(sliceCopy(c.IgnoreResources), user.IgnoreResources...),
	}

	// If user didn't specify includes, use defaults
	if len(merged.Includes) == 0 {
		merged.Includes = c.Includes
	}

	for name, fs := range c.FilterSets {
		merged.FilterSets[name] = fs
	}
	for name, fs := range user.FilterSets {
		merged.FilterSets[name] = fs
	}

	return merged
}

// Compile pre-compiles all CEL expressions in the config
func (c *GitAnalyzeConfig) Compile() error {
	env, err := newFilterCELEnv()
	if err != nil {
		return fmt.Errorf("failed to create CEL environment: %w", err)
	}

	c.compiledCommitRules = nil
	for _, rule := range c.IgnoreCommitRules {
		prog, err := compileCEL(env, rule.CEL)
		if err != nil {
			return fmt.Errorf("failed to compile commit rule '%s': %w", rule.CEL, err)
		}
		c.compiledCommitRules = append(c.compiledCommitRules, prog)
	}
	return nil
}

// ResolveActiveFilters resolves the active filter sets and returns a merged flat config
func (c *GitAnalyzeConfig) ResolveActiveFilters(include, exclude []string) *GitAnalyzeConfig {
	active := make(map[string]bool)
	for _, name := range c.Includes {
		active[name] = true
	}
	for _, name := range include {
		active[name] = true
	}
	for _, name := range exclude {
		delete(active, name)
	}

	resolved := &GitAnalyzeConfig{
		IgnoreCommits:       sliceCopy(c.IgnoreCommits),
		IgnoreFiles:         sliceCopy(c.IgnoreFiles),
		IgnoreAuthors:       sliceCopy(c.IgnoreAuthors),
		IgnoreCommitTypes:   sliceCopy(c.IgnoreCommitTypes),
		IgnoreCommitRules:   sliceCopy(c.IgnoreCommitRules),
		IgnoreResources:     sliceCopy(c.IgnoreResources),
		compiledCommitRules: c.compiledCommitRules,
	}

	for name, enabled := range active {
		if !enabled {
			continue
		}
		fs, ok := c.FilterSets[name]
		if !ok {
			logger.Warnf("Unknown filter set: %s", name)
			continue
		}
		resolved.IgnoreCommits = append(resolved.IgnoreCommits, fs.IgnoreCommits...)
		resolved.IgnoreFiles = append(resolved.IgnoreFiles, fs.IgnoreFiles...)
		resolved.IgnoreAuthors = append(resolved.IgnoreAuthors, fs.IgnoreAuthors...)
		resolved.IgnoreCommitTypes = append(resolved.IgnoreCommitTypes, fs.IgnoreCommitTypes...)
		resolved.IgnoreCommitRules = append(resolved.IgnoreCommitRules, fs.IgnoreCommitRules...)
		resolved.IgnoreResources = append(resolved.IgnoreResources, fs.IgnoreResources...)
	}

	return resolved
}

func newFilterCELEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("commit", cel.MapType(cel.StringType, cel.AnyType)),
	)
}

func compileCEL(env *cel.Env, expr string) (cel.Program, error) {
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	return env.Program(ast)
}

func sliceCopy[T any](s []T) []T {
	if s == nil {
		return nil
	}
	out := make([]T, len(s))
	copy(out, s)
	return out
}

// matchesAuthor checks if a commit author matches any of the ignore patterns
func matchesAuthor(commit Commit, patterns []string) (bool, string) {
	for _, pattern := range patterns {
		if commit.Author.Matches(pattern) || commit.Committer.Matches(pattern) {
			return true, fmt.Sprintf("author matches '%s'", pattern)
		}
	}
	return false, ""
}

// matchesCommitMessage checks if a commit subject matches any of the ignore patterns
func matchesCommitMessage(subject string, patterns []string) (bool, string) {
	if len(patterns) == 0 {
		return false, ""
	}
	matched, _ := collections.MatchItem(subject, patterns...)
	if matched {
		return true, fmt.Sprintf("commit message matches '%v'", patterns)
	}
	return false, ""
}

// matchesCommitType checks if a commit type matches any of the ignore patterns
func matchesCommitType(commitType CommitType, patterns []string) (bool, string) {
	if len(patterns) == 0 {
		return false, ""
	}
	ct := string(commitType)
	matched, _ := collections.MatchItem(ct, patterns...)
	if matched {
		return true, fmt.Sprintf("commit type '%s' matches '%v'", ct, patterns)
	}
	return false, ""
}

// matchesFile checks if a file path matches any of the ignore patterns
func matchesFile(file string, patterns []string) (bool, string) {
	if len(patterns) == 0 {
		return false, ""
	}
	if matched, _ := collections.MatchItem(file, patterns...); matched {
		return true, fmt.Sprintf("file '%s' matches patterns", file)
	}
	if matched, _ := collections.MatchItem(filepath.Base(file), patterns...); matched {
		return true, fmt.Sprintf("file '%s' matches patterns (basename)", file)
	}
	return false, ""
}

// evalCommitCELRules evaluates CEL rules against a commit and returns true if any match
func evalCommitCELRules(rules []CommitRule, programs []cel.Program, commit Commit, changes []CommitChange) (bool, string) {
	if len(programs) == 0 {
		return false, ""
	}

	ctx := buildFilterCommitContext(commit, changes)

	for i, prog := range programs {
		result, _, err := prog.Eval(ctx)
		if err != nil {
			continue
		}
		if boolVal, ok := result.(types.Bool); ok && boolVal == types.True {
			return true, fmt.Sprintf("CEL rule '%s'", rules[i].CEL)
		}
	}
	return false, ""
}

func buildFilterCommitContext(commit Commit, changes []CommitChange) map[string]any {
	files := make([]string, 0, len(changes))
	totalAdds, totalDels := 0, 0
	for _, c := range changes {
		files = append(files, c.File)
		totalAdds += c.Adds
		totalDels += c.Dels
	}

	isMerge := strings.HasPrefix(commit.Subject, "Merge ")

	return map[string]any{
		"commit": map[string]any{
			"hash":          commit.Hash,
			"author":        commit.Author.Name,
			"author_email":  commit.Author.Email,
			"subject":       commit.Subject,
			"body":          commit.Body,
			"type":          string(commit.CommitType),
			"scope":         string(commit.Scope),
			"is_merge":      isMerge,
			"files_changed": len(changes),
			"line_changes":  totalAdds + totalDels,
			"additions":     totalAdds,
			"deletions":     totalDels,
			"files":         files,
			"tags":          commit.Tags,
			"is_tagged":     len(commit.Tags) > 0,
		},
	}
}
