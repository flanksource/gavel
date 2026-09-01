package commit

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	captainapi "github.com/flanksource/captain/pkg/api"
	clickyai "github.com/flanksource/gavel/ai"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/status"
)

type agentLifecycleProbe struct {
	response   *clickyai.PromptResponse
	executeErr error
	closeErr   error
	closeCalls int
}

func (a *agentLifecycleProbe) ExecutePrompt(context.Context, clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	if a.executeErr != nil {
		return nil, a.executeErr
	}
	return a.response, nil
}

func (a *agentLifecycleProbe) ExecuteBatch(context.Context, []clickyai.PromptRequest) (map[string]*clickyai.PromptResponse, error) {
	return nil, errors.New("unexpected batch execution")
}

func (a *agentLifecycleProbe) GetCosts() clickyai.Costs { return nil }
func (a *agentLifecycleProbe) Close() error {
	a.closeCalls++
	return a.closeErr
}

var _ = Describe("commit AI agent lifecycle", func() {
	var previousAgentFactory func(clickyai.AgentConfig) (clickyai.Agent, error)
	var previousCommitAnalyzer func(context.Context, models.CommitAnalysis, clickyai.Agent, gavelgit.AnalyzeOptions) (models.CommitAnalysis, error)
	var previousGatherStatus func(string) (*status.Result, error)
	var previousTestEnv string

	BeforeEach(func() {
		previousAgentFactory = newAgentFunc
		previousCommitAnalyzer = analyzeCommitMessageWithAIFunc
		previousGatherStatus = gatherStatusFunc
		previousTestEnv = os.Getenv(testEnvVar)
		Expect(os.Unsetenv(testEnvVar)).To(Succeed())
	})

	AfterEach(func() {
		newAgentFunc = previousAgentFactory
		analyzeCommitMessageWithAIFunc = previousCommitAnalyzer
		gatherStatusFunc = previousGatherStatus
		if previousTestEnv == "" {
			Expect(os.Unsetenv(testEnvVar)).To(Succeed())
		} else {
			Expect(os.Setenv(testEnvVar, previousTestEnv)).To(Succeed())
		}
	})

	It("rejects a nil agent returned without a factory error", func() {
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return nil, nil }

		agent, err := BuildAgent(Options{}, captainapi.Model{Name: "test-model", Mode: captainapi.ModeAgent})

		Expect(agent).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("agent factory returned nil")))
	})

	It("closes normal commit message generation after success", func() {
		agent := &agentLifecycleProbe{}
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }
		analyzeCommitMessageWithAIFunc = func(context.Context, models.CommitAnalysis, clickyai.Agent, gavelgit.AnalyzeOptions) (models.CommitAnalysis, error) {
			return models.CommitAnalysis{Commit: models.Commit{
				CommitType: models.CommitTypeFix,
				Scope:      models.ScopeTypeReliability,
				Subject:    "close commit agent",
			}}, nil
		}

		analysis, err := generateCommitAnalysis(context.Background(), Options{}, "diff")

		Expect(err).NotTo(HaveOccurred())
		Expect(analysis.Message).To(Equal("fix(reliability): close commit agent"))
		Expect(agent.closeCalls).To(Equal(1))
	})

	It("preserves prompt and close errors together", func() {
		promptErr := errors.New("generate commit message")
		closeErr := errors.New("stop commit agent")
		agent := &agentLifecycleProbe{closeErr: closeErr}
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }
		analyzeCommitMessageWithAIFunc = func(context.Context, models.CommitAnalysis, clickyai.Agent, gavelgit.AnalyzeOptions) (models.CommitAnalysis, error) {
			return models.CommitAnalysis{}, promptErr
		}

		_, err := generateCommitAnalysis(context.Background(), Options{}, "diff")

		Expect(errors.Is(err, promptErr)).To(BeTrue())
		Expect(errors.Is(err, closeErr)).To(BeTrue())
		Expect(agent.closeCalls).To(Equal(1))
	})

	It("closes AI grouping after decoding the response", func() {
		agent := &agentLifecycleProbe{response: &clickyai.PromptResponse{StructuredData: json.RawMessage(`{
			"groups":[{"label":"fix: close grouped agent","files":["commit.go"]}],
			"ignore":[]
		}`)}}
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "commit.go"}}}, nil
		}

		groups, err := groupChangesByAI(context.Background(), Options{MaxCommits: 2}, stagedSource{
			Changes: []stagedChange{{Path: "commit.go", Status: "modified", Adds: 3, Dels: 1}},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(groups).To(HaveLen(1))
		Expect(agent.closeCalls).To(Equal(1))
	})

	It("closes AI grouping when prompt execution fails", func() {
		promptErr := errors.New("group commit files")
		agent := &agentLifecycleProbe{executeErr: promptErr}
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "commit.go"}}}, nil
		}

		_, err := groupChangesByAI(context.Background(), Options{MaxCommits: 2}, stagedSource{
			Changes: []stagedChange{{Path: "commit.go", Status: "modified"}},
		})

		Expect(errors.Is(err, promptErr)).To(BeTrue())
		Expect(agent.closeCalls).To(Equal(1))
	})

	It("closes since-message simplification after success", func() {
		agent := &agentLifecycleProbe{response: &clickyai.PromptResponse{Result: "fix: merge issue commits"}}
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }

		message, err := simplifyGroupMessage(context.Background(), Options{}, []string{"fix: first", "fix: second"})

		Expect(err).NotTo(HaveOccurred())
		Expect(message).To(Equal("fix: merge issue commits"))
		Expect(agent.closeCalls).To(Equal(1))
	})

	It("closes since-message simplification when prompting fails", func() {
		promptErr := errors.New("simplify issue commits")
		agent := &agentLifecycleProbe{executeErr: promptErr}
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }

		_, err := simplifyGroupMessage(context.Background(), Options{}, []string{"fix: first", "fix: second"})

		Expect(errors.Is(err, promptErr)).To(BeTrue())
		Expect(agent.closeCalls).To(Equal(1))
	})

	It("closes PR content generation before a dry-run returns", func() {
		agent := &agentLifecycleProbe{}
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }
		deps := defaultPushDeps()
		deps.defaultBranch = func(github.Options) (string, error) { return "main", nil }
		deps.generatePRPrompt = func(context.Context, clickyai.Agent, PRContentInput) (PRContent, error) {
			return PRContent{Title: "fix: close agent", Branch: "fix/close-agent"}, nil
		}

		err := executeNewPRPush(context.Background(), Options{DryRun: true}, github.Options{}, deps, "fix/close-agent", &Result{
			Commits: []CommitResult{{Message: "fix: close agent"}},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(agent.closeCalls).To(Equal(1))
	})

	It("closes PR content generation when the prompt fails", func() {
		promptErr := errors.New("generate PR content")
		agent := &agentLifecycleProbe{}
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }
		deps := defaultPushDeps()
		deps.defaultBranch = func(github.Options) (string, error) { return "main", nil }
		deps.generatePRPrompt = func(context.Context, clickyai.Agent, PRContentInput) (PRContent, error) {
			return PRContent{}, promptErr
		}

		err := executeNewPRPush(context.Background(), Options{DryRun: true}, github.Options{}, deps, "fix/close-agent", &Result{
			Commits: []CommitResult{{Message: "fix: close agent"}},
		})

		Expect(errors.Is(err, promptErr)).To(BeTrue())
		Expect(agent.closeCalls).To(Equal(1))
	})
})
