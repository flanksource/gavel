package git

import (
	"fmt"
	"strings"

	. "github.com/flanksource/gavel/models"
)

// ParsePatch takes a git patch string and returns a slice of CommitChange
// representing the changes made in the commit.
func ParsePatch(patch string) ([]CommitChange, error) {
	if patch == "" {
		return nil, nil
	}

	lines := strings.Split(patch, "\n")
	var changes []CommitChange
	var currentFile string
	var adds, dels int
	var changeType SourceChangeType
	var currentLine int

	// Pre-allocate linesChanged with upper bound to avoid reallocations
	linesChanged := make([]int, 0, len(lines))

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "diff --git") {
			if currentFile != "" {
				change := CommitChange{
					File:         currentFile,
					Type:         changeType,
					Adds:         adds,
					Dels:         dels,
					LinesChanged: NewLineRanges(linesChanged),
				}
				changes = append(changes, change)
			}

			// Extract file path from "diff --git a/<path> b/<path>"
			// Handle both quoted and unquoted paths
			idx := strings.Index(line, ` "b/`)
			if idx != -1 {
				// Quoted path: extract between "b/ and closing "
				path := line[idx+4:] // +4 to skip ` "b/`
				if endQuote := strings.Index(path, `"`); endQuote != -1 {
					currentFile = path[:endQuote]
				}
			} else {
				// Unquoted path: extract after b/ to end of line
				idx = strings.Index(line, " b/")
				if idx != -1 {
					currentFile = line[idx+3:] // +3 to skip " b/"
				}
			}
			adds, dels = 0, 0
			linesChanged = linesChanged[:0] // Reset length while keeping capacity
			currentLine = 0
			changeType = SourceChangeTypeModified
		} else if strings.HasPrefix(line, "new file") {
			changeType = SourceChangeTypeAdded
		} else if strings.HasPrefix(line, "deleted file") {
			changeType = SourceChangeTypeDeleted
		} else if strings.HasPrefix(line, "rename from") {
			changeType = SourceChangeTypeRenamed
		} else if strings.HasPrefix(line, "@@") {
			// Parse hunk header to get starting line number
			// Format: @@ -old_start,old_count +new_start,new_count @@
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				newRange := strings.TrimPrefix(parts[2], "+")
				fmt.Sscanf(newRange, "%d", &currentLine)
			}
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			adds++
			linesChanged = append(linesChanged, currentLine)
			currentLine++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			dels++
			// Deletions don't increment the line number in the new file
		} else if !strings.HasPrefix(line, "diff") && !strings.HasPrefix(line, "index") &&
			!strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "+++") &&
			!strings.HasPrefix(line, "new file") && !strings.HasPrefix(line, "deleted file") &&
			!strings.HasPrefix(line, "rename") && !strings.HasPrefix(line, "similarity") &&
			!strings.HasPrefix(line, "Binary files") && currentLine > 0 {
			// Context lines increment the line number
			currentLine++
		}
	}

	if currentFile != "" {
		changes = append(changes, CommitChange{
			File:         currentFile,
			Type:         changeType,
			Adds:         adds,
			Dels:         dels,
			LinesChanged: NewLineRanges(linesChanged),
		})
	}

	return changes, nil
}
