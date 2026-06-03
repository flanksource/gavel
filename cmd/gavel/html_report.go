package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky/api"
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

	body := []api.Textable{el("h1", nil, txt("Gavel results"))}
	if s.Git != nil && s.Git.SHA != "" {
		body = append(body, el("div", attrs("class", "muted"), txt(s.Git.Repo+" @ "+shortSHA(s.Git.SHA))))
	}

	body = append(body, el("div", attrs("class", "cards"),
		metricCard("Passed", totalPassed, "ok"),
		metricCard("Failed", totalFailed, "fail"),
		metricCard("Skipped", totalSkipped, "skip"),
		metricCard("Pending", totalPending, "pending"),
		el("div", attrs("class", "card"),
			el("div", attrs("class", "muted"), txt("Duration")),
			el("div", attrs("class", "metric"), txt(formatHTMLDuration(totalDuration))),
		),
	))

	body = append(body,
		el("h2", nil, txt("Packages")),
		el("table", nil,
			el("thead", nil,
				el("tr", nil,
					el("th", nil, txt("Status")),
					el("th", nil, txt("Package")),
					el("th", nil, txt("Framework")),
					el("th", attrs("class", "num"), txt("Pass")),
					el("th", attrs("class", "num"), txt("Fail")),
					el("th", attrs("class", "num"), txt("Skip")),
					el("th", attrs("class", "num"), txt("Duration")),
				),
			),
			el("tbody", nil, packageRows(pkgs)...),
		),
	)

	if len(lintFailures) > 0 {
		body = append(body, el("div", attrs("class", "lint"),
			el("h2", nil, txt("Lint")),
			el("table", nil,
				el("thead", nil,
					el("tr", nil,
						el("th", nil, txt("Linter")),
						el("th", nil, txt("Rule")),
						el("th", nil, txt("File")),
						el("th", nil, txt("Message")),
					),
				),
				el("tbody", nil, lintRows(lintFailures)...),
			),
		))
	}

	return renderHTML(
		raw("<!doctype html>"),
		el("html", nil,
			el("head", nil,
				el("meta", attrs("charset", "utf-8")),
				el("title", nil, txt("Gavel results")),
				el("style", nil, raw(reportCSS)),
			),
			el("body", nil, body...),
		),
	)
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

const reportCSS = `
	body{font-family:Inter,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;margin:2rem;color:#172033;background:#f7f8fb}
	h1{margin-bottom:.25rem}.muted{color:#667085}.cards{display:flex;gap:1rem;flex-wrap:wrap;margin:1.5rem 0}.card{background:#fff;border:1px solid #e5e7eb;border-radius:12px;padding:1rem 1.25rem;box-shadow:0 1px 2px rgba(16,24,40,.04);min-width:8rem}.metric{font-size:1.6rem;font-weight:700}.ok{color:#15803d}.fail{color:#b91c1c}.skip{color:#a16207}.pending{color:#475569}
	table{width:100%;border-collapse:separate;border-spacing:0;background:#fff;border:1px solid #e5e7eb;border-radius:12px;overflow:hidden;box-shadow:0 1px 2px rgba(16,24,40,.04)}th,td{padding:.75rem .9rem;border-bottom:1px solid #e5e7eb;text-align:left;vertical-align:top}th{background:#f2f4f7;font-size:.8rem;text-transform:uppercase;letter-spacing:.04em;color:#475467}tr:last-child td{border-bottom:0}.num{text-align:right;font-variant-numeric:tabular-nums}.status{font-weight:700}.package{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}
	details summary{cursor:pointer;list-style:none}details summary::-webkit-details-marker{display:none}.trace{margin:.75rem 0 0 0;background:#101828;color:#f9fafb;border-radius:8px;padding:1rem;overflow:auto;white-space:pre-wrap;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.85rem}.failure-title{font-weight:700;margin:.5rem 0 .25rem}.lint{margin-top:2rem}.badge{display:inline-block;border-radius:999px;padding:.15rem .5rem;font-size:.75rem;font-weight:700;background:#eef2ff;color:#3730a3}
`

func packageRows(pkgs []htmlPackageSummary) []api.Textable {
	rows := make([]api.Textable, 0, len(pkgs))
	for _, pkg := range pkgs {
		status, cls := packageStatus(pkg)
		rows = append(rows, el("tr", nil,
			el("td", attrs("class", "status "+cls), txt(status)),
			el("td", attrs("class", "package"), packageCell(pkg)),
			el("td", nil, txt(pkg.Framework)),
			el("td", attrs("class", "num ok"), txt(fmt.Sprintf("%d", pkg.Passed))),
			el("td", attrs("class", "num fail"), txt(fmt.Sprintf("%d", pkg.Failed))),
			el("td", attrs("class", "num skip"), txt(fmt.Sprintf("%d", pkg.Skipped))),
			el("td", attrs("class", "num"), txt(formatHTMLDuration(pkg.Duration))),
		))
	}
	return rows
}

func packageCell(pkg htmlPackageSummary) api.Textable {
	if len(pkg.Failures) == 0 {
		return txt(pkg.Package)
	}
	content := []api.Textable{el("summary", nil, txt(pkg.Package), txt(" "), el("span", attrs("class", "badge"), txt("details")))}
	for _, failure := range pkg.Failures {
		content = append(content,
			el("div", attrs("class", "failure-title"), txt(testDisplayName(failure))),
			el("pre", attrs("class", "trace"), txt(testTrace(failure))),
		)
	}
	return el("details", nil, content...)
}

func lintRows(rows [][4]string) []api.Textable {
	out := make([]api.Textable, 0, len(rows))
	for _, row := range rows {
		out = append(out, el("tr", nil,
			el("td", nil, txt(row[0])),
			el("td", nil, txt(row[1])),
			el("td", attrs("class", "package"), txt(row[2])),
			el("td", nil, txt(row[3])),
		))
	}
	return out
}

func metricCard(label string, value int, cls string) api.Textable {
	return el("div", attrs("class", "card"),
		el("div", attrs("class", "muted"), txt(label)),
		el("div", attrs("class", "metric "+cls), txt(fmt.Sprintf("%d", value))),
	)
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

func renderHTML(elements ...api.Textable) string {
	var b strings.Builder
	for _, element := range elements {
		b.WriteString(element.HTML())
	}
	// api.HtmlElement currently renders an extra space for tags with no
	// attributes (for example, <h2 >). Normalize it so generated reports remain
	// clean and existing fixture-free checks keep matching exact headings.
	return strings.ReplaceAll(b.String(), " >", ">")
}

func el(tag string, attributes map[string]string, children ...api.Textable) api.HtmlElement {
	return api.HtmlElement{Tag: tag, Attributes: attributes, Content: renderHTML(children...), Fallback: api.Text{Content: renderPlain(children...)}}
}

func raw(content string) api.HtmlElement {
	return api.HtmlElement{Tag: "", Content: content, Fallback: api.Text{Content: content}}
}

func txt(content string) api.Text {
	return api.Text{Content: content}
}

func attrs(keyValues ...string) map[string]string {
	attributes := make(map[string]string, len(keyValues)/2)
	for i := 0; i+1 < len(keyValues); i += 2 {
		attributes[keyValues[i]] = keyValues[i+1]
	}
	return attributes
}

func renderPlain(elements ...api.Textable) string {
	var b strings.Builder
	for _, element := range elements {
		b.WriteString(element.String())
	}
	return b.String()
}

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
