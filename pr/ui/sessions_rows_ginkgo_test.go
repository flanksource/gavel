package ui

import (
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/github"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("buildSessionRows", func() {
	const (
		projectDir  = "/src/acme"
		linkedPath  = "/src/acme/.shell/worktrees/feature-1"
		repo        = "example/acme"
		featureName = "feature/sessions"
	)
	var (
		now      = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
		projects []sessionProject
	)

	ago := func(d time.Duration) *time.Time { at := now.Add(-d); return &at }
	str := func(s string) *string { return &s }
	overview := func(title string, lastActivity time.Duration, mutate ...func(*captaindb.SessionListSummary)) captaindb.SessionListSummary {
		o := captaindb.SessionListSummary{
			ID: uuid.New(), Source: "claude", Title: str(title), LifecycleStatus: "succeeded",
			ActivityState: "idle", LastActivityAt: ago(lastActivity), ActivityAt: *ago(lastActivity), CWD: str(projectDir),
		}
		for _, fn := range mutate {
			fn(&o)
		}
		return o
	}
	titles := func(rows []sessionRow) []string {
		out := make([]string, len(rows))
		for i, row := range rows {
			out[i] = row.Title
		}
		return out
	}

	BeforeEach(func() {
		projects = []sessionProject{{
			Project: Project{Name: "acme", Dir: projectDir, Repos: []string{repo}},
			Git: &gitstate.State{
				Base: "main", CurrentBranch: "main", BaseCheckedOut: true,
				Worktrees: []gitstate.Worktree{
					{Worktree: gavelgit.Worktree{Path: projectDir, Branch: "main", Primary: true}},
					{
						Worktree: gavelgit.Worktree{Path: linkedPath, Branch: featureName},
						Changes:  gitstate.Changes{Unstaged: 3, Untracked: 1, Adds: 40, Dels: 2},
						Ahead:    2,
					},
				},
				Branches: []gavelgit.BranchInfo{
					{Name: featureName, Ahead: 2, Behind: 5},
					{Name: "landed-run", Ahead: 1, Behind: 0},
				},
			},
		}}
	})

	It("keeps Claude/Codex roots and transcript children, never subagents or admission rows", func() {
		parent := uuid.New()
		rows := buildSessionRows([]captaindb.SessionListSummary{
			overview("interactive root", time.Minute),
			overview("codex root", 2*time.Minute, func(o *captaindb.SessionListSummary) { o.Source = "codex" }),
			overview("gavel run agent", 3*time.Minute, func(o *captaindb.SessionListSummary) {
				o.ParentSessionID, o.ParentRelation = &parent, captaindb.SessionParentRelationTranscript
			}),
			overview("subagent", 4*time.Minute, func(o *captaindb.SessionListSummary) {
				o.ParentSessionID, o.ParentRelation = &parent, captaindb.SessionParentRelationAgent
			}),
			overview("admission row", 5*time.Minute, func(o *captaindb.SessionListSummary) { o.Source = "gavel" }),
		}, projects, nil, nil, "", 0)

		Expect(titles(rows)).To(Equal([]string{"interactive root", "codex root", "gavel run agent"}))
	})

	It("orders sessions waiting on input, then live ones, then by last activity", func() {
		rows := buildSessionRows([]captaindb.SessionListSummary{
			overview("finished recently", time.Minute),
			overview("running", 30*time.Minute, func(o *captaindb.SessionListSummary) { o.ProcessActive = true }),
			overview("finished earlier", time.Hour),
			overview("awaiting approval", 2*time.Hour, func(o *captaindb.SessionListSummary) {
				o.ProcessActive, o.ActivityState = true, string(captaindb.SessionActivityApproval)
			}),
		}, projects, nil, nil, "", 0)

		Expect(titles(rows)).To(Equal([]string{"awaiting approval", "running", "finished recently", "finished earlier"}))
	})

	It("attributes a session to the most specific worktree containing its directory, with that worktree's git state", func() {
		rows := buildSessionRows([]captaindb.SessionListSummary{
			overview("in linked worktree", time.Minute, func(o *captaindb.SessionListSummary) { o.CWD = str(linkedPath + "/pr/ui") }),
		}, projects, nil, nil, "", 0)

		Expect(rows).To(HaveLen(1))
		Expect(rows[0].Project).To(Equal("acme"))
		Expect(rows[0].Worktree).To(Equal(&sessionWorktree{Path: linkedPath, Branch: featureName, Base: "main"}))
		Expect(rows[0].Git).To(Equal(&sessionGit{
			Changes: gitstate.Changes{Unstaged: 3, Untracked: 1, Adds: 40, Dels: 2},
			Ahead:   2, Behind: 5, BaseCheckedOut: true, WorktreeLive: true,
		}))
	})

	It("attributes a session in the project directory to the primary checkout's branch", func() {
		rows := buildSessionRows([]captaindb.SessionListSummary{overview("in main checkout", time.Minute)}, projects, nil, nil, "", 0)

		Expect(rows[0].Worktree).To(Equal(&sessionWorktree{Path: projectDir, Branch: "main", Base: "main", Primary: true}))
		Expect(rows[0].Git).To(Equal(&sessionGit{BaseCheckedOut: true, WorktreeLive: true}))
	})

	It("leaves a session outside every project unattributed", func() {
		rows := buildSessionRows([]captaindb.SessionListSummary{
			overview("elsewhere", time.Minute, func(o *captaindb.SessionListSummary) { o.CWD = str("/src/acme-other") }),
		}, projects, nil, nil, "", 0)

		Expect(rows[0].Project).To(BeEmpty())
		Expect(rows[0].Worktree).To(BeNil())
	})

	It("uses a run's recorded worktree and falls back to its surviving branch once the worktree is removed", func() {
		session := overview("todo run", time.Minute, func(o *captaindb.SessionListSummary) { o.CWD = str("/gone/worktree") })
		todo := &sessionTodo{ID: uuid.NewString(), Title: "Land branches", Status: "in_progress"}
		rows := buildSessionRows([]captaindb.SessionListSummary{session}, projects, map[uuid.UUID]sessionRun{
			session.ID: {Project: "acme", Todo: todo, Worktree: &api.WorktreeState{Path: "/gone/worktree", Branch: "landed-run"}},
		}, nil, "", 0)

		Expect(rows[0].Todo).To(Equal(todo))
		Expect(rows[0].Worktree).To(Equal(&sessionWorktree{Path: "/gone/worktree", Branch: "landed-run", Base: "main"}))
		Expect(rows[0].Git).To(Equal(&sessionGit{Ahead: 1, BaseCheckedOut: true}))
	})

	It("matches the PR whose head is the session's branch in one of the project's repositories", func() {
		checks := &github.CheckSummary{Passed: 3, Failed: 1}
		rows := buildSessionRows([]captaindb.SessionListSummary{
			overview("with PR", time.Minute, func(o *captaindb.SessionListSummary) { o.CWD = str(linkedPath) }),
		}, projects, nil, []github.PRListItem{
			{Number: 7, Repo: "example/other", Source: featureName, State: "OPEN"},
			{Number: 12, Repo: repo, Source: featureName, State: "OPEN", IsDraft: true, URL: "https://example.test/pr/12", CheckStatus: checks, ReviewDecision: "REVIEW_REQUIRED"},
		}, "", 0)

		Expect(rows[0].PR).To(Equal(&sessionPR{
			Number: 12, URL: "https://example.test/pr/12", State: "draft", CheckStatus: "failure", ReviewDecision: "REVIEW_REQUIRED",
		}))
	})

	It("filters to the named project before applying the limit", func() {
		rows := buildSessionRows([]captaindb.SessionListSummary{
			overview("outside 1", time.Minute, func(o *captaindb.SessionListSummary) { o.CWD = str("/tmp/x") }),
			overview("acme newest", 2*time.Minute),
			overview("acme older", 3*time.Minute),
			overview("acme oldest", 4*time.Minute),
		}, projects, nil, nil, "acme", 2)

		Expect(titles(rows)).To(Equal([]string{"acme newest", "acme older"}))
	})

	It("titles an untitled session by the first line of its prompt", func() {
		rows := buildSessionRows([]captaindb.SessionListSummary{
			overview("", time.Minute, func(o *captaindb.SessionListSummary) { o.InitialPrompt = str("  Fix the sniffer\nmore detail") }),
			overview("", 2*time.Minute),
		}, projects, nil, nil, "", 0)

		Expect(titles(rows)).To(Equal([]string{"Fix the sniffer", untitledSession}))
	})
})
