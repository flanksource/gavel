// Package labels owns the presentation contract for TODO labels: the palette a
// label is coloured from, the definition that stores that choice, and the
// resolution chain that turns a raw label string into something renderable.
//
// It is a leaf package — it imports clicky and nothing else from gavel — so the
// storage layer (todos/native), the domain type (todos/types) and the provider
// contract (todos) can all depend on it without a cycle.
package labels

import (
	"fmt"
	"sort"
	"strings"
)

// Color is one of the palette hues. A definition stores the hue token rather
// than a hex value so both renderers can resolve it: the terminal needs a
// tailwind class (clicky's formatANSI resolves "bg-violet-100 text-violet-700"
// to termenv foreground AND background), and the dashboard needs a class
// literal Tailwind's JIT has actually seen in source. A hex column would serve
// neither without a second lookup table on each side.
type Color string

const (
	ColorSlate   Color = "slate"
	ColorRed     Color = "red"
	ColorOrange  Color = "orange"
	ColorAmber   Color = "amber"
	ColorYellow  Color = "yellow"
	ColorLime    Color = "lime"
	ColorGreen   Color = "green"
	ColorEmerald Color = "emerald"
	ColorTeal    Color = "teal"
	ColorCyan    Color = "cyan"
	ColorSky     Color = "sky"
	ColorBlue    Color = "blue"
	ColorIndigo  Color = "indigo"
	ColorViolet  Color = "violet"
	ColorPurple  Color = "purple"
	ColorFuchsia Color = "fuchsia"
	ColorPink    Color = "pink"
	ColorRose    Color = "rose"
)

// palette is the ordered hue list. The order is load-bearing: Hash indexes into
// it, so reordering repaints every undefined label. It is also the exact set the
// todo_labels_color_check constraint enforces, and the set the dashboard's
// TAG_PALETTE mirrors.
var palette = []Color{
	ColorSlate, ColorRed, ColorOrange, ColorAmber, ColorYellow, ColorLime,
	ColorGreen, ColorEmerald, ColorTeal, ColorCyan, ColorSky, ColorBlue,
	ColorIndigo, ColorViolet, ColorPurple, ColorFuchsia, ColorPink, ColorRose,
}

// Palette returns the ordered hue list.
func Palette() []Color {
	out := make([]Color, len(palette))
	copy(out, palette)
	return out
}

// PaletteStrings returns the palette as plain strings, for the SQL check
// constraint and for JSON schema enums.
func PaletteStrings() []string {
	out := make([]string, len(palette))
	for i, c := range palette {
		out[i] = string(c)
	}
	return out
}

var paletteIndex = func() map[Color]struct{} {
	index := make(map[Color]struct{}, len(palette))
	for _, c := range palette {
		index[c] = struct{}{}
	}
	return index
}()

// IsColor reports whether value names a palette hue.
func IsColor(value string) bool {
	_, ok := paletteIndex[Color(value)]
	return ok
}

// ParseColor validates a stored or user-supplied hue. It fails loudly rather
// than falling back to a default: a colour that silently changed is a worse
// outcome than a rejected write.
func ParseColor(value string) (Color, error) {
	normalized := Normalize(value)
	if normalized == "" {
		return "", fmt.Errorf("label color is required (one of: %s)", strings.Join(PaletteStrings(), ", "))
	}
	if !IsColor(normalized) {
		return "", fmt.Errorf("unknown label color %q (want one of: %s)", value, strings.Join(PaletteStrings(), ", "))
	}
	return Color(normalized), nil
}

// darkText holds the hues whose 700 stop is too light to read on their own 100
// tint; they use 800 instead. This mirrors the dashboard's palette exactly.
var darkText = map[Color]struct{}{
	ColorLime:   {},
	ColorAmber:  {},
	ColorYellow: {},
}

// BackgroundClass is the chip's tint, e.g. "bg-violet-100".
func (c Color) BackgroundClass() string { return "bg-" + string(c) + "-100" }

// TextClass is the chip's label colour, e.g. "text-violet-700". Lime, amber and
// yellow drop to the 800 stop so they stay legible on their own tint.
func (c Color) TextClass() string {
	shade := "700"
	if _, dark := darkText[c]; dark {
		shade = "800"
	}
	return "text-" + string(c) + "-" + shade
}

// Classes is the full chip style. One string colours the terminal (clicky
// resolves both halves to termenv colours), the HTML report and the dashboard.
func (c Color) Classes() string { return c.BackgroundClass() + " " + c.TextClass() }

func (c Color) String() string { return string(c) }

// Normalize is the single label-token normalization: lowercase and trim. It is
// what the todo_labels name check enforces in SQL and what native's
// normalizeStrings applies to todo_issues.labels, so a definition written here
// always matches a label stored there.
func Normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// Key splits a namespaced label into its key: "source:todo" -> "source",
// "area/ui" -> "area", "bug" -> "". The first ':' or '/' wins. Namespaced
// labels resolve and colour as a family, so every "area/*" reads as one
// dimension instead of eighteen unrelated hues.
func Key(label string) string {
	normalized := Normalize(label)
	if idx := strings.IndexAny(normalized, ":/"); idx > 0 {
		return normalized[:idx]
	}
	return ""
}

// Hash picks the deterministic hue for a label with no stored definition, so an
// unconfigured backlog is still visually separable. It hashes the label's Key
// when it has one, so namespace members share a hue.
func Hash(label string) Color {
	seed := Key(label)
	if seed == "" {
		seed = Normalize(label)
	}
	return palette[fnv1a32(seed)%uint32(len(palette))]
}

// fnv1a32 is FNV-1a. The dashboard implements the same function over the same
// palette order so a client that has not yet loaded definitions derives exactly
// the colour the server would.
func fnv1a32(s string) uint32 {
	const (
		offset uint32 = 2166136261
		prime  uint32 = 16777619
	)
	hash := offset
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= prime
	}
	return hash
}

// sortNames orders label names for stable rendering and stable test output.
func sortNames(names []string) {
	sort.Strings(names)
}
