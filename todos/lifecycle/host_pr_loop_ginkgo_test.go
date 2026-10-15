package lifecycle

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/commit"
	fixtureverifier "github.com/flanksource/gavel/fixtures/verifier"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos/types"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

type localPRClient struct {
	provider.Client
	remote        string
	info          model.PRInfo
	resolved      []string
	resolvedHeads []string
}

func prLoopGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %v: %w: %s", args, err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func (p *localPRClient) Status(ctx context.Context, _ int, _ model.StatusOptions) (*model.StatusSnapshot, error) {
	head, err := prLoopGit(ctx, p.remote, "rev-parse", "refs/heads/topic")
	if err != nil {
		return nil, err
	}
	pr := p.info
	pr.HeadRefOID = head
	pr.ReviewThreads = append([]model.PRComment(nil), p.info.ReviewThreads...)
	return &model.StatusSnapshot{PR: &pr}, nil
}

func (p *localPRClient) ResolveThread(ctx context.Context, _ int, id string) error {
	status, err := p.Status(ctx, 27, model.StatusOptions{})
	if err != nil {
		return err
	}
	for i := range p.info.ReviewThreads {
		if p.info.ReviewThreads[i].ThreadID != id {
			continue
		}
		p.info.ReviewThreads[i].IsResolved = true
		p.resolved = append(p.resolved, id)
		p.resolvedHeads = append(p.resolvedHeads, status.PR.HeadRefOID)
		return nil
	}
	return fmt.Errorf("unexpected thread %s", id)
}

type prLoopProvider struct{ requests []api.Spec }

func (p *prLoopProvider) GetModel() string        { return "fixture-model" }
func (p *prLoopProvider) GetRuntime() api.Runtime { return api.RuntimeOf(api.OpenAI, api.ModeAgent) }
func (p *prLoopProvider) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, fmt.Errorf("expected streaming")
}
func (p *prLoopProvider) ExecuteStream(_ context.Context, req api.Spec) (<-chan captainai.Event, error) {
	p.requests = append(p.requests, req)
	conclusion, thread := "FAILURE", "addressed"
	if len(p.requests) == 2 {
		conclusion, thread = "SUCCESS", "remaining"
	}
	if err := os.WriteFile(filepath.Join(req.Setup.Cwd, "status.json"), []byte(fmt.Sprintf(`{"check":"parser-test","conclusion":"%s","threadId":"remaining"}`, conclusion)), 0600); err != nil {
		return nil, err
	}
	ch := make(chan captainai.Event, 3)
	ch <- captainai.Event{Kind: captainai.EventSystem, SessionID: "pr-loop-session"}
	ch <- captainai.Event{Kind: captainai.EventText, Text: fmt.Sprintf(`{"summary":"Addressed parser feedback","endStatus":"completed","addressedThreadIds":["%s"]}`, thread)}
	ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
	close(ch)
	return ch, nil
}

