package git

import "strings"

const maxDiffBytes = 256 * 1024

// TruncateDiff caps a rendered diff at a complete line so API responses cannot
// grow without bound. The bool reports whether content was removed.
func TruncateDiff(diff string) (string, bool) {
	if len(diff) <= maxDiffBytes {
		return diff, false
	}
	cut := diff[:maxDiffBytes]
	if newline := strings.LastIndexByte(cut, '\n'); newline > 0 {
		cut = cut[:newline+1]
	}
	return cut, true
}

// IsBinaryDiff reports whether a unified diff holds at least one file and every
// file in it is binary, i.e. there is no text hunk a diff viewer could show.
func IsBinaryDiff(diff string) bool {
	files := parseCommitFiles(diff)
	for _, file := range files {
		if !file.Binary {
			return false
		}
	}
	return len(files) > 0
}
