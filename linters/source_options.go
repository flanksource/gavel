package linters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/models"
)

type TagsMapper interface {
	TagsArgs(tags []string) []string
}

func filterGeneratedViolations(opts RunOptions, violations []models.Violation) ([]models.Violation, error) {
	if opts.Generated == nil || *opts.Generated {
		return violations, nil
	}
	generated := map[string]bool{}
	kept := make([]models.Violation, 0, len(violations))
	for _, violation := range violations {
		if violation.File == "" {
			kept = append(kept, violation)
			continue
		}
		path := violation.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(opts.WorkDir, path)
		}
		if _, checked := generated[path]; !checked {
			content, err := os.ReadFile(path)
			if err != nil {
				return violations, fmt.Errorf("check generated source %s: %w", path, err)
			}
			generated[path] = generatedSource(filepath.Base(path), string(content))
		}
		if !generated[path] {
			kept = append(kept, violation)
		}
	}
	return kept, nil
}

func generatedSource(name, content string) bool {
	if strings.HasSuffix(name, "_gen.go") || strings.HasSuffix(name, ".gen.go") || strings.HasSuffix(name, ".pb.go") || strings.Contains(name, ".generated.") {
		return true
	}
	blockEnd := ""
	for line := range strings.SplitSeq(strings.TrimPrefix(content, "\ufeff"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if blockEnd == "" {
			switch {
			case strings.HasPrefix(line, "/*"):
				blockEnd = "*/"
			case strings.HasPrefix(line, "<!--"):
				blockEnd = "-->"
			case strings.HasPrefix(line, "//"), strings.HasPrefix(line, "#"):
			default:
				return false
			}
		}
		comment := line
		if blockEnd != "" {
			if before, after, closed := strings.Cut(line, blockEnd); closed {
				comment, blockEnd = before, ""
				if strings.TrimSpace(after) != "" {
					return generatedMarker(comment)
				}
			}
		}
		if generatedMarker(comment) {
			return true
		}
	}
	return false
}

func generatedMarker(comment string) bool {
	comment = strings.ToLower(comment)
	return strings.Contains(comment, "generated") && strings.Contains(comment, "do not edit") ||
		strings.Contains(" "+strings.Trim(comment, "/*#<!-> \t")+" ", " @generated ")
}
