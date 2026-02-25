package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

const (
	maxWholeFileLines = 200
	contextRadius     = 10
)

func ReadSourceLines(workDir string, ref types.PathRef) (string, error) {
	path := ref.File
	if !filepath.IsAbs(path) {
		path = filepath.Join(workDir, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	if ref.IsWholeFile() {
		if len(lines) > maxWholeFileLines {
			lines = lines[:maxWholeFileLines]
		}
		return formatLines(lines, 1), nil
	}
	if ref.EndLine > 0 {
		start := clamp(ref.Line-1, 0, len(lines))
		end := clamp(ref.EndLine, 0, len(lines))
		return formatLines(lines[start:end], ref.Line), nil
	}
	start := clamp(ref.Line-1-contextRadius, 0, len(lines))
	end := clamp(ref.Line+contextRadius, 0, len(lines))
	return formatLines(lines[start:end], start+1), nil
}

func formatLines(lines []string, startLine int) string {
	var sb strings.Builder
	for i, line := range lines {
		fmt.Fprintf(&sb, "%4d | %s\n", startLine+i, line)
	}
	return sb.String()
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
