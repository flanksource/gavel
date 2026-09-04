package labels_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/gavel/todos/labels"
)

const schemaPath = "../../internal/database/schema/todos.hcl"

// The todo_labels_color_check constraint enumerates the palette in SQL. If the
// Go palette gains a hue and the constraint does not, every write of that hue
// fails at the database with a constraint violation instead of at validation —
// so the two lists are asserted equal rather than trusted to stay in step.
func TestColorCheckConstraintMatchesPalette(t *testing.T) {
	source, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	constraint := regexp.MustCompile(`(?s)check "todo_labels_color_check" \{.*?expr\s*=\s*"(.*?)"\s*\n`)
	match := constraint.FindSubmatch(source)
	require.NotNil(t, match, "todo_labels_color_check not found in %s", schemaPath)

	hues := regexp.MustCompile(`'([a-z]+)'::text`).FindAllSubmatch(match[1], -1)
	require.NotEmpty(t, hues)

	inSQL := make([]string, 0, len(hues))
	for _, hue := range hues {
		inSQL = append(inSQL, string(hue[1]))
	}

	assert.Equal(t, labels.PaletteStrings(), inSQL,
		"todo_labels_color_check and labels.Palette() have drifted")
}

// The name check must mirror labels.Normalize, or a definition written through
// the repository can be rejected by the database (or worse, a hand-written row
// can be stored in a form the resolver never matches).
func TestNameCheckMirrorsNormalize(t *testing.T) {
	source, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	assert.Contains(t, string(source), `name <> '' AND name = lower(btrim(name))`,
		"todo_labels_name_normalized must mirror labels.Normalize (lowercase + trim)")
}
