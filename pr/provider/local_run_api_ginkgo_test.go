// Tests a local caller's run API for a run on substrate: the run's program asks the caller over a
// host call, and the caller reads the PR and pushes its source branch with its own credentials.
package provider

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/sandbox/runclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// localPR is the caller's own client for one PR, whose source branch is topic of remote.
type localPR struct {
	remote   string
	resolved []string
}

func (p *localPR) Kind() string                    { return "github" }
func (p *localPR) Preflight(context.Context) error { return nil }
func (p *localPR) DefaultBranch(context.Context) (string, error) {
	return "main", nil
}
func (p *localPR) OpenPRs(context.Context) ([]model.PRInfo, error) { return nil, nil }
func (p *localPR) CreatePR(context.Context, model.CreatePRInput) (*model.CreatePRResult, error) {
	return nil, nil
}
func (p *localPR) MergePR(context.Context, int, model.MergeInput) error { return nil }
func (p *localPR) Status(ctx context.Context, number int, _ model.StatusOptions) (*model.StatusSnapshot, error) {
	head, err := targetGit(ctx, p.remote, "rev-parse", "refs/heads/topic")
	if err != nil {
		return nil, err
	}
	return &model.StatusSnapshot{PR: &model.PRInfo{Number: number, State: "OPEN", HeadRefOID: head, HeadRefName: "topic", HeadRepositoryURL: p.remote}}, nil
}
func (p *localPR) ResolveThread(_ context.Context, _ int, thread string) error {
	p.resolved = append(p.resolved, thread)
	return nil
}

var _ = Describe("a local caller's run API for a run on substrate", func() {
	var (
		pr     *localPR
		target *RunTarget
		work   string
		ctx    context.Context
	)
	BeforeEach(func() {
		ctx = context.Background()
		root := GinkgoT().TempDir()
		seed, remote, caller := filepath.Join(root, "seed"), filepath.Join(root, "fork.git"), filepath.Join(root, "caller")
		work = filepath.Join(root, "environment")
		git := func(dir string, args ...string) string {
			out, err := targetGit(ctx, dir, append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com"}, args...)...)
			Expect(err).NotTo(HaveOccurred())
			return out
		}
		Expect(os.MkdirAll(seed, 0o755)).To(Succeed())
		git(seed, "init", "-q", "-b", "topic")
		git(seed, "commit", "-q", "--allow-empty", "-m", "first")
		git(root, "clone", "-q", "--bare", seed, remote)
		git(root, "clone", "-q", remote, caller)
		git(root, "clone", "-q", remote, work)
		git(work, "config", "user.name", "t")
		git(work, "config", "user.email", "t@example.com")
		pr = &localPR{remote: remote}
		head, err := Revision(ctx, work, "HEAD")
		Expect(err).NotTo(HaveOccurred())
		target = &RunTarget{RepositoryURL: remote, Branch: "topic", HeadSHA: head}
		answer := runclient.AnswerHostCall(LocalRunAPI(pr, 7, *target, caller), "run-1", "", runclient.PRReach)
		UseRunAPI(runclient.OverHostCall(func(_ string, data json.RawMessage) (json.RawMessage, error) {
			return answer(ctx, data)
		}))
		DeferCleanup(func() { UseRunAPI(nil) })
	})

	It("reads the PR and resolves its threads with the caller's client", func() {
		client, number, err := ResolveURL(Options{Context: ctx}, "https://github.com/acme/app/pull/7")
		Expect(err).NotTo(HaveOccurred())
		snapshot, err := client.Status(ctx, number, model.StatusOptions{Threads: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.PR.HeadRefOID).To(Equal(target.HeadSHA))
		Expect(client.ResolveThread(ctx, number, "thread-1")).To(Succeed())
		Expect(pr.resolved).To(Equal([]string{"thread-1"}))
	})

	It("pushes the run's commits to the PR's source branch, continuing the head it saw", func() {
		Expect(target.FetchHead(ctx, work)).To(Succeed())
		_, err := targetGit(ctx, work, "commit", "-q", "--allow-empty", "-m", "second")
		Expect(err).NotTo(HaveOccurred())
		second, err := Revision(ctx, work, "HEAD")
		Expect(err).NotTo(HaveOccurred())

		pushed, err := target.PushHead(ctx, work)
		Expect(err).NotTo(HaveOccurred())
		Expect(pushed).To(Equal(second))
		Expect(targetGit(ctx, pr.remote, "rev-parse", "refs/heads/topic")).To(Equal(second), "the caller pushed with its own credentials")

		stale := &RunTarget{RepositoryURL: pr.remote, Branch: "topic", HeadSHA: strings.Repeat("a", 40)}
		_, err = runclient.OverHostCall(func(_ string, data json.RawMessage) (json.RawMessage, error) {
			return runclient.AnswerHostCall(LocalRunAPI(pr, 7, *stale, work), "run-1", "", runclient.PRReach)(ctx, data)
		}).Do(ctx, "POST", "pr/push?base="+stale.HeadSHA, strings.NewReader("not a bundle"))
		Expect(err).To(MatchError("HTTP 409: the PR's source branch is at " + second + ", not " + stale.HeadSHA))
	})
})
