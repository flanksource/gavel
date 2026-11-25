package verify

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
)

// EditConfig configures which files `gavel edit` lists beneath each .gavel.yaml.
type EditConfig struct {
	// Extensions are the file extensions listed (e.g. ".md", ".go"); empty lists
	// the built-in editable documents and images. Extensions the editor has no
	// richer view for open as plain source.
	Extensions []string `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	// IncludeIgnored are root-relative doublestar patterns whose files are listed
	// even though .gitignore excludes them; a bare directory includes everything
	// below it.
	IncludeIgnored []string `yaml:"includeIgnored,omitempty" json:"includeIgnored,omitempty"`
}

// Normalize lower-cases the extensions and cleans the patterns, failing on an
// extension that is not ".<name>" or a pattern that is invalid or leaves the root.
func (c EditConfig) Normalize() (EditConfig, error) {
	out := EditConfig{}
	for _, ext := range c.Extensions {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if len(ext) < 2 || ext[0] != '.' || strings.ContainsAny(ext, `/\*`) {
			return EditConfig{}, fmt.Errorf("edit.extensions: %q must look like .md", ext)
		}
		out.Extensions = append(out.Extensions, ext)
	}
	for _, pattern := range c.IncludeIgnored {
		cleaned := path.Clean(filepath.ToSlash(strings.TrimSpace(pattern)))
		if strings.TrimSpace(pattern) == "" || cleaned == "." || path.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
			return EditConfig{}, fmt.Errorf("edit.includeIgnored: %q must be a pattern inside the folder", pattern)
		}
		if !doublestar.ValidatePattern(cleaned) {
			return EditConfig{}, fmt.Errorf("edit.includeIgnored: invalid glob %q", pattern)
		}
		out.IncludeIgnored = append(out.IncludeIgnored, cleaned)
	}
	return out, nil
}

var editConfigMu sync.Mutex

// SaveEditConfig replaces the edit section of directory's .gavel.yaml, keeping
// every other section.
func SaveEditConfig(directory string, cfg EditConfig) error {
	cfg, err := cfg.Normalize()
	if err != nil {
		return err
	}
	editConfigMu.Lock()
	defer editConfigMu.Unlock()
	file := filepath.Join(directory, GavelConfigFileName)
	existing, err := LoadSingleGavelConfig(file)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	existing.Edit = cfg
	return SaveGavelConfig(file, existing)
}
