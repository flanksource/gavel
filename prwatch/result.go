package prwatch

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	clickymarkdown "github.com/flanksource/clicky/markdown"
	"github.com/flanksource/gavel/github"
)

type PRWatchResult struct {
	PR           *github.PRInfo                `json:"pr"`
	Runs         map[int64]*github.WorkflowRun `json:"runs,omitempty"`
	GavelResults []*GavelResultsSummary        `json:"gavelResults,omitempty"`
	Comments     []github.PRComment            `json:"comments,omitempty"`
}

// HasFailedRun reports whether any workflow run has a failed job. A run can
// carry a failed job while its rollup context still reports success, so this is
// an independent failure signal rather than a restatement of the rollup.
func (r PRWatchResult) HasFailedRun() bool {
	for _, run := range r.Runs {
		if run != nil && github.RunHasFailedJob(run) {
			return true
		}
	}
	return false
}

func (r PRWatchResult) Pretty() api.Text {
	text := r.PR.Pretty()
	text = text.NewLine().NewLine().Add(r.prettyWorkflows())
	if gt := r.prettyGavelResults(); gt.String() != "" {
		text = text.NewLine().NewLine().Add(gt)
	}
	if ct := r.prettyComments(); ct.String() != "" {
		text = text.NewLine().NewLine().Add(ct)
	}
	return text
}

func (r PRWatchResult) prettyWorkflows() api.Text {
	if len(r.Runs) == 0 && len(r.PR.StatusCheckRollup) == 0 {
		return clicky.Text("  No checks found", "text-gray-500")
	}

	text := clicky.Text("Workflows:", "font-bold")

	runs := sortedRuns(r.Runs)
	labels := runLabels(runs)
	for _, run := range runs {
		text = text.NewLine().Add(run.PrettyAs(labels[run.DatabaseID]))
	}

	for _, check := range r.PR.StatusCheckRollup {
		runID, err := github.ExtractRunID(check.DetailsURL)
		if err == nil && r.Runs[runID] != nil {
			continue
		}
		text = text.NewLine().Append("  ", "").
			Add(github.StatusIcon(check.Status, check.Conclusion)).
			Append(" "+check.Name, "")
		// A rollup-only check has no jobs to expand, so a failure would
		// otherwise render as a bare name with nothing to act on — and it is
		// exactly what drives a non-zero exit code.
		if github.IsFailureConclusion(check.Conclusion) && check.DetailsURL != "" {
			text = text.NewLine().Append("    "+check.DetailsURL, "text-blue-600")
		}
	}

	return text
}

// sortedRuns gives the workflow list a stable order. Runs are held in a map
// keyed by run ID, so ranging over it directly reorders the whole section
// between invocations.
func sortedRuns(runs map[int64]*github.WorkflowRun) []*github.WorkflowRun {
	sorted := make([]*github.WorkflowRun, 0, len(runs))
	for _, run := range runs {
		if run != nil {
			sorted = append(sorted, run)
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		// A run carries no start time of its own, so order by when its first
		// job started — the same chronology the GitHub UI shows.
		if startA, startB := runStartedAt(a), runStartedAt(b); !startA.Equal(startB) {
			return startA.Before(startB)
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.DatabaseID < b.DatabaseID
	})
	return sorted
}

// runStartedAt is the earliest job start in the run. Skipped jobs never start,
// so their zero time is ignored rather than winning the comparison.
func runStartedAt(run *github.WorkflowRun) time.Time {
	var earliest time.Time
	for i := range run.Jobs {
		started := run.Jobs[i].StartedAt
		if started.IsZero() {
			continue
		}
		if earliest.IsZero() || started.Before(earliest) {
			earliest = started
		}
	}
	return earliest
}

// runLabels disambiguates workflows that ran more than once on the same PR.
// Two runs of the same workflow otherwise render under identical headings and
// read as a duplicated section rather than two separate runs.
func runLabels(runs []*github.WorkflowRun) map[int64]string {
	counts := make(map[string]int, len(runs))
	for _, run := range runs {
		counts[run.Name]++
	}

	labels := make(map[int64]string, len(runs))
	for _, run := range runs {
		labels[run.DatabaseID] = run.Name
		if counts[run.Name] > 1 {
			labels[run.DatabaseID] = fmt.Sprintf("%s (run %d)", run.Name, run.DatabaseID)
		}
	}
	return labels
}

var severityOrder = map[string]int{"critical": 0, "major": 1, "minor": 2, "nitpick": 3, "": 4}

const commentPreviewLineLimit = 5

func (r PRWatchResult) prettyComments() api.Text {
	if len(r.Comments) == 0 {
		return clicky.Text("", "")
	}

	// Group by directory then file
	type fileEntry struct {
		dir      string
		file     string
		comments []github.PRComment
	}
	fileMap := make(map[string]*fileEntry)
	var noPath []github.PRComment

	for _, c := range r.Comments {
		if c.Path == "" {
			noPath = append(noPath, c)
			continue
		}
		if _, ok := fileMap[c.Path]; !ok {
			fileMap[c.Path] = &fileEntry{
				dir:  filepath.Dir(c.Path),
				file: filepath.Base(c.Path),
			}
		}
		fileMap[c.Path].comments = append(fileMap[c.Path].comments, c)
	}

	// Sort files: by directory, then filename
	files := make([]*fileEntry, 0, len(fileMap))
	for _, f := range fileMap {
		sort.Slice(f.comments, func(i, j int) bool {
			si, sj := severityOrder[f.comments[i].Severity], severityOrder[f.comments[j].Severity]
			if si != sj {
				return si < sj
			}
			return f.comments[i].Line < f.comments[j].Line
		})
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].dir != files[j].dir {
			return files[i].dir < files[j].dir
		}
		return files[i].file < files[j].file
	})

	text := clicky.Text(fmt.Sprintf("Comments: (%d)", len(r.Comments)), "font-bold")

	prevDir := ""
	for _, f := range files {
		if f.dir != prevDir && f.dir != "." {
			text = text.NewLine().Append("  "+f.dir+"/", "text-gray-500")
			prevDir = f.dir
		}
		indent := "    "
		pathLabel := f.file
		if f.dir == "." {
			indent = "  "
			pathLabel = f.file
		}
		text = text.NewLine().Append(indent+pathLabel, "text-cyan-600")
		for _, c := range f.comments {
			text = text.NewLine().Add(prettyCommentLine(c, indent+"  "))
		}
	}

	// Pathless comments at the end
	for _, c := range noPath {
		text = text.NewLine().Add(prettyCommentLine(c, "  "))
	}

	return text
}

