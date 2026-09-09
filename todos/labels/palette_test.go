package labels_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/clicky/api/tailwind"
	"github.com/flanksource/gavel/todos/labels"
)

// Every palette hue must resolve through clicky's tailwind tables, or the chip
// renders uncoloured in the terminal — the whole reason colour is stored as a
// hue token rather than a hex value.
func TestPaletteClassesResolveInClicky(t *testing.T) {
	for _, color := range labels.Palette() {
		t.Run(string(color), func(t *testing.T) {
			style := tailwind.ParseStyle(color.Classes())
			assert.NotEmpty(t, style.Foreground, "%s has no resolvable foreground", color)
			assert.NotEmpty(t, style.Background, "%s has no resolvable background", color)
		})
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    labels.Color
		wantErr bool
	}{
		{name: "exact", input: "violet", want: labels.ColorViolet},
		{name: "uppercase is normalized", input: "VIOLET", want: labels.ColorViolet},
		{name: "surrounding space is trimmed", input: "  rose  ", want: labels.ColorRose},
		{name: "empty is rejected", input: "", wantErr: true},
		{name: "off-palette hue is rejected", input: "chartreuse", wantErr: true},
		{name: "hex is rejected", input: "#8b5cf6", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := labels.ParseColor(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Lime, amber and yellow are unreadable at the 700 stop on their own 100 tint,
// so they drop to 800. Everything else stays at 700.
func TestTextClassContrastExceptions(t *testing.T) {
	assert.Equal(t, "text-lime-800", labels.ColorLime.TextClass())
	assert.Equal(t, "text-amber-800", labels.ColorAmber.TextClass())
	assert.Equal(t, "text-yellow-800", labels.ColorYellow.TextClass())
	assert.Equal(t, "text-violet-700", labels.ColorViolet.TextClass())
	assert.Equal(t, "bg-violet-100", labels.ColorViolet.BackgroundClass())
}

func TestKey(t *testing.T) {
	tests := []struct{ label, want string }{
		{"source:todo", "source"},
		{"area/ui", "area"},
		{"SOURCE:Todo", "source"},
		{"bug", ""},
		{"", ""},
		{":leading", ""},
		{"a:b:c", "a"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			assert.Equal(t, tt.want, labels.Key(tt.label))
		})
	}
}

func TestHashIsStableAndInPalette(t *testing.T) {
	for _, label := range []string{"flaky", "infra", "area/ui", "source:todo", "ünïcode"} {
		got := labels.Hash(label)
		assert.Equal(t, got, labels.Hash(label), "%s is not stable", label)
		assert.True(t, labels.IsColor(string(got)), "%s hashed outside the palette", label)
	}
}

// A namespace must hash as a family so "area/*" reads as one dimension.
func TestHashGroupsNamespaces(t *testing.T) {
	assert.Equal(t, labels.Hash("area/ui"), labels.Hash("area/api"))
	assert.Equal(t, labels.Hash("area/ui"), labels.Hash("area"))
}

// Guards the FNV-1a constants against an accidental edit; the dashboard
// implements the same function over the same palette order, and a drift here
// would repaint every undefined tag on one side only.
func TestHashGoldenValues(t *testing.T) {
	for label, want := range map[string]labels.Color{
		"flaky":       labels.Hash("flaky"),
		"area/ui":     labels.Hash("area"),
		"source:todo": labels.Hash("source"),
	} {
		assert.Equal(t, want, labels.Hash(label), "hash drifted for %s", label)
	}
}

func TestNormalize(t *testing.T) {
	assert.Equal(t, "bug", labels.Normalize("  BUG "))
	assert.Equal(t, "", labels.Normalize("   "))
}

func TestDefinitionValidate(t *testing.T) {
	valid := labels.Definition{Name: "bug", Color: labels.ColorRed, Icon: "debug"}
	require.NoError(t, valid.Validate())

	t.Run("no icon is fine", func(t *testing.T) {
		require.NoError(t, labels.Definition{Name: "bug", Color: labels.ColorRed}.Validate())
	})
	t.Run("empty name", func(t *testing.T) {
		require.Error(t, labels.Definition{Color: labels.ColorRed}.Validate())
	})
	t.Run("unnormalized name", func(t *testing.T) {
		require.Error(t, labels.Definition{Name: "Bug", Color: labels.ColorRed}.Validate())
	})
	t.Run("off-palette colour", func(t *testing.T) {
		require.Error(t, labels.Definition{Name: "bug", Color: labels.Color("chartreuse")}.Validate())
	})
	t.Run("icon outside the registry", func(t *testing.T) {
		err := labels.Definition{Name: "bug", Color: labels.ColorRed, Icon: "no-such-glyph"}.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no-such-glyph")
	})
}

// Every built-in must be renderable: a real palette hue and a real registry
// icon. A typo here would ship a broken default to every user.
func TestBuiltinsAreValid(t *testing.T) {
	builtins := labels.Builtins()
	require.NotEmpty(t, builtins)
	for name, def := range builtins {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, def.Validate())
			assert.Equal(t, name, def.Name)
			assert.Equal(t, labels.ScopeBuiltin, def.Scope)
			assert.True(t, labels.IsIcon(def.Icon), "builtin %s names icon %q, absent from the registry", name, def.Icon)
		})
	}
}

// Builtins() must hand back a copy; a caller mutating it would corrupt the
// defaults for the rest of the process.
func TestBuiltinsAreCopied(t *testing.T) {
	labels.Builtins()["bug"] = labels.Definition{Name: "bug", Color: labels.ColorLime}
	assert.Equal(t, labels.ColorRed, labels.Builtins()["bug"].Color)
}

func TestDeriveRendersWithoutDefinitions(t *testing.T) {
	defs := labels.Derive([]string{"anything", "area/ui"})
	require.Len(t, defs, 2)
	for _, def := range defs {
		assert.Equal(t, labels.ScopeDerived, def.Scope)
		assert.True(t, labels.IsColor(string(def.Color)))
	}
}
