package ui

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/github"
	"github.com/google/uuid"
)

// sessionRow is one agent session on the Sessions tab: Captain's session
// state joined with where its work sits — the worktree it ran in, that
// worktree's uncommitted and unmerged changes, and the branch's PR.
type sessionRow struct {
	ID                string           `json:"id"`
	ProviderSessionID string           `json:"providerSessionId,omitempty"`
	Title             string           `json:"title"`
	Project           string           `json:"project,omitempty"`
	Source            string           `json:"source"`
	Provider          string           `json:"provider,omitempty"`
	Model             string           `json:"model,omitempty"`
	Effort            string           `json:"effort,omitempty"`
	LifecycleStatus   string           `json:"lifecycleStatus"`
	ActivityState     string           `json:"activityState"`
	StateReason       string           `json:"stateReason,omitempty"`
	ProcessActive     bool             `json:"processActive"`
	PID               int64            `json:"pid,omitempty"`
	StartedAt         *time.Time       `json:"startedAt,omitempty"`
	LastActivityAt    time.Time        `json:"lastActivityAt"`
	DurationMs        int64            `json:"durationMs,omitempty"`
	CWD               string           `json:"cwd,omitempty"`
	Context           *sessionContext  `json:"context,omitempty"`
	Tokens            sessionTokens    `json:"tokens"`
	CostUSD           float64          `json:"costUsd"`
	Todo              *sessionTodo     `json:"todo,omitempty"`
	Worktree          *sessionWorktree `json:"worktree,omitempty"`
	Git               *sessionGit      `json:"git,omitempty"`
	PR                *sessionPR       `json:"pr,omitempty"`
}

type sessionContext struct {
	UsedTokens   int64 `json:"usedTokens"`
	WindowTokens int64 `json:"windowTokens"`
	FreePercent  int   `json:"freePercent"`
}

type sessionTokens struct {
	Input     int64 `json:"input"`
	Output    int64 `json:"output"`
	CacheRead int64 `json:"cacheRead"`
	Total     int64 `json:"total"`
}

type sessionTodo struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// sessionWorktree is the checkout a session worked in. Base is the branch its
// work merges into (the project's base), not the sha the worktree forked at.
type sessionWorktree struct {
	Path    string `json:"path"`
	Branch  string `json:"branch"`
	Base    string `json:"base"`
	Primary bool   `json:"primary"`
}

// sessionGit is the live git state of the session's branch. Changes are the
// worktree's uncommitted files; they are zero once the worktree is gone.
type sessionGit struct {
	Changes        projectGitChanges `json:"changes"`
	Ahead          int               `json:"ahead"`
	Behind         int               `json:"behind"`
	BaseCheckedOut bool              `json:"baseCheckedOut"`
	// WorktreeLive is false when the recorded worktree has been removed.
	WorktreeLive bool `json:"worktreeLive"`
}

type sessionPR struct {
	Number         int    `json:"number"`
	URL            string `json:"url"`
	State          string `json:"state"`
	CheckStatus    string `json:"checkStatus"`
	ReviewDecision string `json:"reviewDecision,omitempty"`
	Mergeable      string `json:"mergeable,omitempty"`
}

// sessionProject is a configured project with its git state; Git is nil when
// the repository could not be inspected (the error is reported separately).
type sessionProject struct {
	Project Project
	Git     *projectGitResponse
}

// sessionRun is what a gavel run recorded about the session that executed it.
type sessionRun struct {
	Project  string
	Todo     *sessionTodo
	Worktree *api.WorktreeState
}

const untitledSession = "Untitled session"

// isTopLevelAgentSession keeps the sessions a person drives or a gavel run
// executes: Claude/Codex roots, and transcript children (the agent session a
// gavel or chat run binds under its own admission row). Subagents ("agent"
// children) belong to their parent's transcript, and Captain's own admission
// rows carry no usage of their own.
func isTopLevelAgentSession(overview captaindb.SessionListSummary) bool {
	if overview.Source != "claude" && overview.Source != "codex" {
		return false
	}
	return overview.ParentSessionID == nil || overview.ParentRelation == captaindb.SessionParentRelationTranscript
}

