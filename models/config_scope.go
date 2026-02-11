package models

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// GitConfig represents git-related configuration
type GitConfig struct {
	Commits              CommitsConfig `yaml:"commits,omitempty"`
	VersionFieldPatterns []string      `yaml:"version_field_patterns,omitempty"`
}

// AnalysisConfig represents git commit analysis configuration
type AnalysisConfig struct {
	AIModel  string `yaml:"ai_model,omitempty"`
	MinScore int    `yaml:"min_score,omitempty"`
	MinWords int    `yaml:"min_words,omitempty"`
	MaxWords int    `yaml:"max_words,omitempty"`
	MaxFiles int    `yaml:"max_files,omitempty"`
}

// CommitsConfig represents git commit validation configuration
type CommitsConfig struct {
	Enabled           bool           `yaml:"enabled"`
	AllowedTypes      []string       `yaml:"allowed_types,omitempty"`
	Blocklist         []string       `yaml:"blocklist,omitempty"`
	RequiredTrailers  []string       `yaml:"required_trailers,omitempty"`
	RequiredReference bool           `yaml:"required_reference,omitempty"`
	RequiredScope     bool           `yaml:"required_scope,omitempty"`
	Analysis          AnalysisConfig `yaml:"analysis,omitempty"`
}

// BuildConfig represents build tool configuration
type BuildConfig struct {
	Enabled  bool              `yaml:"enabled"`
	Tool     string            `yaml:"tool,omitempty"`
	Commands map[string]string `yaml:"commands,omitempty"`
}

// TestingConfig represents testing tool configuration
type TestingConfig struct {
	Tool  string   `yaml:"tool,omitempty"`
	Files []string `yaml:"files,omitempty"`
}

// GolangConfig represents golang-specific configuration
type GolangConfig struct {
	Enabled   bool          `yaml:"enabled"`
	Blocklist []string      `yaml:"blocklist,omitempty"`
	Testing   TestingConfig `yaml:"testing,omitempty"`
}

// PathRule represents a rule for a specific scope
type PathRule struct {
	Path   string `yaml:"path"`
	Prefix string `yaml:"prefix,omitempty"`
}

func (r PathRule) Match(path string) bool {
	matched, _ := doublestar.Match(r.Path, path)
	return matched
}

// ScopesConfig represents scope definitions configuration
type ScopesConfig struct {
	AllowedScopes []string  `yaml:"allowed_scopes,omitempty"`
	Rules         PathRules `yaml:"rules,omitempty"`
}

type TechnologyConfig struct {
	Rules PathRules `yaml:"rules,omitempty"`
}

func (tc TechnologyConfig) GetTechByPath(path string) Technology {
	var techs []ScopeTechnology
	for _, tech := range tc.Rules.Apply(path) {
		techs = append(techs, ScopeTechnology(tech))
	}
	return techs
}

// Validate checks if the scope configuration is valid
// Returns an error if scope names are invalid or patterns have syntax errors
// If AllowedScopes is defined, validates that all scope names in Rules are in AllowedScopes
// If AllowedScopes is empty, allows any scope (including custom user-defined scopes)
func (sc *ScopesConfig) Validate() error {
	if sc == nil {
		return nil
	}

	// Build allowed scopes map if defined
	var allowedScopes map[string]bool
	if len(sc.AllowedScopes) > 0 {
		allowedScopes = make(map[string]bool, len(sc.AllowedScopes))
		for _, scope := range sc.AllowedScopes {
			allowedScopes[scope] = true
		}
	}

	// Validate each scope and its rules
	for scopeName, rules := range sc.Rules {
		// If allowed_scopes is defined, check if scope name is in the list
		if allowedScopes != nil && !allowedScopes[scopeName] {
			return fmt.Errorf("invalid scope name '%s': not in allowed_scopes list %v", scopeName, sc.AllowedScopes)
		}

		// Validate each pattern
		for _, rule := range rules {
			// Check pattern syntax
			_, err := filepath.Match(rule.Path, "test")
			if err != nil {
				return fmt.Errorf("invalid glob pattern '%s' in scope '%s': %w", rule.Path, scopeName, err)
			}
		}
	}

	return nil
}

type PathRules map[string][]PathRule

// ApplyScopeRules matches a file path against scope rules and returns the appropriate scope
// Uses "most specific wins" priority when multiple patterns match
// Falls back to git.GetScopeByPath when no rules match
func (sc *ScopesConfig) GetScopesByPath(path string) Scopes {
	scopes := Scopes{}

	for _, scope := range sc.Rules.Apply(path) {
		scopes = append(scopes, ScopeType(scope))
	}

	return scopes
}

func (pr PathRules) Apply(path string) []string {

	type scopeMatch struct {
		value       string
		specificity int
	}
	var matches []scopeMatch

	// Check each scope's rules
	for scopeName, rules := range pr {
		for _, rule := range rules {
			pattern := rule.Path
			// Try matching against full path
			matched := rule.Match(path)

			// If full path doesn't match, try matching against basename
			if !matched {
				matched = rule.Match(filepath.Base(path))
			}

			if matched {
				// Calculate specificity: longer paths and fewer wildcards = more specific
				specificity := len(pattern)
				wildcardCount := strings.Count(pattern, "*") + strings.Count(pattern, "?")
				specificity -= wildcardCount * 10 // Penalize wildcards

				matches = append(matches, scopeMatch{
					value:       scopeName,
					specificity: specificity,
				})
			}
		}
	}

	var results []string
	// Return most specific match
	if len(matches) > 0 {
		// Sort by specificity (highest first)
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].specificity > matches[j].specificity
		})
		for _, match := range matches {
			results = append(results, match.value)
		}
	}

	return results
}
