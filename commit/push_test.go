package commit

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/github"
)

func runGit(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(),
		"GIT_AUTHOR_NAME=tester", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=tester", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	out, err := cmd.CombinedOutput()
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), "git %v: %s", args, string(out))
}

var _ = Describe("findAncestorPRs", func() {
	var workDir string

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "main")
		runGit(workDir, "commit", "--allow-empty", "-m", "root")

		// Create branches locally then mirror them as origin/<name> refs so
		// `git merge-base --is-ancestor origin/<branch> HEAD` resolves without
		// a real remote.
		runGit(workDir, "branch", "ancestor-a")
		runGit(workDir, "checkout", "-b", "work")
		runGit(workDir, "commit", "--allow-empty", "-m", "work-1")
		runGit(workDir, "branch", "ancestor-b")
		runGit(workDir, "commit", "--allow-empty", "-m", "work-2")

		// Divergent branch: shares root but adds its own commit not in HEAD.
		runGit(workDir, "branch", "divergent", "main")
		runGit(workDir, "checkout", "divergent")
		runGit(workDir, "commit", "--allow-empty", "-m", "divergent-only")
		runGit(workDir, "checkout", "work")

		for _, b := range []string{"ancestor-a", "ancestor-b", "divergent", "work"} {
			runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", b), b)
		}
	})

	It("keeps only PRs whose head branch is an ancestor of HEAD", func() {
		prs := github.PRSearchResults{
			{Number: 1, Source: "ancestor-a"},
			{Number: 2, Source: "ancestor-b"},
			{Number: 3, Source: "divergent"},
			{Number: 4, Source: "work"}, // same as HEAD → trivially ancestor
			{Number: 5, Source: ""},     // skipped: no source branch
			{Number: 6, Source: "does-not-exist"},
		}

		out := findAncestorPRs(workDir, prs, gitIsAncestor)

		var nums []int
		for _, pr := range out {
			nums = append(nums, pr.Number)
		}
		Expect(nums).To(ConsistOf(1, 2, 4))
	})
})

func TestChoosePRReturnsSelectedPR(t *testing.T) {
	previous := promptSelectFunc
	promptSelectFunc = func(options []promptSelectOption, title string) (promptSelectOption, bool) {
		if title != "Pick a PR" {
			t.Fatalf("unexpected title: %q", title)
		}
		if len(options) != 2 {
			t.Fatalf("unexpected option count: %d", len(options))
		}
		return options[1], true
	}
	defer func() {
		promptSelectFunc = previous
	}()

	prs := []github.PRListItem{
		{Number: 17, Title: "First", Source: "feat/a", Target: "main"},
		{Number: 23, Title: "Second", Source: "feat/b", Target: "main"},
	}

	selected, err := choosePR("Pick a PR", prs)
	if err != nil {
		t.Fatalf("choosePR returned error: %v", err)
	}
	if selected == nil || selected.Number != 23 {
		t.Fatalf("expected PR #23, got %#v", selected)
	}
}

func TestChoosePRReturnsNilWhenCancelled(t *testing.T) {
	previous := promptSelectFunc
	promptSelectFunc = func(options []promptSelectOption, title string) (promptSelectOption, bool) {
		return promptSelectOption{}, false
	}
	defer func() {
		promptSelectFunc = previous
	}()

	selected, err := choosePR("Pick a PR", []github.PRListItem{
		{Number: 17, Title: "First", Source: "feat/a", Target: "main"},
	})
	if err != nil {
		t.Fatalf("choosePR returned error: %v", err)
	}
	if selected != nil {
		t.Fatalf("expected nil selection, got %#v", selected)
	}
}

var _ = Describe("decidePushTarget", func() {
	branchMatchPR := github.PRListItem{Number: 42, Source: "feat/foo", Target: "main"}
	otherOpenPR := github.PRListItem{Number: 17, Source: "feat/bar", Target: "main"}

	newDeps := func(prs []github.PRListItem, ancestors map[string]bool) pushDeps {
		d := defaultPushDeps()
		d.searchPRs = func(_ github.Options, _ github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
			return prs, nil, nil
		}
		d.isAncestor = func(_, ref, _ string) bool { return ancestors[ref] }
		return d
	}

	It("picks branch-match when an open PR's head equals current branch", func() {
		deps := newDeps([]github.PRListItem{otherOpenPR, branchMatchPR}, nil)
		tgt, cands, err := decidePushTarget(context.Background(), github.Options{}, "feat/foo", deps)
		Expect(err).ToNot(HaveOccurred())
		Expect(tgt.kind).To(Equal(pushTargetBranchMatch))
		Expect(tgt.pr.Number).To(Equal(42))
		Expect(cands).To(BeNil())
	})

	It("returns newPR when no PR matches and none are ancestors", func() {
		deps := newDeps([]github.PRListItem{otherOpenPR}, map[string]bool{})
		tgt, cands, err := decidePushTarget(context.Background(), github.Options{}, "feat/new", deps)
		Expect(err).ToNot(HaveOccurred())
		Expect(tgt.kind).To(Equal(pushTargetNewPR))
		Expect(cands).To(BeEmpty())
	})

	It("returns ancestor-match with a single candidate when exactly one PR head is ancestor", func() {
		deps := newDeps([]github.PRListItem{otherOpenPR}, map[string]bool{"origin/feat/bar": true})
		tgt, cands, err := decidePushTarget(context.Background(), github.Options{}, "feat/new", deps)
		Expect(err).ToNot(HaveOccurred())
		Expect(tgt.kind).To(Equal(pushTargetAncestorPR))
		Expect(tgt.pr.Number).To(Equal(17))
		Expect(cands).To(HaveLen(1))
	})

	It("returns ancestor-match with multiple candidates (no pr set) for caller to prompt", func() {
		second := github.PRListItem{Number: 23, Source: "feat/baz", Target: "main"}
		deps := newDeps([]github.PRListItem{otherOpenPR, second}, map[string]bool{
			"origin/feat/bar": true,
			"origin/feat/baz": true,
		})
		tgt, cands, err := decidePushTarget(context.Background(), github.Options{}, "feat/new", deps)
		Expect(err).ToNot(HaveOccurred())
		Expect(tgt.kind).To(Equal(pushTargetAncestorPR))
		Expect(tgt.pr).To(BeNil(), "caller must prompt when >1 candidate")
		Expect(cands).To(HaveLen(2))
	})

	It("surfaces search errors", func() {
		deps := defaultPushDeps()
		deps.searchPRs = func(_ github.Options, _ github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
			return nil, nil, errors.New("boom")
		}
		_, _, err := decidePushTarget(context.Background(), github.Options{}, "feat/new", deps)
		Expect(err).To(MatchError(ContainSubstring("boom")))
	})
})
