package repomap

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type PathPattern struct {
	// A double asterisk (**) matches zero or more directories
	// A single asterisk (*) matches zero or more characters within a single directory
	// A '/' at the end matches only directories
	Pattern string
	// A prefix of '!' or '-' negates the pattern
	Negate bool
}

func (p PathPattern) Match(path string) bool {
	matched, err := doublestar.Match(p.Pattern, path)
	if err != nil {
		return false
	}
	if p.Negate {
		return !matched
	}
	return matched
}

// Multiple path patterns can be separated by commas or semicolons
type PathPatterns []PathPattern

// A map of string key-value pairs associated with a path pattern, if there is only
// single string without a key-value structure, the key will be "_"
type PathValue map[string]string

type PathMap map[PathPattern]PathValue

// A tree structure representing paths and their associated patterns
// keys are the directory path where the pathmap was defined
type PathTree map[string]PathMap

// GetValue get the PathValue for a given path by searching for a PathMap
// by traversing up the directory tree, multiple PathValues are merged
// with the closest path taking precedence, tr
// traversal continues until a .git directory in the parent is found
func (p PathTree) GetValue(path string) PathValue {
	result := make(PathValue)
	segments := strings.Split(path, "/")
	for i := len(segments); i >= 0; i-- {
		subPath := strings.Join(segments[:i], "/")
		if pathMap, ok := p[subPath]; ok {
			for pattern, value := range pathMap {

				if pattern.Match(path) {
					for k, v := range value {
						if _, exists := result[k]; !exists {
							result[k] = v
						}
					}
				}
			}
		}
		if _, err := os.Stat(filepath.Join(subPath, ".git")); err == nil {
			break
		}
	}
	return result
}

func ParseGitIgnorePattern(pattern string) PathMap {
	pathMap := make(PathMap)
	patterns := splitPatterns(pattern)
	for _, p := range patterns {
		negate := false
		trimmed := p
		if len(p) > 0 && (p[0] == '!' || p[0] == '-') {
			negate = true
			trimmed = p[1:]
		}
		// extract key-value if present (e.g. path/**. key=value key2=value2 or path/**. key)
		parts := strings.SplitN(trimmed, " ", 2)
		if len(parts) == 2 {
			trimmed = parts[0]
			kvParts := strings.Fields(parts[1])
			kvMap := make(PathValue)
			for _, kv := range kvParts {
				kvSplit := strings.SplitN(kv, "=", 2)
				if len(kvSplit) == 2 {
					kvMap[kvSplit[0]] = kvSplit[1]
				} else {
					kvMap["_"] = kvSplit[0]
				}
			}
			pathMap[PathPattern{
				Pattern: trimmed,
				Negate:  negate,
			}] = kvMap
			continue
		}
	}
	return pathMap
}

func splitPatterns(patterns string) []string {
	var result []string
	current := ""
	for _, line := range strings.Split(patterns, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		for _, r := range line {
			if r == ',' || r == ';' {
				if current != "" {
					result = append(result, current)
					current = ""
				}
			} else {
				current += string(r)
			}
		}
		if current != "" {
			result = append(result, current)
			current = ""
		}
	}
	return result
}
