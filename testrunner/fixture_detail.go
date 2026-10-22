package testrunner

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// fixtureDetail is a fixture test's detail: fixtures.ResultDetail, laid over
// the progress node's executionDetail when the result lands on one. Its JSON is
// the whole record the UI and reports read. Its terminal rendering keeps only
// what the failure line and the stdout/stderr blocks do not already show; the
// step, outputs, exports and each attempt's embedded result would repeat them.
type fixtureDetail json.RawMessage

func (d fixtureDetail) MarshalJSON() ([]byte, error) {
	return json.RawMessage(d).MarshalJSON()
}

type fixtureDetailView struct {
	CELTrace string                `json:"cel_trace"`
	Wait     *fixtureWaitView      `json:"wait"`
	Attempts []fixtureAttemptView  `json:"attempts"`
	Traces   []parsers.TraceResult `json:"traces"`
}

// fixtureWaitView is the part of a fixtures.WaitReport the terminal shows.
type fixtureWaitView struct {
	Outcome    string `json:"outcome"`
	DurationMs int64  `json:"durationMs"`
	Error      string `json:"error"`
	Cause      string `json:"cause"`
	Trace      string `json:"trace"`
}

type fixtureAttemptView struct {
	Attempt    int    `json:"attempt"`
	Status     string `json:"status"`
	DurationMs int64  `json:"durationMs"`
	Error      string `json:"error"`
}

func (d fixtureDetail) Pretty() api.Text {
	var view fixtureDetailView
	if err := json.Unmarshal(d, &view); err != nil {
		return clicky.Text(fmt.Sprintf("fixture detail is not readable: %v", err), "text-red-500")
	}
	s := clicky.Text("")
	if view.CELTrace != "" {
		s = s.NewLine().Append(view.CELTrace, "font-mono text-red-500")
	}
	if w := view.Wait; w != nil && w.Outcome != fixtures.WaitSatisfied {
		s = s.NewLine().Append(fmt.Sprintf("wait %s after %s", w.Outcome, millis(w.DurationMs)), "text-amber-600")
		if reason := firstNonEmpty(w.Cause, w.Error); reason != "" {
			s = s.Append(": "+firstLine(reason), "text-muted")
		}
		if w.Trace != "" {
			s = s.NewLine().Append(w.Trace, "font-mono text-red-500")
		}
	}
	if len(view.Attempts) > 1 {
		for _, a := range view.Attempts {
			s = s.NewLine().Append(fmt.Sprintf("attempt %d %s (%s)", a.Attempt, a.Status, millis(a.DurationMs)), "text-muted")
			if a.Error != "" {
				s = s.Append(": "+firstLine(a.Error), "text-muted")
			}
		}
	}
	return appendTraceLines(s, view.Traces)
}

// Pretty renders a progress node that no result landed on: its error is
// already the test's message, so only its trace windows are left to show.
func (d executionDetail) Pretty() api.Text {
	return appendTraceLines(clicky.Text(""), d.Traces)
}

func appendTraceLines(s api.Text, traces []parsers.TraceResult) api.Text {
	for _, trace := range traces {
		s = s.NewLine().Append(fmt.Sprintf("%s: %d records", trace.Kind, trace.Count), "text-muted")
		if summary := formatTraceSummary(trace.Summary); summary != "" {
			s = s.Append(" "+summary, "text-muted")
		}
		if trace.Error != "" {
			s = s.Append(" "+firstLine(trace.Error), "text-red-500")
		}
	}
	return s
}

// formatTraceSummary keeps a trace summary's scalar figures; nested values,
// such as pg_proxy's per-statement list, stay in the JSON detail.
func formatTraceSummary(summary map[string]any) string {
	keys := make([]string, 0, len(summary))
	for key := range summary {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var parts []string
	for _, key := range keys {
		switch v := summary[key].(type) {
		case float64:
			parts = append(parts, key+"="+strconv.FormatFloat(v, 'g', 6, 64))
		case int, int64, bool:
			parts = append(parts, fmt.Sprintf("%s=%v", key, v))
		case string:
			if v != "" && !strings.Contains(v, "\n") {
				parts = append(parts, key+"="+v)
			}
		}
	}
	return strings.Join(parts, " ")
}

func millis(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return line
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
