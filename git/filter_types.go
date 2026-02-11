package git

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/gavel/repomap"

	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/commons/collections"
	"github.com/samber/lo"
)

type HistoryOptions struct {
	Path      string    `flag:"path" help:"Path to git repository (defaults to current directory)" default:"."`
	Args      []string  `flag:"arg" help:"Commit SHAs, commit ranges (main..branch), or file paths to filter by" args:"true" stdin:"true"`
	Since     time.Time `flag:"since" help:"Start date for commit filtering (e.g., '2024-01-01', '2 weeks ago')"`
	Until     time.Time `flag:"until" help:"End date for commit filtering (e.g., '2024-12-31', 'yesterday')"`
	Author    []string  `flag:"author" help:"Filter commits by author name or email (partial match, supports multiple values)"`
	Message   string    `flag:"message" help:"Filter commits by message content"`
	ShowPatch bool      `flag:"show-patch" help:"Include patch/diff in the commit data" default:"false"`

	// Internal fields populated by ParseArgs
	CommitShas   []string `flag:"-" json:"-"`
	CommitRanges []string `flag:"-" json:"-"`
	FilePaths    []string `flag:"-" json:"-"`
}

func (opt HistoryOptions) Pretty() api.Text {
	t := clicky.Text("").Append(icons.Folder).Space().Append(opt.Path, "font-mono italic")

	if !opt.Since.IsZero() {
		t = t.Append(" since=", "text-muted").Append(opt.Since)
	}

	if !opt.Until.IsZero() {
		t = t.Append(" until=", "text-muted").Append(opt.Until)
	}

	for _, author := range opt.Author {
		t = t.Append(" author=", "text-muted").Append(fmt.Sprintf("%v", author), "font-mono")
	}
	if opt.Message != "" {
		t = t.Append(" message=", "text-muted").Append(opt.Message, "font-mono")
	}
	return t
}

func (opt AnalyzeOptions) Pretty() api.Text {
	t := opt.HistoryOptions.Pretty()
	if opt.Model != "" {
		t = t.Append(" model=", "text-muted").Append(opt.Model, "font-mono")
	}
	if opt.MaxConcurrent > 0 {
		t = t.Append(" max-concurrent=", "text-muted").Append(fmt.Sprintf("%d", opt.MaxConcurrent), "font-mono")
	}
	if len(opt.ScopeTypes) > 0 {
		t = t.Append(" scopes=", "text-muted").Append(fmt.Sprintf("%v", opt.ScopeTypes), "font-mono")
	}
	if len(opt.CommitTypes) > 0 {
		t = t.Append(" commit-types=", "text-muted").Append(fmt.Sprintf("%v", opt.CommitTypes), "font-mono")
	}
	if len(opt.Technologies) > 0 {
		t = t.Append(" tech=", "text-muted").Append(fmt.Sprintf("%v", opt.Technologies), "font-mono")
	}
	return t

}

type AnalyzeOptions struct {
	HistoryOptions `json:",inline"`
	Input          []string         `json:"input" flag:"input" help:"Input JSON files from previous analysis runs (supports multiple files)"`
	Model          string           `json:"model" flag:"model" help:"AI model to use for analysis"`
	MaxConcurrent  int              `json:"max_concurrent" flag:"max-concurrent" help:"Maximum concurrent analysis tasks" default:"4"`
	ScopeTypes     []string         `json:"scope_types" flag:"scope" help:"Limit analysis to these scope types"`
	CommitTypes    []string         `json:"commit_types" flag:"commit-types" help:"Limit analysis to these commit types"`
	Technologies   []string         `json:"technologies" flag:"tech" help:"Limit analysis to these technologies"`
	Debug          bool             `json:"debug" flag:"debug" help:"Enable debug logging"`
	MinScore       int              `json:"min_score" flag:"ai-min-score" help:"Minimum score for commit to be excluded from ai analysis" default:"15"`
	AI             bool             `json:"ai" flag:"ai" help:"Enable AI-powered analysis"`
	AITimeout      time.Duration    `json:"ai_timeout" flag:"ai-timeout" help:"Timeout for AI analysis per commit" default:"60s"`
	Summary        bool             `json:"summary" flag:"summary" help:"Generate a tree based summary of the analysis results"`
	SummaryWindow  GroupByWindow    `json:"summary_window,omitempty" flag:"summary-window" help:"Time window for summary grouping (day, week, month, year), dynamically groups based on total time range if not set"`
	Short          bool             `json:"show_files" flag:"short"  help:"Show short summary with files changed instead of full analysis"`
	agent          ai.Agent         `json:"-"`
	arch           repomap.ArchConf `json:"-"`
}

func match(item any, list []string) bool {
	if len(list) == 0 {
		return true
	}
	if len(lo.Filter(list, func(i string, _ int) bool { return i != "" })) == 0 {
		return true
	}

	s := ""
	if item == nil {
		s = ""
	} else if _s, ok := item.(string); ok {
		s = _s
	} else {
		s = fmt.Sprintf("%v", item)
		return false
	}

	if s == "" {
		_, negated := collections.MatchItem("", list...)
		return !negated
	}

	match, _ := collections.MatchItem(s, list...)
	return match
}

func (a HistoryOptions) Matches(commit Commit) bool {

	if len(a.Author) > 0 {
		matched := false
		for _, author := range a.Author {
			if commit.Author.Matches(author) || commit.Committer.Matches(author) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if !match(commit.Subject, []string{a.Message}) {
		return false
	}

	if !a.Since.IsZero() && commit.Author.Date.Before(a.Since) {
		return false
	}
	if !a.Until.IsZero() && commit.Author.Date.After(a.Until) {
		return false
	}

	return true

}

func (a AnalyzeOptions) Matches(commit Commit, change CommitChange) bool {
	f, _ := a.arch.GetFileMap(change.File, commit.Hash)

	if f != nil {
		if len(a.ScopeTypes) > 0 {
			matched, negated := collections.MatchAny(f.Scopes.ToString(), a.ScopeTypes...)
			if negated || !matched {
				return false
			}
		}
		if len(a.Technologies) > 0 {
			matched, negated := collections.MatchAny(f.Tech.ToString(), a.Technologies...)
			if negated || !matched {
				return false
			}
		}

	}

	if len(change.Scope) != 0 && !match(change.Scope, a.ScopeTypes) {
		return false
	}
	if len(change.Tech) > 0 && !match(change.Tech, a.Technologies) {
		return false
	}
	if !match(commit.CommitType, a.CommitTypes) {
		return false
	}

	return true

}

// ParseArgs separates Args into CommitShas, CommitRanges, and FilePaths
func (o *HistoryOptions) ParseArgs() error {
	for _, arg := range o.Args {
		switch detectArgType(arg) {
		case "commit":
			o.CommitShas = append(o.CommitShas, arg)
		case "range":
			o.CommitRanges = append(o.CommitRanges, arg)
		case "path":
			o.FilePaths = append(o.FilePaths, arg)
		}
	}
	return nil
}

// detectArgType returns "commit", "range", or "path" based on argument pattern
func detectArgType(arg string) string {
	// Range: contains .. or ...
	if strings.Contains(arg, "..") {
		return "range"
	}

	// Path: contains wildcards or path separators
	if strings.ContainsAny(arg, "*?/") {
		return "path"
	}

	// Commit SHA: 7-40 character hexadecimal string
	matched, _ := regexp.MatchString("^[0-9a-f]{7,40}$", arg)
	if matched {
		return "commit"
	}

	// Default to commit for branch names like "main", "feature-branch"
	// Will be validated later in GetCommitHistory
	return "commit"
}
