package parsers

import (
	"encoding/json"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/formatters"
)

// Detail is a test's rich clicky document — the source/normalized input, a live
// per-step trace, or any other Textable content. It is live-mutable during a run
// (Add appends sections to a root api.Text) and serializes to the
// {version:1,node} ClickyDocument JSON consumed by @flanksource/clicky-ui's
// <Clicky data={detail} /> component.
//
// On reload from a snapshot the original api.Text builder tree is not
// reconstructed (snapshots are read-only for display); the parsed ClickyDocument
// is held verbatim and re-emitted unchanged on the next marshal.
type Detail struct {
	root *api.Text
	doc  *formatters.ClickyDocument
}

// NewDetail returns an empty, live-mutable Detail.
func NewDetail() *Detail {
	return &Detail{root: &api.Text{}}
}

// Add appends a Textable section (e.g. an api.Collapsed or api.Code) to the
// document's root and returns the Detail for chaining:
//
//	test.Detail().Add(api.Collapsed{Label: "source", Content: code})
func (d *Detail) Add(child api.Textable) *Detail {
	if d.root == nil {
		d.root = &api.Text{}
	}
	next := d.root.Add(child)
	d.root = &next
	// A live edit supersedes any reloaded document.
	d.doc = nil
	return d
}

// IsEmpty reports whether the Detail carries no content (never added to and not
// reloaded), so callers/marshalers can omit it.
func (d *Detail) IsEmpty() bool {
	if d == nil {
		return true
	}
	if d.doc != nil {
		return false
	}
	return d.root == nil || (d.root.Content == "" && len(d.root.Children) == 0)
}

// Document returns the structured ClickyDocument this Detail serializes to: the
// live root rendered via NewClickyDocument, or the verbatim reloaded document.
// Useful for inspecting the rendered tree (e.g. in tests) without re-parsing the
// JSON.
func (d *Detail) Document() formatters.ClickyDocument {
	if d == nil {
		return formatters.NewClickyDocument(api.Text{})
	}
	if d.doc != nil {
		return *d.doc
	}
	if d.root == nil {
		return formatters.NewClickyDocument(api.Text{})
	}
	return formatters.NewClickyDocument(*d.root)
}

// MarshalJSON emits the structured ClickyDocument: the live root rendered via
// NewClickyDocument, or the verbatim reloaded document.
func (d Detail) MarshalJSON() ([]byte, error) {
	if d.doc != nil {
		return json.Marshal(d.doc)
	}
	if d.root == nil {
		return json.Marshal(formatters.NewClickyDocument(api.Text{}))
	}
	return json.Marshal(formatters.NewClickyDocument(*d.root))
}

// UnmarshalJSON parses the ClickyDocument verbatim; it is re-emitted unchanged.
func (d *Detail) UnmarshalJSON(b []byte) error {
	var doc formatters.ClickyDocument
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	d.doc = &doc
	d.root = nil
	return nil
}

// Detail lazily allocates and returns the test's rich Detail document, so
// callers can append sections during a run:
//
//	test.Detail().Add(api.Collapsed{Label: "source", Content: code})
func (t *Test) Detail() *Detail {
	if t.DetailDoc == nil {
		t.DetailDoc = NewDetail()
	}
	return t.DetailDoc
}
