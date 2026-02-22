package prwatch

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/github"
)

type PRWatchResult struct {
	PR       *github.PRInfo                `json:"pr"`
	Runs     map[int64]*github.WorkflowRun `json:"runs,omitempty"`
	Comments []github.PRComment            `json:"comments,omitempty"`
}

func (r PRWatchResult) Pretty() api.Text {
	text := r.PR.Pretty()
	text = text.NewLine().NewLine().Add(r.prettyWorkflows())
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

	rendered := make(map[int64]bool)
	for _, run := range r.Runs {
		if rendered[run.DatabaseID] {
			continue
		}
		rendered[run.DatabaseID] = true
		text = text.NewLine().Add(run.Pretty())
	}

	for _, check := range r.PR.StatusCheckRollup {
		runID, err := github.ExtractRunID(check.DetailsURL)
		if err == nil && rendered[runID] {
			continue
		}
		text = text.NewLine().Append("  ", "").
			Add(github.StatusIcon(check.Status, check.Conclusion)).
			Append(" "+check.Name, "")
	}

	return text
}

var severityOrder = map[string]int{"critical": 0, "major": 1, "minor": 2, "nitpick": 3, "": 4}

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
	title := c.Title()
	if len(title) > 100 {
		title = title[:97] + "..."
	}
	style := ""
	if c.IsResolved || c.IsOutdated {
		style = "text-gray-500 line-through"
		tag := "resolved"
		if c.IsOutdated {
			tag = "outdated"
		}
		title = title + " (" + tag + ")"
	}
	return text.Append(" "+title, style)
}
