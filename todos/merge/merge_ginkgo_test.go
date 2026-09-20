package merge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMerge(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Todos Merge Suite")
}

// recorder captures every write a merge performs, in order, so the specs assert
// on the sequence storage actually sees.
type recorder struct {
	todos.Provider
	writes   []string
	edits    []todos.EditRequest
	states   []todos.StateUpdate
	plans    []string
	comments map[string][]string
	planErr  error
}

func newRecorder() *recorder { return &recorder{comments: map[string][]string{}} }

func (p *recorder) Comment(_ context.Context, todo *types.TODO, body string) error {
	p.writes = append(p.writes, "comment:"+Ref(todo))
	p.comments[Ref(todo)] = append(p.comments[Ref(todo)], body)
	return nil
}

func (p *recorder) Link(_ context.Context, todo *types.TODO, target string, relation types.RelationKind) (*todos.Link, error) {
	p.writes = append(p.writes, fmt.Sprintf("link:%s->%s:%s", Ref(todo), target, relation))
	return &todos.Link{Relation: relation, TargetShortID: target}, nil
}

func (p *recorder) Unlink(context.Context, *types.TODO, string, types.RelationKind) error { return nil }

func (p *recorder) Links(context.Context, *types.TODO) ([]todos.Link, error) { return nil, nil }

func (p *recorder) Delete(_ context.Context, todo *types.TODO) error {
	p.writes = append(p.writes, "delete:"+Ref(todo))
	return nil
}

func (p *recorder) Edit(_ context.Context, todo *types.TODO, edit todos.EditRequest) error {
	p.writes = append(p.writes, "edit:"+Ref(todo))
	p.edits = append(p.edits, edit)
	return nil
}

func (p *recorder) UpdateState(_ context.Context, todo *types.TODO, update todos.StateUpdate) error {
	p.writes = append(p.writes, "state:"+Ref(todo))
	p.states = append(p.states, update)
	return nil
}

func (p *recorder) SavePlanRevision(_ context.Context, todo *types.TODO, markdown, actor string) (*types.TODO, error) {
	if p.planErr != nil {
		return nil, p.planErr
	}
	p.writes = append(p.writes, "plan:"+Ref(todo)+":"+actor)
	p.plans = append(p.plans, markdown)
	return todo, nil
}

// stubAgent answers one canned structured response and records what it was asked.
type stubAgent struct {
	prompt   string
	response string
	closed   bool
}

func (a *stubAgent) GetConfig() clickyai.AgentConfig { return clickyai.AgentConfig{} }

func (a *stubAgent) ExecutePrompt(_ context.Context, req clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	a.prompt = req.Spec.Prompt.User
	return &clickyai.PromptResponse{StructuredData: json.RawMessage(a.response)}, nil
}

func (a *stubAgent) ExecuteBatch(context.Context, []clickyai.PromptRequest) (map[string]*clickyai.PromptResponse, error) {
	return nil, nil
}

func (a *stubAgent) GetCosts() clickyai.Costs { return clickyai.Costs{} }

func (a *stubAgent) Close() error { a.closed = true; return nil }

func todoAt(short, title string, mutate ...func(*types.TODO)) *types.TODO {
	todo := &types.TODO{
		ID:      "0000-" + short,
		ShortID: short,
		TODOFrontmatter: types.TODOFrontmatter{
			Title: title, CWD: "/repo", Priority: types.PriorityLow,
		},
		MarkdownBody: title + " body",
	}
	for _, apply := range mutate {
		apply(todo)
	}
	return todo
}

func targetsFor(provider todos.Provider, todoList ...*types.TODO) []bulk.Target {
	targets := make([]bulk.Target, 0, len(todoList))
	for _, todo := range todoList {
		targets = append(targets, bulk.Target{Ref: todo.ShortID, Provider: provider, Todo: todo})
	}
	return targets
}

const validFixture = "```yaml test\npackages: ./todos\n```"

func proposalJSON(fields map[string]any) string {
	base := map[string]any{
		"title":   "One merged TODO",
		"body":    "Problem.\n\n## Acceptance Criteria\n\n- [ ] merged\n\n## Scope\n\n```\nnothing else\n```",
		"summary": "Folded two TODOs into one.",
	}
	for key, value := range fields {
		base[key] = value
	}
	raw, err := json.Marshal(base)
	Expect(err).NotTo(HaveOccurred())
	return string(raw)
}

func options(agent *stubAgent, mutate ...func(*Options)) Options {
	opts := Options{
		WorkDir: "/repo",
		Base:    api.Spec{Model: api.Model{Name: "api:claude-sonnet-5"}},
		Saved:   captainconfig.Config{AI: captainconfig.AIDefaults{DefaultModel: "api:claude-sonnet-5"}},
		Runtime: captaincli.AIRuntimeOptions{},
		NewAgent: func(captaincli.AIRuntimeResolved) (clickyai.Agent, error) {
			return agent, nil
		},
	}
	for _, apply := range mutate {
		apply(&opts)
	}
	return opts
}

