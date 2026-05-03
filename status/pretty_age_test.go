package status

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrettyRowIncludesAge(t *testing.T) {
	now := time.Now()
	files := []FileStatus{{
		Path:       "cmd/gavel/main.go",
		State:      StateUnstaged,
		WorkKind:   KindModified,
		Adds:       3,
		Dels:       1,
		ModifiedAt: now.Add(-3 * time.Hour),
	}}

	r := &Result{Branch: "feat/x", Files: files}
	plain := stripANSIBytes(r.Pretty().ANSI())

	assert.Contains(t, plain, "3h", "pretty row should render the file's relative age")
}

func TestPrettyRowOmitsAgeWhenUnknown(t *testing.T) {
	files := []FileStatus{{
		Path:     "cmd/gavel/main.go",
		State:    StateUnstaged,
		WorkKind: KindModified,
	}}
	r := &Result{Branch: "feat/x", Files: files}
	plain := stripANSIBytes(r.Pretty().ANSI())

	// No mtime → no "Xs/m/h/d" tokens between delta column and enrichment.
	for _, suffix := range []string{"s ", "m ", "h ", "d "} {
		assert.NotContains(t, plain, " 0"+suffix,
			"row without a known mtime must not render a zero-age chip")
	}
}

// stripANSIBytes removes ANSI SGR sequences for readable substring asserts.
func stripANSIBytes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] != 'm') {
				j++
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
