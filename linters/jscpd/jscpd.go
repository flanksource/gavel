package jscpd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

type JSCPD struct {
	linters.RunOptions
}

func NewJSCPD(workDir string) *JSCPD {
	return &JSCPD{
		RunOptions: linters.RunOptions{WorkDir: workDir},
	}
}

func (j *JSCPD) SetOptions(opts linters.RunOptions) {
	j.RunOptions = opts
}

func (j *JSCPD) Name() string { return "jscpd" }

func (j *JSCPD) DefaultIncludes() []string {
	return []string{"**/*"}
}

func (j *JSCPD) DefaultExcludes() []string {
	return []string{
		"**/*_test.go",
		"**/*_test.py",
		"**/*.test.*",
		"**/*.spec.*",
		"**/*.generated.*",
		"**/*.pb.go",
		"**/*_gen.go",
		"**/mock_*",
		"**/*_mock.*",
		"**/testdata/**",
		"**/*.min.js",
		"**/*.min.css",
		"**/package-lock.json",
		"**/yarn.lock",
		"**/go.sum",
		"**/Cargo.lock",
	}
}

func (j *JSCPD) GetSupportedLanguages() []string {
	return []string{"go", "javascript", "typescript", "python", "java", "ruby", "rust", "css", "html"}
}

func (j *JSCPD) GetEffectiveExcludes(language string, config *models.Config) []string {
	if config == nil {
		return j.DefaultExcludes()
	}
	return config.GetAllLanguageExcludes(language, j.DefaultExcludes())
}

func (j *JSCPD) GetEffectiveIncludes(language string, config *models.Config) []string {
	if config == nil {
		return j.DefaultIncludes()
	}
	return config.GetAllLanguageIncludes(language, j.DefaultIncludes())
}

func (j *JSCPD) SupportsJSON() bool        { return true }
func (j *JSCPD) JSONArgs() []string         { return []string{"--reporters", "json"} }
func (j *JSCPD) SupportsFix() bool          { return false }
func (j *JSCPD) FixArgs() []string          { return nil }

func (j *JSCPD) ValidateConfig(config *models.LinterConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	return nil
}

func (j *JSCPD) Run(ctx commonsContext.Context, task *clicky.Task) ([]models.Violation, error) {
	tempDir, err := os.MkdirTemp("", "jscpd-report-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	args := []string{"--reporters", "json", "--output", tempDir, "--gitignore"}

	excludes := j.buildExcludes()
	for _, pattern := range excludes {
		args = append(args, "--ignore", pattern)
	}

	if j.Config != nil {
		args = append(args, j.Config.Args...)
	}
	args = append(args, j.ExtraArgs...)

	if len(j.Files) > 0 {
		args = append(args, j.Files...)
	} else {
		args = append(args, ".")
	}

	cmd := exec.CommandContext(ctx, "jscpd", args...)
	cmd.Dir = j.WorkDir

	logger.Infof("Executing: jscpd %s", strings.Join(args, " "))

	output, err := cmd.CombinedOutput()

	reportPath := filepath.Join(tempDir, "jscpd-report.json")
	reportExists := fileExists(reportPath)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && reportExists {
			logger.Debugf("jscpd exit code %d with report file - treating as success with violations", exitErr.ExitCode())
			err = nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("jscpd execution failed: %w\nOutput:\n%s", err, string(output))
	}

	if !reportExists {
		return []models.Violation{}, nil
	}

	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read jscpd report: %w", err)
	}

	return j.parseViolations(reportData, excludes)
}

func (j *JSCPD) buildExcludes() []string {
	var excludes []string
	if j.ArchConfig != nil {
		excludes = j.ArchConfig.GetAllLanguageExcludes("", j.DefaultExcludes())
	} else {
		excludes = append(models.GetBuiltinExcludePatterns(), j.DefaultExcludes()...)
	}
	return append(excludes, j.Ignores...)
}

func (j *JSCPD) parseViolations(data []byte, excludes []string) ([]models.Violation, error) {
	var report JscpdReport
	if err := json.Unmarshal(data, &report); err != nil {
		logger.Debugf("Failed to parse jscpd JSON output: %v\nOutput: %s", err, string(data))
		return nil, fmt.Errorf("failed to parse jscpd JSON output: %w", err)
	}

	var violations []models.Violation
	for _, dup := range report.Duplicates {
		if matchesAny(dup.FirstFile.Name, excludes) || matchesAny(dup.SecondFile.Name, excludes) {
			continue
		}

		firstFile := normalizePath(j.WorkDir, dup.FirstFile.Name)
		secondFile := normalizePath(j.WorkDir, dup.SecondFile.Name)

		var msg string
		if firstFile == secondFile {
			msg = fmt.Sprintf("Duplicate code (%d lines, %s) lines %d-%d, also at lines %d-%d",
				dup.Lines, dup.Format,
				dup.FirstFile.StartLoc.Line, dup.FirstFile.EndLoc.Line,
				dup.SecondFile.StartLoc.Line, dup.SecondFile.EndLoc.Line)
		} else {
			msg = fmt.Sprintf("Duplicate code (%d lines, %s) lines %d-%d, also in %s:%d-%d",
				dup.Lines, dup.Format,
				dup.FirstFile.StartLoc.Line, dup.FirstFile.EndLoc.Line,
				secondFile, dup.SecondFile.StartLoc.Line, dup.SecondFile.EndLoc.Line)
		}

		b := models.NewViolationBuilder().
			WithFile(firstFile).
			WithLocation(dup.FirstFile.StartLoc.Line, 0).
			WithMessage(msg).
			WithSource("jscpd").
			WithRuleFromLinter("jscpd", fmt.Sprintf("duplicate-%s", dup.Format))
		if dup.Fragment != "" {
			b = b.WithCode(dup.Fragment)
		}
		violations = append(violations, b.Build())
	}

	return violations, nil
}

func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := doublestar.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

func normalizePath(workDir, name string) string {
	cleaned := filepath.Clean(name)
	if filepath.IsAbs(cleaned) {
		return cleaned
	}
	return filepath.Join(workDir, cleaned)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type JscpdReport struct {
	Duplicates []JscpdDuplicate `json:"duplicates"`
}

type JscpdDuplicate struct {
	Format     string       `json:"format"`
	Lines      int          `json:"lines"`
	Fragment   string       `json:"fragment"`
	Tokens     int          `json:"tokens"`
	FirstFile  JscpdFileRef `json:"firstFile"`
	SecondFile JscpdFileRef `json:"secondFile"`
}

type JscpdFileRef struct {
	Name     string   `json:"name"`
	Start    int      `json:"start"`
	End      int      `json:"end"`
	StartLoc JscpdLoc `json:"startLoc"`
	EndLoc   JscpdLoc `json:"endLoc"`
}

type JscpdLoc struct {
	Line     int `json:"line"`
	Column   int `json:"column"`
	Position int `json:"position"`
}
