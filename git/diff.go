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
