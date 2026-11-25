package yamldiag

import (
	"fmt"
	"strings"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/models"
)

type Error struct {
	Context    string
	Violations []models.Violation
	Cause      error
}

func (e *Error) Unwrap() error { return e.Cause }

func (e *Error) Error() string { return e.Pretty().String() }

func (e *Error) Pretty() api.Text {
	out := api.Text{}
	for i, v := range e.Violations {
		if i > 0 {
			out = out.NewLine().NewLine()
		}
		out = out.Append(fmt.Sprintf("%s:%d:%d: %s: %s", v.File, v.Line, v.Column, e.Context, strings.TrimSpace(*v.Message)))
		if v.Excerpt != nil {
			out = out.NewLine().Add(v.Excerpt.Pretty())
		}
	}
	return out
}
