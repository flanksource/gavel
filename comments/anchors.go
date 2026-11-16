package comments

import (
	"fmt"
	"strings"
)

// TextQuote is the selected text with a little of what surrounds it, so a
// host can highlight the exact words inside a patch anchor's lines. It is
// source text, as the file holds it, not the rendered text a reader sees.
type TextQuote struct {
	Prefix string `json:"prefix,omitempty"`
	Exact  string `json:"exact"`
	Suffix string `json:"suffix,omitempty"`
}

// Validate requires the quote to select some text.
func (q TextQuote) Validate() error {
	if q.Exact == "" {
		return fmt.Errorf("%w: quote exact text is required", ErrInvalid)
	}
	return nil
}

// Within confirms the quote's exact text is part of the patch's selected
// lines; CRLF line endings in either compare as LF.
func (q TextQuote) Within(patch PatchAnchor) error {
	if err := q.Validate(); err != nil {
		return err
	}
	lines := strings.ReplaceAll(strings.Join(patch.Lines, "\n"), "\r\n", "\n")
	if !strings.Contains(lines, strings.ReplaceAll(q.Exact, "\r\n", "\n")) {
		return fmt.Errorf("%w: quote %q is not inside lines %d-%d", ErrInvalid, q.Exact, patch.StartLine, patch.EndLine)
	}
	return nil
}

// Region is a box on an image: its origin, width and height are fractions of
// the image's width and height, so it survives any rendered size. The server
// checks the box strictly; a client clamps a drag to the image before saving.
type Region struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// Validate confirms the box is non-empty and lies inside the image.
func (r Region) Validate() error {
	if r.X < 0 || r.Y < 0 || r.W <= 0 || r.H <= 0 || r.X+r.W > 1 || r.Y+r.H > 1 {
		return fmt.Errorf("%w: region %+v must be a non-empty box inside [0,1]", ErrInvalid, r)
	}
	return nil
}