func prettyCommentLine(c github.PRComment, indent string) api.Text {
	text := clicky.Text(indent, "").Add(github.SeverityIcon(c.Severity))
	if c.Line > 0 {
		text = text.Append(fmt.Sprintf(" :%d", c.Line), "text-gray-500")
	}
	body := renderCommentBody(c.Body)
	lines := body.preview
	if len(lines) == 0 {
		lines = []string{c.Title()}
	}
	style := ""
	if c.IsResolved || c.IsOutdated {
		style = "text-gray-500 line-through"
		tag := "resolved"
		if c.IsOutdated {
			tag = "outdated"
		}
		lines[0] = lines[0] + " (" + tag + ")"
	}
	text = text.Append(" "+lines[0], style)
	for _, line := range lines[1:] {
		text = text.NewLine().Append(indent+"  "+line, style)
	}
	if body.hasMore && body.content != nil {
		text = text.NewLine().Append(indent+"  ", "").Add(api.Collapsed{
			Label:        "show more",
			Content:      body.content,
			CollapseANSI: true,
		})
	}
	return text
}

type renderedCommentBody struct {
	preview []string
	content api.Textable
	hasMore bool
}

var (
	alertMarkerPattern = regexp.MustCompile(`(?m)^\s*>?\s*\[!(?:NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*$\n?`)
	detailsOpenPattern = regexp.MustCompile(`(?i)<details\b[^>]*>`)
	fencePattern       = regexp.MustCompile("^\\s*(```|~~~)")
)

const detailsCloseTag = "</details>"

// stripDetailsBlocks removes every <details>…</details> span, including the
// nested ones CodeRabbit uses. Depth counting rather than a regex because Go's
// RE2 has no negative lookahead, so "innermost block" is not expressible.
// An unbalanced tag leaves the rest of the body untouched rather than eating it.
func stripDetailsBlocks(body string) string {
	var out strings.Builder
	for {
		open := detailsOpenPattern.FindStringIndex(body)
		if open == nil {
			out.WriteString(body)
			return out.String()
		}

		out.WriteString(body[:open[0]])
		depth, cursor := 1, open[1]
		for depth > 0 {
			next := detailsOpenPattern.FindStringIndex(body[cursor:])
			close := strings.Index(strings.ToLower(body[cursor:]), detailsCloseTag)
			if close < 0 {
				// Unterminated: keep what follows verbatim.
				out.WriteString(body[open[0]:])
				return out.String()
			}
			if next != nil && next[0] < close {
				depth++
				cursor += next[1]
				continue
			}
			depth--
			cursor += close + len(detailsCloseTag)
		}
		body = body[cursor:]
	}
}

// sanitizeCommentBody removes the GitHub-flavoured scaffolding that carries no
// information once a comment is rendered as terminal text: collapsed <details>
// sections (whose summary text otherwise survives as a stray "Show more
// details" line) and alert markers such as "[!CAUTION]", which render literally
// because the severity is already shown by the comment's own icon.
//
// Fenced blocks are left alone — a review comment quoting this markup is
// showing it deliberately. HTML comments are not touched here either; the
// markdown parser already drops them, and it knows where the fences are.
func sanitizeCommentBody(body string) string {
	var out, prose []string
	fenced := false

	flushProse := func() {
		if len(prose) == 0 {
			return
		}
		cleaned := alertMarkerPattern.ReplaceAllString(stripDetailsBlocks(strings.Join(prose, "\n")), "")
		out = append(out, cleaned)
		prose = nil
	}

	for _, line := range strings.Split(body, "\n") {
		switch {
		case fencePattern.MatchString(line):
			// An unterminated fence runs to the end of the body, matching how
			// a markdown renderer treats it.
			if !fenced {
				flushProse()
			}
			fenced = !fenced
			out = append(out, line)
		case fenced:
			out = append(out, line)
		default:
			prose = append(prose, line)
		}
	}
	flushProse()

	return strings.Join(out, "\n")
}

func renderCommentBody(body string) renderedCommentBody {
	body = strings.TrimSpace(sanitizeCommentBody(body))
	if body == "" {
		return renderedCommentBody{}
	}

	var content api.Textable
	plain := body
	if doc, err := clicky.ParseMarkdown(body, clickymarkdown.WithPreserveHTML(false)); err == nil && doc != nil {
		content = doc
		plain = doc.String()
	} else {
		content = clicky.Text(body, "whitespace-pre-wrap")
	}

	lines := commentPreviewLines(plain)
	if len(lines) == 0 {
		return renderedCommentBody{content: content}
	}
	preview := lines
	if len(preview) > commentPreviewLineLimit {
		preview = preview[:commentPreviewLineLimit]
	}
	return renderedCommentBody{
		preview: preview,
		content: content,
		hasMore: len(lines) > commentPreviewLineLimit,
	}
}

func commentPreviewLines(body string) []string {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}

	rawLines := strings.Split(body, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
