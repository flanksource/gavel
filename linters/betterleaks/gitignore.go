package betterleaks

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/utils"
)

func gitIgnoredScanPaths(workDir string) ([]string, error) {
	gitRoot, err := utils.GitTopLevel(workDir)
	if err != nil {
		return nil, err
	}
	if gitRoot == "" {
		return nil, nil
	}

	paths := make(map[string]struct{})
	commands := [][]string{
		{"ls-files", "--others", "--ignored", "--exclude-standard", "--directory", "-z"},
		{"ls-files", "--cached", "--ignored", "--exclude-standard", "-z"},
	}
	for _, args := range commands {
		cmd := exec.Command("git", args...)
		cmd.Dir = workDir
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("git %s in %q: %w: %s", strings.Join(args, " "), workDir, err, strings.TrimSpace(stderr.String()))
		}
		for _, path := range strings.Split(string(output), "\x00") {
			if path != "" {
				paths[filepath.ToSlash(path)] = struct{}{}
			}
		}
	}

	generatedDir := filepath.Join(workDir, ".tmp")
	ignoredGenerated, err := utils.GitIgnoredPaths([]string{generatedDir}, workDir)
	if err != nil {
		return nil, err
	}
	if _, ok := ignoredGenerated[generatedDir]; ok {
		paths[".tmp/"] = struct{}{}
	}

	result := make([]string, 0, len(paths))
	for path := range paths {
		result = append(result, path)
	}
	sort.Strings(result)
	return result, nil
}
