package models

import (
	"fmt"
	"time"

	"github.com/flanksource/clicky/api"
)

type ViolationSeverity string

const (
	SeverityInfo    ViolationSeverity = "info"
	SeverityWarning ViolationSeverity = "warning"
	SeverityError   ViolationSeverity = "error"
)

type Violation struct {
	File             string            `json:"file,omitempty"`
	Line             int               `json:"line,omitempty"`
	Column           int               `json:"column,omitempty"`
	EndLine          int               `json:"end_line,omitempty"`
	EndColumn        int               `json:"end_column,omitempty"`
	Message          *string           `json:"message,omitempty"`
	Source           string            `json:"source,omitempty"`
	Rule             *Rule             `json:"rule,omitempty"`
	Severity         ViolationSeverity `json:"severity,omitempty" yaml:"severity,omitempty"`
	Fixable          bool              `json:"fixable,omitempty"`
	FixApplicability string            `json:"fix_applicability,omitempty"`
	Code             *string           `json:"code,omitempty"`
	Anchor           *ViolationAnchor  `json:"anchor,omitempty"`
	// Locations lists every place one fix resolves, File and Line naming the
	// first; a violation found at one place leaves it empty. The fingerprint
	// ignores it, so the message must not name a single place.
	Locations []ViolationLocation `json:"locations,omitempty"`
	// Fingerprint identifies the violation across runs without its line; see
	// linters.AssignFingerprints.
	Fingerprint string `json:"fingerprint,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
}

// ViolationLocation is one place a violation applies, Message saying what is
// there (e.g. the statement a plan finding came from).
type ViolationLocation struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message,omitempty"`
}

func (l ViolationLocation) String() string {
	if l.Line > 0 {
		return fmt.Sprintf("%s:%d", l.File, l.Line)
	}
	return l.File
}

// ViolationAnchor names what a violation is about independently of where it
// sits: a uir symbol id, the declaration witness of a syntax-only finding, or
// the heading path of a markdown section.
type ViolationAnchor struct {
	Symbol      string `json:"symbol,omitempty"`
	Declaration string `json:"declaration,omitempty"`
	Section     string `json:"section,omitempty"`
}

func (v Violation) Pretty() api.Text {
	t := api.Text{}.Append(v.File).Append(":").Append(v.Line)
	if len(v.Locations) > 1 {
		t = t.Append(fmt.Sprintf(" (%d locations)", len(v.Locations)), "text-muted")
	}
	return t
}

type AnalysisResult struct {
	Violations []Violation `json:"violations,omitempty"`
	FileCount  int         `json:"file_count,omitempty"`
	RuleCount  int         `json:"rule_count,omitempty"`
}

func StringPtr(s string) *string {
	return &s
}
