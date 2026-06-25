package commit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxBodyLinesForDiff(t *testing.T) {
	cases := []struct {
		name         string
		changedLines int
		want         int
	}{
		{"trivial subject only", 5, 0},
		{"boundary 20 subject only", 20, 0},
		{"small", 21, 3},
		{"small boundary 100", 100, 3},
		{"medium", 101, 6},
		{"medium boundary 300", 300, 6},
		{"large", 301, 10},
		{"large boundary 800", 800, 10},
		{"huge", 801, 15},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, maxBodyLinesForDiff(tc.changedLines))
		})
	}
}

func TestCountDiffLines(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git a/foo.go b/foo.go",
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -1,3 +1,3 @@",
		" unchanged",
		"-removed line",
		"+added line",
		"+another added",
	}, "\n")

	// Only the two +/- content lines count; +++/--- headers are excluded.
	assert.Equal(t, 3, countDiffLines(diff))
}