// buildSessionRows joins the session overviews with project git state, run
// links and PRs, keeps the top-level agent sessions of the named project (all
// when empty), and orders them sessions-that-need-you, running, then most
// recent activity first.
func buildSessionRows(
	overviews []captaindb.SessionListSummary,
	projects []sessionProject,
	runs map[uuid.UUID]sessionRun,
	prs []github.PRListItem,
	projectFilter string,
	limit int,
) []sessionRow {
	rows := make([]sessionRow, 0, len(overviews))
	for _, overview := range overviews {
		if !isTopLevelAgentSession(overview) {
			continue
		}
		row := sessionRowFromOverview(overview)
		attachSessionWork(&row, overview, projects, runs)
		if projectFilter != "" && row.Project != projectFilter {
			continue
		}
		row.PR = sessionPRFor(row, projects, prs)
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool { return sessionRowBefore(rows[i], rows[j]) })
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

func sessionNeedsInput(row sessionRow) bool {
	return row.ActivityState == string(captaindb.SessionActivityAsk) || row.ActivityState == string(captaindb.SessionActivityApproval)
}

func sessionRowBefore(a, b sessionRow) bool {
	if na, nb := sessionNeedsInput(a), sessionNeedsInput(b); na != nb {
		return na
	}
	if a.ProcessActive != b.ProcessActive {
		return a.ProcessActive
	}
	return a.LastActivityAt.After(b.LastActivityAt)
}

func sessionRowFromOverview(o captaindb.SessionListSummary) sessionRow {
	row := sessionRow{
		ID:              o.ID.String(),
		Title:           sessionTitle(o),
		Source:          o.Source,
		Provider:        o.Provider,
		Model:           deref(o.Model),
		Effort:          deref(o.Effort),
		LifecycleStatus: o.LifecycleStatus,
		ActivityState:   o.ActivityState,
		StateReason:     deref(o.StateReason),
		ProcessActive:   o.ProcessActive,
		StartedAt:       o.StartedAt,
		LastActivityAt:  o.ActivityAt,
		CWD:             sessionCWD(o),
		Tokens: sessionTokens{
			Input: o.InputTokens, Output: o.OutputTokens, CacheRead: o.CacheReadTokens, Total: o.TotalTokens,
		},
		CostUSD: o.CostUSD,
	}
	if o.ProviderSessionID != nil {
		row.ProviderSessionID = *o.ProviderSessionID
	}
	if o.PID != nil && o.ProcessActive {
		row.PID = *o.PID
	}
	if o.StartedAt != nil {
		end := row.LastActivityAt
		if o.EndedAt != nil {
			end = *o.EndedAt
		}
		row.DurationMs = max(end.Sub(*o.StartedAt).Milliseconds(), 0)
	}
	if o.ContextTokens != nil && o.ContextWindowTokens != nil && *o.ContextWindowTokens > 0 {
		row.Context = &sessionContext{UsedTokens: *o.ContextTokens, WindowTokens: *o.ContextWindowTokens}
		if o.ContextFreePercent != nil {
			row.Context.FreePercent = *o.ContextFreePercent
		} else {
			row.Context.FreePercent = int(100 - (*o.ContextTokens*100) / *o.ContextWindowTokens)
		}
	}
	return row
}

func sessionTitle(o captaindb.SessionListSummary) string {
	for _, candidate := range []*string{o.Title, o.Slug, o.InitialPrompt} {
		if text := firstLine(deref(candidate)); text != "" {
			return text
		}
	}
	return untitledSession
}

func firstLine(text string) string {
	text = strings.TrimSpace(text)
	if index := strings.IndexByte(text, '\n'); index >= 0 {
		text = strings.TrimSpace(text[:index])
	}
	const maxTitle = 160
	if len(text) > maxTitle {
		text = text[:maxTitle] + "…"
	}
	return text
}

// sessionCWD prefers the directory the session recorded; a live process's
// directory is the fallback for sessions that never reported one.
func sessionCWD(o captaindb.SessionListSummary) string {
	if cwd := deref(o.CWD); cwd != "" {
		return cwd
	}
	return deref(o.ProcessCWD)
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// attachSessionWork resolves the session's project, worktree and git state:
// from the gavel run it executed when there is one, else from the project
// worktree its directory lies in.
func attachSessionWork(row *sessionRow, overview captaindb.SessionListSummary, projects []sessionProject, runs map[uuid.UUID]sessionRun) {
	if run, ok := runs[overview.ID]; ok {
		row.Project = run.Project
		row.Todo = run.Todo
		if run.Worktree != nil && run.Worktree.Path != "" {
			project := findSessionProject(projects, run.Project)
			row.Worktree, row.Git = recordedWorktree(project, run.Worktree)
			return
		}
	}
	project, worktree := projectWorktreeFor(projects, row.CWD)
	if project == nil {
		return
	}
	if row.Project == "" {
		row.Project = project.Project.Name
	}
	if worktree == nil {
		return
	}
	row.Worktree = &sessionWorktree{Path: worktree.Path, Branch: worktree.Branch, Base: project.Git.Base, Primary: worktree.Primary}
	row.Git = liveGit(project.Git, *worktree)
}

func findSessionProject(projects []sessionProject, name string) *sessionProject {
	for i := range projects {
		if projects[i].Project.Name == name {
			return &projects[i]
		}
	}
	return nil
}

// recordedWorktree pairs a run's recorded worktree with the project's live git
// state: the live worktree at that path when it still exists, else the branch
// alone when only it survives (its worktree removed at teardown).
func recordedWorktree(project *sessionProject, recorded *api.WorktreeState) (*sessionWorktree, *sessionGit) {
	view := &sessionWorktree{Path: recorded.Path, Branch: recorded.Branch}
	if project == nil || project.Git == nil {
		return view, nil
	}
	view.Base = project.Git.Base
	for _, wt := range project.Git.Worktrees {
		if samePath(wt.Path, recorded.Path) && !wt.Prunable {
			view.Branch, view.Primary = wt.Branch, wt.Primary
			return view, liveGit(project.Git, wt)
		}
	}
	for _, branch := range project.Git.Branches {
		if branch.Name == recorded.Branch {
			return view, &sessionGit{Ahead: branch.Ahead, Behind: branch.Behind, BaseCheckedOut: project.Git.BaseCheckedOut}
		}
	}
	return view, nil
}

func liveGit(state *projectGitResponse, wt projectGitWorktree) *sessionGit {
	git := &sessionGit{Changes: wt.Changes, Ahead: wt.Ahead, BaseCheckedOut: state.BaseCheckedOut, WorktreeLive: true}
	for _, branch := range state.Branches {
		if branch.Name == wt.Branch {
			git.Behind = branch.Behind
		}
	}
	return git
}

// projectWorktreeFor finds the project whose worktree (or, failing that,
// whose directory) most specifically contains dir: the longest matching path
// wins, so a linked worktree nested under the primary checkout beats it.
func projectWorktreeFor(projects []sessionProject, dir string) (*sessionProject, *projectGitWorktree) {
	if dir == "" {
		return nil, nil
	}
	var (
		bestProject  *sessionProject
		bestWorktree *projectGitWorktree
		bestLen      = -1
	)
	consider := func(project *sessionProject, worktree *projectGitWorktree, root string) {
		if root == "" || !pathWithin(dir, root) || len(root) <= bestLen {
			return
		}
		bestProject, bestWorktree, bestLen = project, worktree, len(root)
	}
	// Worktrees are considered before the project directory: the primary
	// checkout's path equals it, and on a tie the worktree must win.
	for i := range projects {
		project := &projects[i]
		if project.Git != nil {
			for j := range project.Git.Worktrees {
				if wt := &project.Git.Worktrees[j]; !wt.Prunable {
					consider(project, wt, wt.Path)
				}
			}
		}
		consider(project, nil, project.Project.ResolvedDir())
	}
	return bestProject, bestWorktree
}

func samePath(a, b string) bool { return filepath.Clean(a) == filepath.Clean(b) }

func pathWithin(dir, root string) bool {
	dir, root = filepath.Clean(dir), filepath.Clean(root)
	return dir == root || strings.HasPrefix(dir, root+string(filepath.Separator))
}

// sessionPRFor is the PR whose head is the session's branch, in one of the
// session project's repositories (any repository when the project lists none).
func sessionPRFor(row sessionRow, projects []sessionProject, prs []github.PRListItem) *sessionPR {
	if row.Worktree == nil || row.Worktree.Branch == "" {
		return nil
	}
	var repos []string
	if project := findSessionProject(projects, row.Project); project != nil {
		repos = project.Project.Repos
	}
	for _, pr := range prs {
		if pr.Source != row.Worktree.Branch || (len(repos) > 0 && !containsFold(repos, pr.Repo)) {
			continue
		}
		return &sessionPR{
			Number: pr.Number, URL: pr.URL, State: sessionPRState(pr), CheckStatus: sessionCheckStatus(pr.CheckStatus),
			ReviewDecision: pr.ReviewDecision, Mergeable: pr.Mergeable,
		}
	}
	return nil
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func sessionPRState(pr github.PRListItem) string {
	state := strings.ToLower(pr.State)
	if state == "open" && pr.IsDraft {
		return "draft"
	}
	return state
}

func sessionCheckStatus(checks *github.CheckSummary) string {
	switch {
	case checks == nil:
		return "none"
	case checks.Failed > 0:
		return "failure"
	case checks.Running > 0 || checks.Pending > 0:
		return "pending"
	case checks.Passed > 0:
		return "success"
	default:
		return "none"
	}
}
