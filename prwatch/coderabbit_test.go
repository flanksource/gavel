package prwatch

import (
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
)

func TestParseSeverityFromBadge(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{"critical", "_🔴 Critical_\n\n**Some issue**", "critical"},
		{"major", "_🟠 Major_\n\nDescription", "major"},
		{"minor", "_🟡 Minor_\n\nSmall thing", "minor"},
		{"no badge", "Just a regular comment", ""},
		{"badge in middle", "prefix _🔴 Critical_ suffix", "critical"},
		{"combined badges uses first", "_⚠️ Potential issue_ | _🟠 Major_", "major"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseSeverityFromBadge(tc.body))
		})
	}
}

func TestParseNitpickComments(t *testing.T) {
	t.Run("real coderabbit format with blockquote", func(t *testing.T) {
		body := `**Actionable comments posted: 8**

<details>
<summary>🧹 Nitpick comments (2)</summary><blockquote>

<details>
<summary>task/manager_output.go (1)</summary><blockquote>

` + "`23-44`" + `: **` + "`bufferingWriter`" + ` does not handle partial lines.**

Write is called with whatever chunk Read returns.

</blockquote></details>
<details>
<summary>shutdown/shutdown.go (1)</summary><blockquote>

` + "`90-96`" + `: **` + "`restoreTerminal()`" + ` is called twice on the panic path.**

Shutdown() already calls restoreTerminal().

<details>
<summary>♻️ Suggested fix</summary>

` + "```diff\n-restoreTerminal()\n```" + `
</details>

</blockquote></details>

</blockquote></details>`

		comment := github.PRComment{
			ID: 100, Author: "coderabbitai[bot]", URL: "https://example.com",
			Body: body,
		}
		results := parseNitpickComments(comment)
		assert.Len(t, results, 2)

		assert.Equal(t, "task/manager_output.go", results[0].Path)
		assert.Equal(t, 23, results[0].Line)
		assert.Equal(t, "nitpick", results[0].Severity)
		assert.Contains(t, results[0].Body, "does not handle partial lines")

		assert.Equal(t, "shutdown/shutdown.go", results[1].Path)
		assert.Equal(t, 90, results[1].Line)
		assert.Equal(t, "nitpick", results[1].Severity)
		assert.Contains(t, results[1].Body, "called twice on the panic path")
		assert.NotContains(t, results[1].Body, "Suggested fix", "nested details should be stripped")
	})

	t.Run("inherits the parent's resolution state but is not itself a thread", func(t *testing.T) {
		body := `<details>
<summary>🧹 Nitpick comments (1)</summary><blockquote>

<details>
<summary>a.go (1)</summary><blockquote>

` + "`12-14`" + `: **tidy this up.**

</blockquote></details>

</blockquote></details>`
		parent := github.PRComment{
			ID: 400, Author: "coderabbitai[bot]", Body: body,
			IsReviewThread: true, IsResolved: true,
		}

		results := parseNitpickComments(parent)

		assert.Len(t, results, 1)
		assert.True(t, results[0].IsResolved, "a nitpick under a resolved thread is not live work")
		assert.False(t, results[0].IsUnresolved())
		assert.False(t, results[0].IsReviewThread, "a parsed fragment has no resolve button of its own")
	})

	t.Run("no nitpick section returns nil", func(t *testing.T) {
		comment := github.PRComment{
			ID: 200, Body: "LGTM! No nitpicks.", Author: "reviewer",
		}
		assert.Nil(t, parseNitpickComments(comment))
	})

	t.Run("empty file blocks skipped", func(t *testing.T) {
		body := `<details>
<summary>🧹 Nitpick comments (1)</summary><blockquote>

<details>
<summary>empty.go (1)</summary><blockquote>

</blockquote></details>

</blockquote></details>`
		comment := github.PRComment{ID: 300, Body: body, Author: "bot"}
		assert.Empty(t, parseNitpickComments(comment))
	})
}
