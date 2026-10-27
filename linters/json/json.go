package jsonlint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
)

type JSON struct {
	linters.RunOptions
	fileCount int
}

func New(workDir string) *JSON           { return &JSON{RunOptions: linters.RunOptions{WorkDir: workDir}} }
func (j *JSON) Name() string             { return "json" }
func (j *JSON) RequiresExecutable() bool { return false }
func (j *JSON) HasDefaultActivation(workDir string) bool {
	return utils.AnyGlobMatchGitIgnored(workDir, j.DefaultIncludes(), nil)
}
func (j *JSON) SetOptions(options linters.RunOptions) { j.RunOptions = options }
func (j *JSON) DefaultIncludes() []string {
	return []string{"**/*.json", "**/*.yaml", "**/*.yml", "**/*.ndjson"}
}
func (j *JSON) DefaultExcludes() []string {
	return []string{"node_modules/**", "vendor/**", ".gavel/**"}
}
func (j *JSON) SupportsJSON() bool { return true }
func (j *JSON) JSONArgs() []string { return nil }
func (j *JSON) SupportsFix() bool  { return false }
func (j *JSON) FixArgs() []string  { return nil }
func (j *JSON) ValidateConfig(config *models.LinterConfig) error {
	if config == nil {
		return fmt.Errorf("json config cannot be nil")
	}
	if len(config.Args) > 0 {
		return fmt.Errorf("json does not accept native arguments; use options.schema")
	}
	return nil
}
func (j *JSON) GetFileCount() int { return j.fileCount }
func (j *JSON) GetRuleCount() int { return 2 }

func (j *JSON) DryRunCommand() (string, []string) {
	args := []string{"lint", "json"}
	if j.Schema != "" {
		args = append(args, "--schema", j.Schema)
	}
	return "gavel", append(args, j.Files...)
}

func (j *JSON) Run(ctx commonsContext.Context, task *clicky.Task) ([]models.Violation, error) {
	j.fileCount = 0
	if len(j.ExtraArgs) > 0 {
		return nil, fmt.Errorf("json does not accept native arguments; use --schema")
	}
	if j.Config != nil {
		if err := j.ValidateConfig(j.Config); err != nil {
			return nil, err
		}
	}
	files, err := j.files(ctx)
	if err != nil {
		return nil, err
	}
	validator := newValidator(ctx)
	if j.Schema != "" {
		if _, err := validator.compile(j.Schema, j.WorkDir); err != nil {
			return nil, err
		}
	}
	var violations []models.Violation
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return violations, err
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return violations, fmt.Errorf("read %s: %w", file, err)
		}
		j.fileCount++
		found, err := validator.lintFile(fileInput{path: file, data: data, schema: j.Schema, workDir: j.WorkDir})
		violations = append(violations, found...)
		if err != nil {
			return violations, fmt.Errorf("%s: %w", file, err)
		}
	}
	return violations, nil
}

func (j *JSON) files(ctx context.Context) ([]string, error) {
	patterns := j.Files
	if len(patterns) == 0 {
		patterns = j.DefaultIncludes()
	}
	seen := map[string]bool{}
	for _, pattern := range patterns {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		path := pattern
		if !filepath.IsAbs(path) {
			path = filepath.Join(j.WorkDir, path)
		}
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			seen[path] = true
			continue
		}
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		root, include := j.WorkDir, []string{pattern}
		if err == nil && info.IsDir() {
			root, include = path, j.DefaultIncludes()
		}
		matches, err := utils.GlobFilesBounded(root, include)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			seen[match] = true
		}
	}
	var files []string
	for file := range seen {
		if !supported(file) {
			continue
		}
		relative, err := filepath.Rel(j.WorkDir, file)
		if err != nil {
			return nil, err
		}
		ignored := false
		for _, pattern := range append(j.DefaultExcludes(), j.Ignores...) {
			matched, err := doublestar.Match(filepath.ToSlash(pattern), filepath.ToSlash(relative))
			if err != nil {
				return nil, fmt.Errorf("invalid ignore %q: %w", pattern, err)
			}
			ignored = ignored || matched
		}
		if !ignored {
			files = append(files, file)
		}
	}
	sort.Strings(files)
	return files, nil
}

func supported(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".yaml", ".yml", ".ndjson":
		return true
	default:
		return false
	}
}
