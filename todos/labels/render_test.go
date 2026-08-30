package labels_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/gavel/todos/labels"
)

// The reason this package renders with clicky.Text instead of api.LabelBadge:
// LabelBadge.ANSI() returns its plain String(), so a badge is monochrome in a
// terminal. This asserts a chip actually carries SGR escapes.
func TestChipIsColouredInTheTerminal(t *testing.T) {
	def := labels.Definition{Name: "bug", Color: labels.ColorRed, Icon: "debug", Scope: labels.ScopeBuiltin}

	ansi := def.Text().ANSI()
	assert.Contains(t, ansi, "\x1b[", "chip carries no ANSI escape — it would render monochrome")
	assert.Contains(t, ansi, "bug")

	plain := def.Text().String()
	assert.NotContains(t, plain, "\x1b[", "String() must stay plain for non-tty output")
	assert.Contains(t, plain, "bug")
}

func TestChipCarriesItsIconGlyph(t *testing.T) {
	withIcon := labels.Definition{Name: "bug", Color: labels.ColorRed, Icon: "debug"}
	without := labels.Definition{Name: "bug", Color: labels.ColorRed}

	assert.Greater(t, len(withIcon.Text().String()), len(without.Text().String()),
		"the icon glyph should prefix the chip")
}

// A derived chip has no icon, but must still render its name and colour.
func TestDerivedChipRendersWithoutIcon(t *testing.T) {
	def := labels.NewResolver(nil, nil).Resolve("undefined-label")
	require.Equal(t, labels.ScopeDerived, def.Scope)

	assert.Equal(t, "", def.Icon)
	assert.Contains(t, def.Text().String(), "undefined-label")
	assert.Contains(t, def.Text().ANSI(), "\x1b[")
}

// An icon naming a key outside the registry must degrade to no glyph rather
// than printing a placeholder box.
func TestUnknownIconDegradesToNoGlyph(t *testing.T) {
	def := labels.Definition{Name: "bug", Color: labels.ColorRed, Icon: "no-such-glyph"}
	assert.Equal(t, "", def.IconFor().Unicode)
	assert.Contains(t, def.Text().String(), "bug")
}

// The glyph must take the definition's own hue, so a chip's icon and its tint
// never disagree.
func TestIconAdoptsTheDefinitionColour(t *testing.T) {
	def := labels.Definition{Name: "perf", Color: labels.ColorAmber, Icon: "performance"}
	assert.Equal(t, labels.ColorAmber.TextClass(), def.IconFor().Style)
}

func TestDefinitionsTextJoinsChips(t *testing.T) {
	defs := labels.Derive([]string{"one", "two"})
	rendered := defs.Text().String()
	assert.Contains(t, rendered, "one")
	assert.Contains(t, rendered, "two")
	assert.Equal(t, 1, strings.Count(rendered, "one"))
}

// The dashboard cannot resolve a clicky registry key, so the JSON must carry the
// Iconify name it renders from.
func TestDefinitionJSONCarriesIconifyName(t *testing.T) {
	encoded, err := json.Marshal(labels.Definition{
		Name: "bug", Color: labels.ColorRed, Icon: "debug", Scope: labels.ScopeBuiltin,
	})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(encoded, &decoded))

	assert.Equal(t, "bug", decoded["name"])
	assert.Equal(t, "red", decoded["color"])
	assert.Equal(t, "debug", decoded["icon"], "the stored registry key survives")
	assert.NotEmpty(t, decoded["iconify"], "the resolved Iconify name must reach the browser")
	assert.Contains(t, decoded["iconify"], ":", "an Iconify name is prefix:glyph")

	t.Run("a definition with no icon omits it", func(t *testing.T) {
		encoded, err := json.Marshal(labels.Definition{Name: "x", Color: labels.ColorSlate})
		require.NoError(t, err)
		assert.NotContains(t, string(encoded), "iconify")
	})
}

func TestPrettyRowColumns(t *testing.T) {
	def := labels.Definition{
		Name: "bug", Color: labels.ColorRed, Icon: "debug",
		Description: "Something is broken.", Scope: labels.ScopeBuiltin,
	}
	row := def.PrettyRow(nil)

	assert.Contains(t, row, "Label")
	assert.Contains(t, row, "Color")
	assert.Contains(t, row, "Scope")
	assert.Contains(t, row, "Icon")
	assert.Contains(t, row, "Description")
	assert.Contains(t, row["Color"].String(), "red")
	assert.Contains(t, row["Scope"].String(), "builtin")

	t.Run("omits empty optional columns", func(t *testing.T) {
		bare := labels.Definition{Name: "x", Color: labels.ColorSlate, Scope: labels.ScopeDerived}.PrettyRow(nil)
		assert.NotContains(t, bare, "Icon")
		assert.NotContains(t, bare, "Description")
	})
}
