package labels

import (
	"fmt"

	"github.com/flanksource/clicky/api/icons"
)

// Scope records where a resolved definition came from, so the editor can show
// "inherited from global" and the API can distinguish a stored row from a
// built-in default that has never been written.
type Scope string

const (
	// ScopeWorkspace is a definition stored for one workspace; it shadows a
	// global definition of the same name.
	ScopeWorkspace Scope = "workspace"
	// ScopeGlobal is a definition stored once and applied to every workspace.
	ScopeGlobal Scope = "global"
	// ScopeBuiltin is a well-known default that has never been written to the
	// table. Storing an edit of one inserts a row that shadows it.
	ScopeBuiltin Scope = "builtin"
	// ScopeDerived is the hashed fallback for a label nothing defines. Nothing
	// is stored, so a derived definition never appears in a listing.
	ScopeDerived Scope = "derived"
)

// Definition is one resolved label presentation.
type Definition struct {
	Name        string `json:"name"`
	Color       Color  `json:"color"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	Scope       Scope  `json:"scope"`
	// MatchedKey is set when the definition was found by the label's namespace
	// key rather than its full name ("source" for "source:todo"), so a reader
	// can tell an exact match from an inherited family colour.
	MatchedKey string `json:"matchedKey,omitempty"`
}

// Definitions is a renderable set of label presentations.
type Definitions []Definition

// IsIcon reports whether name is a key in clicky's icon registry. The registry
// is the single source for both the terminal glyph and the Iconify name, so an
// icon that renders in the dashboard always has a terminal counterpart.
func IsIcon(name string) bool {
	if name == "" {
		return false
	}
	_, ok := icons.All[Normalize(name)]
	return ok
}

// IconNames returns every registry key, sorted — the candidate list an editor
// or a CLI error message offers.
func IconNames() []string {
	names := make([]string, 0, len(icons.All))
	for name := range icons.All {
		names = append(names, name)
	}
	sortNames(names)
	return names
}

// Validate rejects a definition that could not resolve or could not render: an
// unnormalized or empty name, a hue outside the palette, or an icon absent from
// the registry. Storage calls it before writing so a bad definition fails at the
// write rather than silently rendering wrong on every read.
func (d Definition) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("label name is required")
	}
	if d.Name != Normalize(d.Name) {
		return fmt.Errorf("label name %q must be lowercase and trimmed", d.Name)
	}
	if _, err := ParseColor(string(d.Color)); err != nil {
		return fmt.Errorf("label %q: %w", d.Name, err)
	}
	if d.Icon != "" && !IsIcon(d.Icon) {
		return fmt.Errorf("label %q: unknown icon %q (not in the clicky icon registry)", d.Name, d.Icon)
	}
	return nil
}

// Names returns the definitions' names in order.
func (d Definitions) Names() []string {
	names := make([]string, len(d))
	for i, def := range d {
		names[i] = def.Name
	}
	return names
}
