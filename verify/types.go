package verify

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

type Finding struct {
	File    string `json:"file" yaml:"file"`
	Line    int    `json:"line,omitempty" yaml:"line,omitempty"`
	Message string `json:"message" yaml:"message"`
}

type SectionResult struct {
	Name     string    `json:"name" yaml:"name"`
	Score    int       `json:"score" yaml:"score"`
	Warnings []Finding `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Errors   []Finding `json:"errors,omitempty" yaml:"errors,omitempty"`
}

type VerifyResult struct {
	Score    int             `json:"score" yaml:"score"`
	Sections []SectionResult `json:"sections" yaml:"sections"`
}

func scoreColor(score int) string {
	switch {
	case score >= 80:
		return "text-green-600"
	case score >= 60:
		return "text-yellow-600"
	default:
		return "text-red-600"
	}
}

func (f Finding) location() string {
	if f.Line > 0 {
		return fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	return f.File
}

func (r VerifyResult) Pretty() api.Text {
	text := clicky.Text("Code Review", "font-bold").
		Append(fmt.Sprintf(" — Score: %d/100", r.Score), scoreColor(r.Score))

	for _, s := range r.Sections {
		text = text.NewLine().NewLine().Add(s.Pretty())
	}
	return text
}

func (s SectionResult) Pretty() api.Text {
	text := clicky.Text(fmt.Sprintf("  %s", s.Name), "font-bold").
		Append(fmt.Sprintf(" %d/100", s.Score), scoreColor(s.Score))

	for _, e := range s.Errors {
		text = text.NewLine().
			Append("    ", "").
			Add(icons.Cross.WithStyle("text-red-600")).
			Append(fmt.Sprintf(" %s — %s", e.location(), e.Message), "")
	}

	for _, w := range s.Warnings {
		text = text.NewLine().
			Append("    ", "").
			Add(icons.Warning.WithStyle("text-yellow-600")).
			Append(fmt.Sprintf(" %s — %s", w.location(), w.Message), "")
	}

	return text
}
