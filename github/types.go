package github

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

type PRInfo struct {
	Number            int          `json:"number"`
	Title             string       `json:"title"`
	Author            PRAuthor     `json:"author"`
	HeadRefName       string       `json:"headRefName"`
	BaseRefName       string       `json:"baseRefName"`
	State             string       `json:"state"`
	IsDraft           bool         `json:"isDraft"`
	ReviewDecision    string       `json:"reviewDecision"`
	Mergeable         string       `json:"mergeable"`
	URL               string       `json:"url"`
	StatusCheckRollup StatusChecks `json:"statusCheckRollup"`
}

type PRAuthor struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

type StatusChecks []StatusCheck

type StatusCheck struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	WorkflowName string `json:"workflowName"`
	DetailsURL   string `json:"detailsUrl"`
}

type WorkflowRun struct {
	DatabaseID int64  `json:"databaseId"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	URL        string `json:"url"`
	Jobs       []Job  `json:"jobs"`
}

type Job struct {
	DatabaseID  int64     `json:"databaseId"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
	URL         string    `json:"url"`
	Steps       []Step    `json:"steps"`
}

type Step struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Number     int    `json:"number"`
	Logs       string `json:"logs,omitempty"`
}

func (sc StatusChecks) AllComplete() bool {
	if len(sc) == 0 {
		return false
	}
	for _, c := range sc {
		if c.Status != "COMPLETED" {
			return false
		}
	}
	return true
}

func (sc StatusChecks) HasFailure() bool {
	for _, c := range sc {
		if c.Conclusion == "FAILURE" || c.Conclusion == "TIMED_OUT" || c.Conclusion == "STARTUP_FAILURE" {
			return true
		}
	}
	return false
}

func StatusIcon(status, conclusion string) api.Text {
	switch {
	case status == "COMPLETED" && (conclusion == "SUCCESS" || conclusion == "NEUTRAL"):
		return clicky.Text("").Add(icons.Check.WithStyle("text-green-600"))
	case status == "COMPLETED" && (conclusion == "FAILURE" || conclusion == "TIMED_OUT" || conclusion == "STARTUP_FAILURE"):
		return clicky.Text("").Add(icons.Cross.WithStyle("text-red-600"))
	case status == "COMPLETED" && conclusion == "CANCELLED":
		return clicky.Text("").Add(icons.Stop.WithStyle("text-gray-500"))
	case status == "COMPLETED" && conclusion == "SKIPPED":
		return clicky.Text("").Add(icons.Skip.WithStyle("text-gray-500"))
	case status == "IN_PROGRESS":
		return clicky.Text("●", "text-yellow-600")
	case status == "QUEUED" || status == "PENDING" || status == "WAITING":
		return clicky.Text("○", "text-gray-500")
	default:
		return clicky.Text("?", "text-gray-500")
	}
}

func (s Step) Pretty() api.Text {
	text := clicky.Text("      ", "").
		Add(StatusIcon(strings.ToUpper(s.Status), strings.ToUpper(s.Conclusion))).
		Append(" "+s.Name, "text-gray-500")
	if s.Logs != "" {
		for _, line := range strings.Split(strings.TrimSpace(s.Logs), "\n") {
			text = text.NewLine().Append("        "+line, "text-gray-500")
		}
	}
	return text
}

func (j Job) Pretty() api.Text {
	text := clicky.Text("    ", "").
		Add(StatusIcon(strings.ToUpper(j.Status), strings.ToUpper(j.Conclusion))).
		Append(" "+j.Name, "").
		Append(" "+FormatDuration(j), "text-gray-500")

	if !strings.EqualFold(j.Conclusion, "failure") {
		return text
	}
	for _, step := range j.Steps {
		text = text.NewLine().Add(step.Pretty())
	}
	return text
}

func (r WorkflowRun) Pretty() api.Text {
	text := clicky.Text("  ", "").
		Add(StatusIcon(strings.ToUpper(r.Status), strings.ToUpper(r.Conclusion))).
		Append(" "+r.Name, "font-bold")
	for _, job := range r.Jobs {
		text = text.NewLine().Add(job.Pretty())
	}
	return text
}

func (pr PRInfo) Pretty() api.Text {
	title := clicky.Text(fmt.Sprintf("PR #%d: ", pr.Number), "font-bold").
		Append(pr.Title, "font-bold")

	meta := clicky.Text("  ", "").
		Append(pr.BaseRefName, "text-cyan-600").
		Append(" ← ", "text-gray-500").
		Append(pr.HeadRefName, "text-cyan-600").
		Append(" | ", "text-gray-500").
		Append(pr.Author.Login, "text-blue-600")

	if pr.State != "" {
		meta = meta.Append(" | ", "text-gray-500").Append(pr.State, StateStyle(pr.State))
	}
	if pr.IsDraft {
		meta = meta.Append(" | ", "text-gray-500").Append("DRAFT", "text-gray-500")
	}
	if pr.ReviewDecision != "" {
		meta = meta.Append(" | Review: ", "text-gray-500").
			Append(pr.ReviewDecision, ReviewStyle(pr.ReviewDecision))
	}
	if pr.Mergeable != "" {
		meta = meta.Append(" | ", "text-gray-500").
			Append(pr.Mergeable, MergeableStyle(pr.Mergeable))
	}

	return title.NewLine().Add(meta)
}

func FormatDuration(job Job) string {
	if job.StartedAt.IsZero() {
		if strings.EqualFold(job.Status, "in_progress") {
			return "(running...)"
		}
		return ""
	}
	end := job.CompletedAt
	if end.IsZero() {
		end = time.Now()
		return fmt.Sprintf("(running %s...)", end.Sub(job.StartedAt).Truncate(time.Second))
	}
	d := end.Sub(job.StartedAt).Truncate(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("(%s)", d)
	}
	return fmt.Sprintf("(%dm %ds)", int(d.Minutes()), int(d.Seconds())%60)
}

func StateStyle(state string) string {
	switch state {
	case "OPEN":
		return "text-green-600"
	case "MERGED":
		return "text-purple-600"
	case "CLOSED":
		return "text-red-600"
	default:
		return ""
	}
}

func ReviewStyle(decision string) string {
	switch decision {
	case "APPROVED":
		return "text-green-600"
	case "CHANGES_REQUESTED":
		return "text-red-600"
	default:
		return "text-yellow-600"
	}
}

func MergeableStyle(mergeable string) string {
	switch mergeable {
	case "MERGEABLE":
		return "text-green-600"
	case "CONFLICTING":
		return "text-red-600"
	default:
		return "text-yellow-600"
	}
}
