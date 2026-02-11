package models

import (
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
	File     string            `json:"file,omitempty"`
	Line     int               `json:"line,omitempty"`
	Column   int               `json:"column,omitempty"`
	Message  *string           `json:"message,omitempty"`
	Source   string            `json:"source,omitempty"`
	Rule     *Rule             `json:"rule,omitempty"`
	Severity ViolationSeverity `yaml:"severity,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
}

func (v Violation) Pretty() api.Text {
	return api.Text{}.Append(v.File).Append(":").Append(v.Line)
}

type AnalysisResult struct {
	Violations []Violation `json:"violations,omitempty"`
	FileCount  int         `json:"file_count,omitempty"`
	RuleCount  int         `json:"rule_count,omitempty"`
}

func StringPtr(s string) *string {
	return &s
}
