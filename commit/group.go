package commit

import (
	"fmt"
	"strings"
)

// commitGroup is one logical commit: a label and the staged changes it collects.
type commitGroup struct {
	Label string
	// Message, when set, is used verbatim as the commit message instead of
	// invoking the LLM. Used by the chore group (lock/generated files) so
	// trivial bundles do not spend an analysis call.
	Message string
	Changes []stagedChange
}

func (g commitGroup) Files() []string {
	files := make([]string, 0, len(g.Changes))
	for _, change := range g.Changes {
		files = append(files, change.Path)
	}
	return files
}

func (g commitGroup) GitPaths() []string {
	seen := make(map[string]struct{}, len(g.Changes)*2)
	var paths []string
	for _, change := range g.Changes {
		for _, path := range change.GitPaths() {
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			paths = append(paths, path)
		}
	}
	return paths
}

func (g commitGroup) diff() string {
	var patches []string
	for _, change := range g.Changes {
		patches = append(patches, strings.TrimRight(change.Patch, "\n"))
	}
	if len(patches) == 0 {
		return ""
	}
	return strings.Join(patches, "\n") + "\n"
}

func (g commitGroup) labelOrDefault() string {
	if g.Label != "" {
		return g.Label
	}
	if len(g.Changes) == 1 {
		return g.Changes[0].Path
	}
	return fmt.Sprintf("%d files", len(g.Changes))
}
