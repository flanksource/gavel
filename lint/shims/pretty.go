package shims

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

// Pretty renders one line per shim plus the PATH hint when a shim is shadowed.
func (r InstallReport) Pretty() api.Text {
	t := clicky.Text("Shims: ", "text-muted").Append(r.Dir).NewLine()
	for _, s := range r.Shims {
		t = t.Append(fmt.Sprintf("%-15s", s.Tool), "font-bold").Append(fmt.Sprintf("%-18s", s.Action))
		if s.Real != "" {
			t = t.Append("→ ", "text-muted").Append(s.Real)
		}
		if !s.Active {
			t = t.Space().Append("(not first on PATH)", "text-yellow-600")
		}
		if s.Note != "" {
			t = t.Space().Append(s.Note, "text-yellow-600")
		}
		t = t.NewLine()
	}
	if r.PathHint != "" {
		t = t.Append("Add the shim dir to the front of PATH:", "text-muted").NewLine().Append("  " + r.PathHint).NewLine()
	}
	return t
}
