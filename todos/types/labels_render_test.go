package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/gavel/todos/labels"
)

func TestPrettyRowLabelsColumn(t *testing.T) {
	t.Run("renders resolved definitions", func(t *testing.T) {
		todo := TODO{
			Labels:           []string{"bug", "ui"},
			LabelDefinitions: labels.NewResolver(nil, nil).ResolveAll([]string{"bug", "ui"}),
			TODOFrontmatter:  TODOFrontmatter{Title: "Fix it"},
		}
		row := todo.PrettyRow(nil)

		require.Contains(t, row, "Labels")
		rendered := row["Labels"].String()
		assert.Contains(t, rendered, "bug")
		assert.Contains(t, rendered, "ui")
		assert.Contains(t, row["Labels"].ANSI(), "\x1b[", "labels must be coloured in the terminal")
	})

	// A TODO from a source with no definition store still has labels; they must
	// render with the hashed palette colour rather than vanishing.
	t.Run("falls back to derived colours when definitions are absent", func(t *testing.T) {
		todo := TODO{Labels: []string{"bug", "whatever"}, TODOFrontmatter: TODOFrontmatter{Title: "No defs"}}
		row := todo.PrettyRow(nil)

		require.Contains(t, row, "Labels")
		assert.Contains(t, row["Labels"].String(), "whatever")
		assert.Contains(t, row["Labels"].ANSI(), "\x1b[")
	})

	// clicky derives table headers from the first row only, so the column is
	// emitted unconditionally — otherwise it would appear or disappear depending
	// on whether the first todo in the list happened to carry a label.
	t.Run("emits an empty cell rather than dropping the column", func(t *testing.T) {
		todo := TODO{TODOFrontmatter: TODOFrontmatter{Title: "Bare"}}
		row := todo.PrettyRow(nil)
		require.Contains(t, row, "Labels")
		assert.Empty(t, row["Labels"].String())
	})
}

func TestPrettyDetailedParentLine(t *testing.T) {
	const parentID = "0a1b2c3d-5f56-4c2c-833f-83e9b7774657"
	child := TODO{ParentID: parentID, TODOFrontmatter: TODOFrontmatter{Title: "Split out"}}
	assert.Contains(t, child.PrettyDetailed().String(), "Parent: "+parentID)

	t.Run("omits the line for a top-level TODO", func(t *testing.T) {
		topLevel := TODO{TODOFrontmatter: TODOFrontmatter{Title: "Top level"}}
		assert.NotContains(t, topLevel.PrettyDetailed().String(), "Parent:")
	})
}

func TestShortID(t *testing.T) {
	for id, want := range map[string]string{
		"0a1b2c3d-5f56-4c2c-833f-83e9b7774657": "0a1b2c3d",
		"0a1b2c3d":                             "0a1b2c3d",
		"0a1b":                                 "0a1b",
		"":                                     "",
	} {
		assert.Equal(t, want, ShortID(id), "ShortID(%q)", id)
	}
}

func TestPrettyDetailedLabelsLine(t *testing.T) {
	todo := TODO{
		Labels:          []string{"security"},
		TODOFrontmatter: TODOFrontmatter{Title: "Lock it down"},
	}
	rendered := todo.PrettyDetailed().String()
	assert.Contains(t, rendered, "Labels:")
	assert.Contains(t, rendered, "security")

	t.Run("omits the line when there are no labels", func(t *testing.T) {
		bare := TODO{TODOFrontmatter: TODOFrontmatter{Title: "Bare"}}
		assert.NotContains(t, bare.PrettyDetailed().String(), "Labels:")
	})
}