var _ = ginkgo.Describe("PR turns through Captain", func() {
	ginkgo.It("commits and pushes before fixture verification, then retries with status and native thread feedback", func(ctx ginkgo.SpecContext) {
		root := ginkgo.GinkgoT().TempDir()
		remote, work := filepath.Join(root, "source.git"), filepath.Join(root, "main")
		git := func(dir string, args ...string) string {
			ginkgo.GinkgoHelper()
			value, err := prLoopGit(ctx, dir, args...)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			return value
		}
		git(root, "init", "--bare", remote)
		git(root, "clone", remote, work)
		git(work, "config", "user.name", "Fixture User")
		git(work, "config", "user.email", "fixture@example.com")
		gomega.Expect(os.WriteFile(filepath.Join(work, "status.json"), []byte(`{"conclusion":"initial"}`), 0600)).To(gomega.Succeed())
		git(work, "add", "status.json")
		git(work, "commit", "-m", "test: initial PR fixture")
		git(work, "push", "origin", "HEAD:refs/heads/topic")
		initial := git(work, "rev-parse", "HEAD")
		gomega.Expect(os.WriteFile(filepath.Join(work, "user.txt"), []byte("Unrelated user work"), 0600)).To(gomega.Succeed())
		client := &localPRClient{remote: remote, info: model.PRInfo{Number: 27, State: "OPEN", URL: "https://github.com/acme/service/pull/27", HeadRepositoryURL: remote, HeadRefName: "topic", HeadRefOID: initial, ReviewThreads: []model.PRComment{{ThreadID: "addressed", CanResolve: true}, {ThreadID: "remaining", CanResolve: true}}}}
		target, err := provider.TargetOf(&client.info)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		state := &prExecution{client: client, number: 27, target: target, snapshot: &model.StatusSnapshot{PR: &client.info}, openThreads: openThreads(&model.StatusSnapshot{PR: &client.info}), workDir: work, write: true}
		prepared := &preparedStep{class: types.ModeRun, request: api.Spec{Prompt: api.Prompt{User: "Fix the parser."}}}
		gomega.Expect(state.configure(prepared)).To(gomega.Succeed())
		prepared.request.Setup.Checkout.Worktree.Path = filepath.Join(root, "run")
		prepared.request.Workflow.Commits[0].Message = "fix: address PR feedback"
		fixture, err := fixtureverifier.NewVerifier("| Name | Command | Exit Code | CEL Validation |\n|---|---|---|---|\n| PR CI feedback at pushed head | git --git-dir='" + remote + "' show refs/heads/topic:status.json | 0 | !stdout.contains('FAILURE') |")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		fixture.Timeout = 10 * time.Second
		hooks := commit.AgentHooks(commit.AgentHooksOptions{Commits: prepared.request.Workflow.Commits, Meta: commit.AgentRunMetadata{IssueID: "pr-loop-issue", SessionID: "pr-loop-session"}, PRTarget: target, ShouldCommit: completedPRTurn})
		hooks = append(hooks, state, &capsetup.Plugin{BaseDir: work}, capverify.New("fixture", fixture))
		modelProvider := &prLoopProvider{}
		runner := agent.Runner[any]{Provider: modelProvider, Request: prepared.request, Hooks: hooks, MaxIterations: 2, Repo: work, Cwd: work}
		result, err := runner.Run(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(agent.VerifyPassed(result.Verdicts)).To(gomega.BeTrue())
		gomega.Expect(result.Verdicts).To(gomega.HaveLen(2))
		gomega.Expect(result.Verdicts[0].Report.Passed).To(gomega.BeFalse())
		gomega.Expect(modelProvider.requests).To(gomega.HaveLen(2))
		first, second := modelProvider.requests[0].Prompt.User, modelProvider.requests[1].Prompt.User
		gomega.Expect(second).To(gomega.HavePrefix(first), "a retry extends the turn-1 prompt with feedback only; nothing rewrites it")
		gomega.Expect(strings.TrimPrefix(second, first)).To(gomega.ContainSubstring(`"conclusion":"FAILURE"`))
		gomega.Expect(second).NotTo(gomega.ContainSubstring(`"reviewThreads"`))
		gomega.Expect(client.resolved).To(gomega.Equal([]string{"addressed", "remaining"}))
		gomega.Expect(client.resolvedHeads[0]).NotTo(gomega.Equal(initial))
		gomega.Expect(client.resolvedHeads[1]).NotTo(gomega.Equal(client.resolvedHeads[0]))
		gomega.Expect(git(remote, "log", "-2", "--format=%s", "refs/heads/topic")).To(gomega.Equal("fix: address PR feedback\nfix: address PR feedback"))
		gomega.Expect(git(remote, "log", "-1", "--format=%B", "refs/heads/topic")).To(gomega.ContainSubstring("Gavel-Issue-Id: pr-loop-issue"))
		gomega.Expect(git(work, "rev-parse", "HEAD")).To(gomega.Equal(initial))
		gomega.Expect(git(work, "status", "--porcelain")).To(gomega.Equal("?? user.txt"))
	})
})