var _ = Describe("Targets", func() {
	It("makes the first TODO the survivor when --into is not given", func() {
		first, second := todoAt("aaa111", "First"), todoAt("bbb222", "Second")
		survivor, retired, err := Targets([]*types.TODO{first, second}, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(survivor).To(Equal(first))
		Expect(retired).To(Equal([]*types.TODO{second}))
	})

	It("honours --into by short id", func() {
		first, second := todoAt("aaa111", "First"), todoAt("bbb222", "Second")
		survivor, retired, err := Targets([]*types.TODO{first, second}, "bbb222")
		Expect(err).NotTo(HaveOccurred())
		Expect(survivor).To(Equal(second))
		Expect(retired).To(Equal([]*types.TODO{first}))
	})

	It("rejects an --into that is not part of the selection", func() {
		_, _, err := Targets([]*types.TODO{todoAt("aaa111", "First"), todoAt("bbb222", "Second")}, "zzz999")
		Expect(err).To(MatchError(ContainSubstring(`--into "zzz999" is not one of the TODOs`)))
	})

	It("refuses fewer than two TODOs", func() {
		_, _, err := Targets([]*types.TODO{todoAt("aaa111", "First")}, "")
		Expect(err).To(MatchError(ContainSubstring("at least two TODOs")))
	})

	It("refuses a selection that spans workspaces", func() {
		other := todoAt("bbb222", "Second", func(t *types.TODO) { t.CWD = "/other" })
		_, _, err := Targets([]*types.TODO{todoAt("aaa111", "First"), other}, "")
		Expect(err).To(MatchError(ContainSubstring("must belong to one workspace")))
	})
})

var _ = Describe("Run", func() {
	var provider *recorder
	var first, second *types.TODO

	BeforeEach(func() {
		provider = newRecorder()
		first = todoAt("aaa111", "First", func(t *types.TODO) { t.Labels = []string{"area:ui"} })
		second = todoAt("bbb222", "Second", func(t *types.TODO) {
			t.Labels = []string{"area:api"}
			t.Priority = types.PriorityHigh
		})
	})

	It("retires each TODO before rewriting the survivor, and links before deleting", func() {
		agent := &stubAgent{response: proposalJSON(nil)}
		result, proposal, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).NotTo(HaveOccurred())
		Expect(proposal.Title).To(Equal("One merged TODO"))
		Expect(provider.writes).To(Equal([]string{
			"comment:bbb222",
			"link:bbb222->aaa111:related_to",
			"delete:bbb222",
			"edit:aaa111",
			"state:aaa111",
			"comment:aaa111",
		}))
		Expect(result.Applied).To(Equal(2))
		Expect(result.Failed).To(BeZero())
		Expect(agent.closed).To(BeTrue())
	})

	It("writes the merged title, body and label union onto the survivor", func() {
		agent := &stubAgent{response: proposalJSON(nil)}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.edits).To(HaveLen(1))
		Expect(*provider.edits[0].Title).To(Equal("One merged TODO"))
		Expect(*provider.edits[0].Body).To(ContainSubstring("## Acceptance Criteria"))
		Expect(*provider.edits[0].Labels).To(Equal([]string{"area:ui", "area:api"}))
		Expect(provider.edits[0].Verification).To(BeNil())
	})

	It("raises the survivor to the highest severity among the sources", func() {
		agent := &stubAgent{response: proposalJSON(nil)}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.states).To(HaveLen(1))
		Expect(*provider.states[0].Priority).To(Equal(types.PriorityHigh))
	})

	It("writes the merged fixture when a source has one", func() {
		second.VerificationMarkdown = validFixture
		agent := &stubAgent{response: proposalJSON(map[string]any{"verification": validFixture})}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).NotTo(HaveOccurred())
		Expect(*provider.edits[0].Verification).To(Equal(validFixture))
	})

	It("refuses a merge that drops a source fixture, writing nothing", func() {
		second.VerificationMarkdown = validFixture
		agent := &stubAgent{response: proposalJSON(nil)}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).To(MatchError(ContainSubstring("dropped the verification fixture")))
		Expect(provider.writes).To(BeEmpty())
	})

	It("refuses an unparseable merged fixture, writing nothing", func() {
		second.VerificationMarkdown = validFixture
		agent := &stubAgent{response: proposalJSON(map[string]any{"verification": "---\nthis: [is not: valid yaml\n---\n"})}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).To(MatchError(ContainSubstring("unparseable verification fixture")))
		Expect(provider.writes).To(BeEmpty())
	})

	It("saves the merged plan as an unapproved revision", func() {
		agent := &stubAgent{response: proposalJSON(map[string]any{"plan": "1. do the thing"})}
		opts := options(agent, func(o *Options) {
			o.Plans = func(_ context.Context, todo *types.TODO) (string, error) {
				if todo == second {
					return "## Plan\n\nold steps", nil
				}
				return "", nil
			}
		})
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.plans).To(Equal([]string{"1. do the thing"}))
		Expect(provider.writes).To(ContainElement("plan:aaa111:merge"))
	})

	It("refuses a merge that drops a source plan, writing nothing", func() {
		agent := &stubAgent{response: proposalJSON(nil)}
		opts := options(agent, func(o *Options) {
			o.Plans = func(context.Context, *types.TODO) (string, error) { return "## Plan\n\nold steps", nil }
		})
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), opts)
		Expect(err).To(MatchError(ContainSubstring("dropped the plan")))
		Expect(provider.writes).To(BeEmpty())
	})

	It("leaves an excluded TODO completely untouched and reports it", func() {
		third := todoAt("ccc333", "Third")
		agent := &stubAgent{response: proposalJSON(map[string]any{"excluded": []string{"ccc333"}})}
		result, _, err := Run(context.Background(), targetsFor(provider, first, second, third), options(agent))
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.writes).NotTo(ContainElement(ContainSubstring("ccc333")))
		var excluded bulk.ItemResult
		for _, item := range result.Results {
			if item.Ref == "ccc333" {
				excluded = item
			}
		}
		Expect(excluded.Status).To(Equal("excluded: not the same work"))
	})

	It("refuses when every other TODO is excluded", func() {
		agent := &stubAgent{response: proposalJSON(map[string]any{"excluded": []string{"bbb222"}})}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).To(MatchError(ContainSubstring("nothing left to merge")))
		Expect(provider.writes).To(BeEmpty())
	})

	It("refuses to exclude the survivor", func() {
		agent := &stubAgent{response: proposalJSON(map[string]any{"excluded": []string{"aaa111"}})}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).To(MatchError(ContainSubstring("which is the TODO being merged into")))
	})

	It("refuses an exclusion naming a TODO outside the selection", func() {
		agent := &stubAgent{response: proposalJSON(map[string]any{"excluded": []string{"zzz999"}})}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).To(MatchError(ContainSubstring("not one of the TODOs being merged")))
	})

	DescribeTable("rejects a proposal that cannot be written", func(fields map[string]any, message string) {
		agent := &stubAgent{response: proposalJSON(fields)}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).To(MatchError(ContainSubstring(message)))
		Expect(provider.writes).To(BeEmpty())
	},
		Entry("no title", map[string]any{"title": ""}, "missing title"),
		Entry("no body", map[string]any{"body": " "}, "missing body"),
		Entry("no summary", map[string]any{"summary": ""}, "missing summary"),
		Entry("bad severity", map[string]any{"priority": "urgent"}, "merge priority"),
	)

	It("writes nothing on a dry run and still reports the proposal", func() {
		agent := &stubAgent{response: proposalJSON(nil)}
		result, proposal, err := Run(context.Background(), targetsFor(provider, first, second),
			options(agent, func(o *Options) { o.DryRun = true }))
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.writes).To(BeEmpty())
		Expect(proposal.Title).To(Equal("One merged TODO"))
		Expect(result.Results).To(HaveLen(2))
		for _, item := range result.Results {
			Expect(item.Status).To(HavePrefix("dry-run:"))
		}
	})

	It("sends every source fixture verbatim and every source plan to the model", func() {
		second.VerificationMarkdown = validFixture
		agent := &stubAgent{response: proposalJSON(map[string]any{"verification": validFixture, "plan": "1. step"})}
		opts := options(agent, func(o *Options) {
			o.Plans = func(_ context.Context, todo *types.TODO) (string, error) {
				return "plan of " + todo.ShortID, nil
			}
		})
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(agent.prompt).To(ContainSubstring("1. First"))
		Expect(agent.prompt).To(ContainSubstring("2. Second"))
		Expect(agent.prompt).To(ContainSubstring(strings.TrimSpace(validFixture)))
		Expect(agent.prompt).To(ContainSubstring("plan of aaa111"))
		Expect(agent.prompt).To(ContainSubstring("plan of bbb222"))
	})

	It("records the rationale on each retired TODO and the summary on the survivor", func() {
		agent := &stubAgent{response: proposalJSON(map[string]any{"rationale": "same parser bug"})}
		_, _, err := Run(context.Background(), targetsFor(provider, first, second), options(agent))
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.comments["bbb222"]).To(ConsistOf(ContainSubstring("Merged into aaa111 — One merged TODO. same parser bug")))
		Expect(provider.comments["aaa111"]).To(ConsistOf(And(
			ContainSubstring("Folded two TODOs into one."),
			ContainSubstring("Merged in: bbb222"),
		)))
	})
})
