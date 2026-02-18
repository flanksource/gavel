package verify

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

type Evidence struct {
	File    string `json:"file" yaml:"file"`
	Line    int    `json:"line,omitempty" yaml:"line,omitempty"`
	Message string `json:"message" yaml:"message"`
}

type CheckResult struct {
	Pass     bool       `json:"pass" yaml:"pass"`
	Evidence []Evidence `json:"evidence,omitempty" yaml:"evidence,omitempty"`
}

type RatingResult struct {
	Score    int        `json:"score" yaml:"score"`
	Findings []Evidence `json:"findings,omitempty" yaml:"findings,omitempty"`
}

type CompletenessResult struct {
	Pass     bool       `json:"pass" yaml:"pass"`
	Summary  string     `json:"summary" yaml:"summary"`
	Evidence []Evidence `json:"evidence,omitempty" yaml:"evidence,omitempty"`
}

type VerifyResult struct {
	Checks       map[string]CheckResult  `json:"checks" yaml:"checks"`
	Ratings      map[string]RatingResult `json:"ratings" yaml:"ratings"`
	Completeness CompletenessResult      `json:"completeness" yaml:"completeness"`
	Score        int                     `json:"score" yaml:"score"`
}

func ratingColor(score int) string {
	switch {
	case score >= 80:
		return "text-green-600"
	case score >= 60:
		return "text-yellow-600"
	default:
		return "text-red-600"
	}
}

func (e Evidence) location() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d", e.File, e.Line)
	}
	return e.File
}

func (r VerifyResult) Pretty() api.Text {
	text := clicky.Text("Code Review", "font-bold").
		Append(fmt.Sprintf(" — Score: %d/100", r.Score), ratingColor(r.Score))

	text = text.NewLine().NewLine().Add(r.prettyChecks())
	text = text.NewLine().NewLine().Add(r.prettyRatings())
	text = text.NewLine().NewLine().Add(r.prettyCompleteness())
	return text
}

func (r VerifyResult) prettyChecks() api.Text {
	text := clicky.Text("Checks", "font-bold")

	passed, failed := 0, 0
	for _, cr := range r.Checks {
		if cr.Pass {
			passed++
		} else {
			failed++
		}
	}
	text = text.Append(fmt.Sprintf(" (%d passed, %d failed)", passed, failed), "")

	byCategory := make(map[string][]string)
	for id := range r.Checks {
		for _, c := range AllChecks {
			if c.ID == id {
				byCategory[c.Category] = append(byCategory[c.Category], id)
				break
			}
		}
	}

	for _, cat := range AllCategories {
		ids := byCategory[cat]
		if len(ids) == 0 {
			continue
		}
		text = text.NewLine().Append(fmt.Sprintf("  %s", cat), "font-bold")
		for _, id := range ids {
			cr := r.Checks[id]
			if cr.Pass {
				text = text.NewLine().Append("    ", "").
					Add(icons.Check.WithStyle("text-green-600")).
					Append(fmt.Sprintf(" %s", id), "")
			} else {
				text = text.NewLine().Append("    ", "").
					Add(icons.Cross.WithStyle("text-red-600")).
					Append(fmt.Sprintf(" %s", id), "")
				for _, e := range cr.Evidence {
					text = text.NewLine().
						Append(fmt.Sprintf("      %s — %s", e.location(), e.Message), "")
				}
			}
		}
	}
	return text
}

func (r VerifyResult) prettyRatings() api.Text {
	text := clicky.Text("Ratings", "font-bold")
	for _, dim := range RatingDimensions {
		rating, ok := r.Ratings[dim]
		if !ok {
			continue
		}
		text = text.NewLine().Append(fmt.Sprintf("  %s ", dim), "font-bold").
			Append(fmt.Sprintf("%d/100", rating.Score), ratingColor(rating.Score))
		for _, f := range rating.Findings {
			text = text.NewLine().
				Append("    ", "").
				Add(icons.Warning.WithStyle("text-yellow-600")).
				Append(fmt.Sprintf(" %s — %s", f.location(), f.Message), "")
		}
	}
	return text
}

func (r VerifyResult) prettyCompleteness() api.Text {
	icon := icons.Check.WithStyle("text-green-600")
	label := "PASS"
	if !r.Completeness.Pass {
		icon = icons.Cross.WithStyle("text-red-600")
		label = "FAIL"
	}
	text := clicky.Text("Completeness ", "font-bold").Add(icon).Append(fmt.Sprintf(" %s", label), "")
	if r.Completeness.Summary != "" {
		text = text.NewLine().Append(fmt.Sprintf("  %s", r.Completeness.Summary), "")
	}
	for _, e := range r.Completeness.Evidence {
		text = text.NewLine().
			Append("  ", "").
			Add(icons.Warning.WithStyle("text-yellow-600")).
			Append(fmt.Sprintf(" %s — %s", e.location(), e.Message), "")
	}
	return text
}
