package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/git/rules"
	"github.com/flanksource/gavel/repomap"
	"github.com/flanksource/commons/logger"

	. "github.com/flanksource/gavel/models"
)

func NewCommit(message string) Commit {
	var c = Commit{
		Trailers: make(map[string]string),
		Headers:  make(map[string]string),
	}
	// Parse subject and body
	lines := strings.SplitN(message, "\n", 2)
	c.Subject = strings.TrimSpace(lines[0])
	if len(lines) > 1 {
		c.Body = strings.TrimSpace(lines[1])
	}

	// Parse trailers
	c.Body, c.Trailers = parseTrailers(c.Body)

	// Parse references
	c.Subject, c.Reference = parseReference(c.Subject)

	// parse commit type and scope
	c.CommitType, c.Scope, c.Subject = parseCommitTypeAndScope(c.Subject)
	c.Body = strings.TrimSpace(c.Body)

	return c
}

// AnalyzerContext holds the context and configuration for git analysis operations
type AnalyzerContext struct {
	context.Context
	Arch           *repomap.ArchConf
	severityEngine *rules.Engine
}

// NewAnalyzerContext creates a new AnalyzerContext with the given context and repository path
func NewAnalyzerContext(ctx context.Context, repoPath string) (*AnalyzerContext, error) {
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", repoPath, err)
	}
	arch, err := repomap.GetConf(abs)
	if err != nil {
		return nil, fmt.Errorf("failed to load arch config for %s: %w", abs, err)
	}

	logger.Debugf("Loaded arch config for repo %s", arch.Pretty().ANSI())

	ac := &AnalyzerContext{
		Context: ctx,
		Arch:    arch,
	}

	// Initialize severity engine if rules are configured
	if len(arch.Severity.Rules) > 0 {
		engine, err := rules.NewEngine(&arch.Severity)
		if err != nil {
			logger.Warnf("Failed to initialize severity engine: %v, will fall back to simple logic", err)
		} else {
			ac.severityEngine = engine
		}
	}

	return ac, nil
}

// RepoPath returns the git repository path
func (ac *AnalyzerContext) RepoPath() string {
	if ac.Arch == nil {
		return ""
	}
	return ac.Arch.RepoPath()
}

// ReadFile reads a file from the repository at the given commit
func (ac *AnalyzerContext) ReadFile(path, commit string) (string, error) {
	if ac.Arch == nil {
		return "", fmt.Errorf("arch config not initialized")
	}
	return ac.Arch.ReadFile(path, commit)
}

// GetFileMap returns file mapping information for the given path
func (ac *AnalyzerContext) GetFileMap(path string, commit string) (*FileMap, error) {
	if ac.Arch == nil {
		return nil, fmt.Errorf("arch config not initialized")
	}
	return ac.Arch.GetFileMap(path, commit)
}

// GetSeverityConfig returns the severity configuration from ArchConf
func (ac *AnalyzerContext) GetSeverityConfig() *rules.SeverityConfig {
	if ac.Arch == nil {
		return nil
	}
	return &ac.Arch.Severity
}

// GetSeverityEngine returns the cached severity engine
func (ac *AnalyzerContext) GetSeverityEngine() *rules.Engine {
	return ac.severityEngine
}
