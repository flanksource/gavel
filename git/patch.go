package git

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/internal/linecounts"
	"github.com/flanksource/gavel/models"
)

// ParsePatch takes a git patch string and returns a slice of CommitChange
// representing the changes made in the commit.
func ParsePatch(patch string) ([]models.CommitChange, error) {
	if patch == "" {
		return nil, nil
	}

	lines := strings.Split(patch, "\n")
	var changes []models.CommitChange
	var currentFile string
	var adds, dels int
	var changeType models.SourceChangeType
	var currentLine int
	var countLines bool

	// Pre-allocate linesChanged with upper bound to avoid reallocations
	linesChanged := make([]int, 0, len(lines))

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "diff --git") {
			if currentFile != "" {
				change := models.CommitChange{
					File:         currentFile,
					Type:         changeType,
					Adds:         adds,
					Dels:         dels,
					LinesChanged: models.NewLineRanges(linesChanged),
				}
				changes = append(changes, change)
			}

			_, currentFile = diffGitPaths(line)
			adds, dels = 0, 0
			countLines = linecounts.Supported(currentFile)
			linesChanged = linesChanged[:0] // Reset length while keeping capacity
			currentLine = 0
			changeType = models.SourceChangeTypeModified
		} else if strings.HasPrefix(line, "new file") {
			changeType = models.SourceChangeTypeAdded
		} else if strings.HasPrefix(line, "deleted file") {
			changeType = models.SourceChangeTypeDeleted
		} else if strings.HasPrefix(line, "rename from") {
			changeType = models.SourceChangeTypeRenamed
		} else if strings.HasPrefix(line, "Binary files ") || line == "GIT binary patch" {
			countLines = false
		} else if !countLines {
			continue
		} else if strings.HasPrefix(line, "@@") {
			// Parse hunk header to get starting line number
			// Format: @@ -old_start,old_count +new_start,new_count @@
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				newRange := strings.TrimPrefix(parts[2], "+")
				_, _ = fmt.Sscanf(newRange, "%d", &currentLine)
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
		changes = append(changes, models.CommitChange{
			File:         currentFile,
			Type:         changeType,
			Adds:         adds,
			Dels:         dels,
			LinesChanged: models.NewLineRanges(linesChanged),
		})
	}

	return changes, nil
}
