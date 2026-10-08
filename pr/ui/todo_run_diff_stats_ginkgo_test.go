package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("runDiffStats", func() {
	var (
		repo    string
		tracked trackedGitServer
	)
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	commit := func(file, content, message string) string {
		Expect(os.WriteFile(filepath.Join(repo, file), []byte(content), 0o644)).To(Succeed())
		git("add", file)
		git("commit", "-q", "-m", message)
		return git("rev-parse", "HEAD")
	}
	runLink := func(issue, run uuid.UUID, at time.Time) native.PromptRunLink {
		return native.PromptRunLink{IssueID: issue, PromptRunID: run, StepKind: native.StepRun, CreatedAt: at}
	}
	overview := func(run uuid.UUID, worktree *api.WorktreeState) captaindb.PromptRunOverview {
		return captaindb.PromptRunOverview{PromptRun: captaindb.PromptRun{ID: run, Workspace: &api.WorkspaceRecord{Worktree: worktree}}}
	}
	BeforeEach(func() {
		tracked = newTrackedGitServer()
		repo = GinkgoT().TempDir()
		// main is the PR base the tracker compares the repository's branches to.
		git("init", "-q", "-b", "main")
		git("config", "user.email", "test@example.com")
		git("config", "user.name", "Test User")
		git("config", "commit.gpgsign", "false")
	})

	It("computes stats from recorded setup..head, not trailers", func() {
		issue := uuid.New()
		commit("base.txt", "base\n", "chore: base")
		setup := commit("wip.txt", "wip\n", "chore(setup): snapshot uncommitted changes\n\nGavel-Issue-Id: "+issue.String())
		commit("a.txt", "a1\na2\n", "feat: a")
		head := commit("a.txt", "a1\n", "fix: trim a")
		// A trailer-tagged commit past head is on no recorded range and must not count.
		commit("stray.txt", "x\ny\nz\n", "feat: stray\n\nGavel-Issue-Id: "+issue.String())
		older, latest, verify := uuid.New(), uuid.New(), uuid.New()
		t0 := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

		stats, err := computeRunDiffStats(tracked.ctx, tracked.tracker,
			map[uuid.UUID][]native.PromptRunLink{issue: {
				runLink(issue, older, t0),
				runLink(issue, latest, t0.Add(time.Hour)),
				{IssueID: issue, PromptRunID: verify, StepKind: native.StepVerify, CreatedAt: t0.Add(2 * time.Hour)},
			}},
			[]captaindb.PromptRunOverview{
				overview(older, &api.WorktreeState{Repo: repo, Setup: setup, Head: setup}),
				overview(latest, &api.WorktreeState{Repo: repo, Setup: setup, Head: head}),
			},
			nil,
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(stats).To(Equal(map[string]gavelgit.DiffStat{issue.String(): {Commits: 2, Files: 1, Adds: 1, Dels: 0}}))
	})

	It("reads a range compared once from git_range_stats, without git", func() {
		issue, run := uuid.New(), uuid.New()
		setup := commit("base.txt", "base\n", "chore: base")
		head := commit("a.txt", "a1\na2\n", "feat: a")
		links := map[uuid.UUID][]native.PromptRunLink{issue: {runLink(issue, run, time.Now())}}
		overviews := []captaindb.PromptRunOverview{overview(run, &api.WorktreeState{Repo: repo, Setup: setup, Head: head})}
		want := map[string]gavelgit.DiffStat{issue.String(): {Commits: 1, Files: 1, Adds: 2}}

		Expect(computeRunDiffStats(tracked.ctx, tracked.tracker, links, overviews, nil)).To(Equal(want))
		var stored int64
		Expect(tracked.db.Raw(`SELECT count(*) FROM git_range_stats WHERE base_sha = ? AND head_sha = ?`, setup, head).Scan(&stored).Error).To(Succeed())
		Expect(stored).To(Equal(int64(1)))

		withoutGit()
		Expect(computeRunDiffStats(tracked.ctx, tracked.tracker, links, overviews, nil)).To(Equal(want))
	})

	It("gives no stats to an issue whose runs recorded no worktree range", func() {
		issue, inCheckout, noWorkspace := uuid.New(), uuid.New(), uuid.New()
		t0 := time.Now()
		stats, err := computeRunDiffStats(tracked.ctx, tracked.tracker,
			map[uuid.UUID][]native.PromptRunLink{issue: {runLink(issue, inCheckout, t0), runLink(issue, noWorkspace, t0.Add(time.Second))}},
			[]captaindb.PromptRunOverview{
				{PromptRun: captaindb.PromptRun{ID: inCheckout, Workspace: &api.WorkspaceRecord{Cwd: repo}}},
				{PromptRun: captaindb.PromptRun{ID: noWorkspace}},
			},
			nil,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(stats).To(BeEmpty())
	})

	It("surfaces a git failure for a recorded range git cannot resolve", func() {
		issue, run := uuid.New(), uuid.New()
		head := commit("a.txt", "a\n", "feat: a")
		_, err := computeRunDiffStats(tracked.ctx, tracked.tracker,
			map[uuid.UUID][]native.PromptRunLink{issue: {runLink(issue, run, time.Now())}},
			[]captaindb.PromptRunOverview{overview(run, &api.WorktreeState{
				Repo: repo, Setup: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", Head: head,
			})},
			nil,
		)
		Expect(err).To(MatchError(ContainSubstring(issue.String())))
	})

	It("fails loudly when a linked run has no overview", func() {
		issue := uuid.New()
		_, err := computeRunDiffStats(tracked.ctx, tracked.tracker, map[uuid.UUID][]native.PromptRunLink{issue: {runLink(issue, uuid.New(), time.Now())}}, nil, nil)
		Expect(err).To(MatchError(captaindb.ErrPromptRunNotFound))
	})

	Context("a landed run", func() {
		// goneSetup/goneHead stand for a run branch that was deleted and gc'd
		// after landing: the recorded range no longer resolves.
		const goneSetup, goneHead = "5e7a0005e7a0005e7a0005e7a0005e7a0005e7a0", "d0b5c96d0b5c96d0b5c96d0b5c96d0b5c96d0b5c"
		// The workspace records a single hook commit whatever the landing carried,
		// so a spec passing only because of it would diff one commit, not two.
		landedOverview := func(run uuid.UUID) captaindb.PromptRunOverview {
			return captaindb.PromptRunOverview{PromptRun: captaindb.PromptRun{ID: run, Workspace: &api.WorkspaceRecord{
				Worktree: &api.WorktreeState{Repo: repo, Branch: "shell/289107d9", Setup: goneSetup, Head: goneHead, BranchDeleted: true},
				Commits:  []api.CommitRecord{{SHA: goneHead}},
			}}}
		}
		landed := func(run uuid.UUID, sha string, commitCount int) map[uuid.UUID]native.RunLanding {
			return map[uuid.UUID]native.RunLanding{run: {
				PromptRunID: run, Via: native.LandingMerge, TargetBranch: "main", LandedSHA: sha, CommitCount: commitCount,
			}}
		}

		It("computes stats from landedSha~N..landedSha, N being the landing's commit count", func() {
			issue, run := uuid.New(), uuid.New()
			commit("base.txt", "base\n", "chore: base")
			commit("a.txt", "a1\na2\n", "feat: a")
			landedSHA := commit("b.txt", "b1\n", "feat: b")

			stats, err := computeRunDiffStats(tracked.ctx, tracked.tracker,
				map[uuid.UUID][]native.PromptRunLink{issue: {runLink(issue, run, time.Now())}},
				[]captaindb.PromptRunOverview{landedOverview(run)},
				landed(run, landedSHA, 2),
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(stats).To(Equal(map[string]gavelgit.DiffStat{issue.String(): {Commits: 2, Files: 2, Adds: 3, Dels: 0}}))
		})

		It("gives no stats when the landed commit is not in the local repository", func() {
			issue, run := uuid.New(), uuid.New()
			commit("base.txt", "base\n", "chore: base")
			stats, err := computeRunDiffStats(tracked.ctx, tracked.tracker,
				map[uuid.UUID][]native.PromptRunLink{issue: {runLink(issue, run, time.Now())}},
				[]captaindb.PromptRunOverview{landedOverview(run)},
				landed(run, "feedfacefeedfacefeedfacefeedfacefeedface", 1),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(stats).To(BeEmpty())
		})

		It("surfaces a git failure when the landed commit has fewer ancestors than it landed", func() {
			issue, run := uuid.New(), uuid.New()
			landedSHA := commit("a.txt", "a\n", "feat: a")
			_, err := computeRunDiffStats(tracked.ctx, tracked.tracker,
				map[uuid.UUID][]native.PromptRunLink{issue: {runLink(issue, run, time.Now())}},
				[]captaindb.PromptRunOverview{landedOverview(run)},
				landed(run, landedSHA, 3),
			)
			Expect(err).To(MatchError(ContainSubstring(issue.String())))
		})
	})
})
