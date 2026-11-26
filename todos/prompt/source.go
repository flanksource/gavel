// Reads the source lines a todo's path refs name into its prompt, from files inside the
// todo's working directory only.
package prompt

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/utils"
)

const (
	maxWholeFileLines = 200
	contextRadius     = 10
)

// ReadSourceLines returns the lines ref names, numbered, from a file inside workDir. A todo's
// path refs come from its frontmatter, which a hosted todo's author writes, so a ref that leads
// outside workDir — by "..", an absolute path, or a symlink — is refused instead of read.
func ReadSourceLines(workDir string, ref types.PathRef) (string, error) {
	data, err := utils.ReadFileWithin(workDir, ref.File)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
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
