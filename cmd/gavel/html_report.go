package main

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func init() {
	formatters.RegisterFormatter("html", func(data interface{}, options formatters.FormatOptions) (string, error) {
		if items, ok := data.([]interface{}); ok && len(items) == 1 {
			data = items[0]
		}
		switch v := data.(type) {
		case testui.Snapshot:
			return renderGavelHTMLReport(v), nil
		case *testui.Snapshot:
			if v == nil {
				return "", nil
			}
			return renderGavelHTMLReport(*v), nil
		default:
			return formatters.NewHTMLFormatter().Format(data, options)
		}
	})
}

type htmlPackageSummary struct {
	Key       string
	Package   string
	Framework string
	Passed    int
	Failed    int
	Skipped   int
	Pending   int
	Duration  time.Duration
	Failures  []parsers.Test
}

func renderGavelHTMLReport(s testui.Snapshot) string {
	pkgs := collectHTMLPackageSummaries(s.Tests)
	lintFailures := collectHTMLLintFailures(s.Lint)

	var totalPassed, totalFailed, totalSkipped, totalPending int
	var totalDuration time.Duration
	for _, pkg := range pkgs {
		totalPassed += pkg.Passed
		totalFailed += pkg.Failed
		totalSkipped += pkg.Skipped
		totalPending += pkg.Pending
		totalDuration += pkg.Duration
	}

	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>Gavel results</title>")
	b.WriteString(`<style>
		body{font-family:Inter,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;margin:2rem;color:#172033;background:#f7f8fb}
		h1{margin-bottom:.25rem}.muted{color:#667085}.cards{display:flex;gap:1rem;flex-wrap:wrap;margin:1.5rem 0}.card{background:#fff;border:1px solid #e5e7eb;border-radius:12px;padding:1rem 1.25rem;box-shadow:0 1px 2px rgba(16,24,40,.04);min-width:8rem}.metric{font-size:1.6rem;font-weight:700}.ok{color:#15803d}.fail{color:#b91c1c}.skip{color:#a16207}.pending{color:#475569}
		table{width:100%;border-collapse:separate;border-spacing:0;background:#fff;border:1px solid #e5e7eb;border-radius:12px;overflow:hidden;box-shadow:0 1px 2px rgba(16,24,40,.04)}th,td{padding:.75rem .9rem;border-bottom:1px solid #e5e7eb;text-align:left;vertical-align:top}th{background:#f2f4f7;font-size:.8rem;text-transform:uppercase;letter-spacing:.04em;color:#475467}tr:last-child td{border-bottom:0}.num{text-align:right;font-variant-numeric:tabular-nums}.status{font-weight:700}.package{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}
		details summary{cursor:pointer;list-style:none}details summary::-webkit-details-marker{display:none}.trace{margin:.75rem 0 0 0;background:#101828;color:#f9fafb;border-radius:8px;padding:1rem;overflow:auto;white-space:pre-wrap;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.85rem}.failure-title{font-weight:700;margin:.5rem 0 .25rem}.lint{margin-top:2rem}.badge{display:inline-block;border-radius:999px;padding:.15rem .5rem;font-size:.75rem;font-weight:700;background:#eef2ff;color:#3730a3}
	</style>`)
	b.WriteString("</head><body>")
	b.WriteString("<h1>Gavel results</h1>")
	if s.Git != nil && s.Git.SHA != "" {
		b.WriteString(fmt.Sprintf("<div class=\"muted\">%s @ %s</div>", esc(s.Git.Repo), esc(shortSHA(s.Git.SHA))))
	}
	b.WriteString("<div class=\"cards\">")
	metric(&b, "Passed", totalPassed, "ok")
	metric(&b, "Failed", totalFailed, "fail")
	metric(&b, "Skipped", totalSkipped, "skip")
	metric(&b, "Pending", totalPending, "pending")
	b.WriteString(fmt.Sprintf("<div class=\"card\"><div class=\"muted\">Duration</div><div class=\"metric\">%s</div></div>", esc(formatHTMLDuration(totalDuration))))
	b.WriteString("</div>")

	b.WriteString("<h2>Packages</h2><table><thead><tr><th>Status</th><th>Package</th><th>Framework</th><th class=\"num\">Pass</th><th class=\"num\">Fail</th><th class=\"num\">Skip</th><th class=\"num\">Duration</th></tr></thead><tbody>")
	for _, pkg := range pkgs {
		status, cls := packageStatus(pkg)
		b.WriteString("<tr>")
		b.WriteString(fmt.Sprintf("<td class=\"status %s\">%s</td>", cls, status))
		b.WriteString("<td class=\"package\">")
		if len(pkg.Failures) > 0 {
			b.WriteString("<details><summary>")
			b.WriteString(esc(pkg.Package))
			b.WriteString(" <span class=\"badge\">details</span></summary>")
			for _, failure := range pkg.Failures {
				b.WriteString("<div class=\"failure-title\">")
				b.WriteString(esc(testDisplayName(failure)))
				b.WriteString("</div><pre class=\"trace\">")
				b.WriteString(esc(testTrace(failure)))
				b.WriteString("</pre>")
			}
			b.WriteString("</details>")
		} else {
			b.WriteString(esc(pkg.Package))
		}
		b.WriteString("</td>")
		b.WriteString(fmt.Sprintf("<td>%s</td><td class=\"num ok\">%d</td><td class=\"num fail\">%d</td><td class=\"num skip\">%d</td><td class=\"num\">%s</td>", esc(pkg.Framework), pkg.Passed, pkg.Failed, pkg.Skipped, esc(formatHTMLDuration(pkg.Duration))))
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table>")

	if len(lintFailures) > 0 {
		b.WriteString("<div class=\"lint\"><h2>Lint</h2><table><thead><tr><th>Linter</th><th>Rule</th><th>File</th><th>Message</th></tr></thead><tbody>")
		for _, row := range lintFailures {
			b.WriteString("<tr><td>" + esc(row[0]) + "</td><td>" + esc(row[1]) + "</td><td class=\"package\">" + esc(row[2]) + "</td><td>" + esc(row[3]) + "</td></tr>")
		}
		b.WriteString("</tbody></table></div>")
	}
	b.WriteString("</body></html>")
	return b.String()
}

func collectHTMLPackageSummaries(tests []parsers.Test) []htmlPackageSummary {
	m := make(map[string]*htmlPackageSummary)
	var walk func(t parsers.Test, pkg, fw string)
	walk = func(t parsers.Test, pkg, fw string) {
		if t.PackagePath != "" {
			pkg = t.PackagePath
		}
		if t.Framework != "" {
			fw = string(t.Framework)
		}
		if len(t.Children) > 0 {
			for _, child := range t.Children {
				walk(child, pkg, fw)
			}
			return
		}
		if pkg == "" {
			pkg = t.Package
		}
		if pkg == "" {
			pkg = "./"
		}
		key := fw + "\x00" + pkg
		p := m[key]
		if p == nil {
			p = &htmlPackageSummary{Key: key, Package: pkg, Framework: fw}
			m[key] = p
		}
		if t.Passed {
			p.Passed++
		}
		if t.Failed || t.TimedOut {
			p.Failed++
			p.Failures = append(p.Failures, t)
		}
		if t.Skipped {
			p.Skipped++
		}
		if t.Pending {
			p.Pending++
		}
		p.Duration += t.Duration
	}
	for _, t := range tests {
		walk(t, "", "")
	}
	out := make([]htmlPackageSummary, 0, len(m))
	for _, p := range m {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func collectHTMLLintFailures(results []*linters.LinterResult) [][4]string {
	var rows [][4]string
	for _, r := range results {
		if r == nil || r.Skipped {
			continue
		}
		for _, v := range r.Violations {
			rule := ""
			if v.Rule != nil {
				rule = v.Rule.Method
			}
			msg := ""
			if v.Message != nil {
				msg = *v.Message
			}
			loc := v.File
			if v.Line > 0 {
				loc = fmt.Sprintf("%s:%d", loc, v.Line)
			}
			rows = append(rows, [4]string{r.Linter, rule, loc, msg})
		}
	}
	return rows
}

func metric(b *strings.Builder, label string, value int, cls string) {
	b.WriteString(fmt.Sprintf("<div class=\"card\"><div class=\"muted\">%s</div><div class=\"metric %s\">%d</div></div>", label, cls, value))
}

func packageStatus(pkg htmlPackageSummary) (string, string) {
	if pkg.Failed > 0 {
		return "✗", "fail"
	}
	if pkg.Pending > 0 {
		return "⏳", "pending"
	}
	if pkg.Skipped > 0 && pkg.Passed == 0 {
		return "⊘", "skip"
	}
	return "✓", "ok"
}

func testDisplayName(t parsers.Test) string {
	parts := append([]string{}, t.Suite...)
	if t.Name != "" {
		parts = append(parts, t.Name)
	}
	if len(parts) == 0 {
		return "failure"
	}
	return strings.Join(parts, " / ")
}

func testTrace(t parsers.Test) string {
	var parts []string
	if t.File != "" {
		loc := t.File
		if t.Line > 0 {
			loc = fmt.Sprintf("%s:%d", loc, t.Line)
		}
		parts = append(parts, loc)
	}
	if t.Message != "" {
		parts = append(parts, t.Message)
	}
	if t.Stderr != "" {
		parts = append(parts, "STDERR:\n"+t.Stderr)
	}
	if t.Stdout != "" {
		parts = append(parts, "STDOUT:\n"+t.Stdout)
	}
	for _, a := range t.Attempts {
		if a.StackTrace != "" {
			parts = append(parts, "STACK:\n"+a.StackTrace)
		}
	}
	if len(parts) == 0 {
		return "No failure trace captured. See gavel.log for raw runner output."
	}
	return strings.Join(parts, "\n\n")
}

func esc(s string) string { return html.EscapeString(s) }

func shortSHA(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

func formatHTMLDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	if d >= time.Minute {
		return d.Round(time.Second).String()
	}
	return d.Round(time.Millisecond).String()
}
