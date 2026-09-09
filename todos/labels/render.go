package labels

import (
	"encoding/json"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

// MarshalJSON adds the resolved Iconify name alongside the stored registry key.
// The dashboard cannot resolve a clicky registry key on its own, and definitions
// travel as one small cached payload rather than per todo — so paying one extra
// string here is what lets the browser render the same glyph the terminal does.
func (d Definition) MarshalJSON() ([]byte, error) {
	type definition Definition // shed the method set to avoid recursing
	return json.Marshal(struct {
		definition
		Iconify string `json:"iconify,omitempty"`
	}{definition: definition(d), Iconify: d.IconFor().Iconify})
}

// IconFor returns the definition's registry icon restyled to its own hue, so a
// label's glyph and its chip never disagree. Icon.ANSI() honours Style, so the
// glyph is coloured in the terminal too. A definition with no icon, or one
// naming a key outside the registry, returns the zero Icon.
func (d Definition) IconFor() icons.Icon {
	base, ok := icons.All[Normalize(d.Icon)]
	if !ok {
		return icons.Icon{}
	}
	return icons.Icon{Unicode: base.Unicode, Iconify: base.Iconify, Style: d.Color.TextClass()}
}

// Text renders one resolved label as a coloured, icon-prefixed chip.
//
// This deliberately does not use api.LabelBadge. LabelBadge.ANSI() returns its
// plain String(), so a badge is monochrome in a terminal — colour exists only in
// its HTML(). clicky.Text with tailwind classes goes through api.formatANSI,
// which resolves both the foreground and the background, so one construction
// colours ANSI, HTML and markdown alike.
func (d Definition) Text() api.Text {
	text := clicky.Text("")
	if icon := d.IconFor(); icon.Unicode != "" {
		text = text.Add(icon).Append(" ", "")
	}
	return text.Append(" "+d.Name+" ", d.Color.Classes())
}

// Text joins a set of chips with single spaces.
func (d Definitions) Text() api.Text {
	result := api.Text{}
	for i, def := range d {
		if i > 0 {
			result = result.Append(" ", "")
		}
		result = result.Add(def.Text())
	}
	return result
}

// Pretty renders one definition as its chip.
func (d Definition) Pretty() api.Text { return d.Text() }

// PrettyRow renders one definition as a table row for `gavel todos labels list`.
func (d Definition) PrettyRow(opts interface{}) map[string]api.Text {
	row := map[string]api.Text{
		"Label": d.Text().Styles("order-0"),
		"Color": clicky.Text(string(d.Color), "order-1 "+d.Color.TextClass()),
		"Scope": clicky.Text(string(d.Scope), "order-3 text-muted"),
	}
	if d.Icon != "" {
		row["Icon"] = clicky.Text("").Add(d.IconFor()).Append(" "+d.Icon, "").Styles("order-2")
	}
	if d.Description != "" {
		row["Description"] = clicky.Text(d.Description, "order-5 text-muted")
	}
	return row
}
