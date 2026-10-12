// Package merge folds several TODOs into one.
//
// It is a one-shot prompt, not a lifecycle step: the model is asked once for the
// combined title, body, verification fixture and plan, and gavel performs every
// write. A lifecycle run is one todo, one run, one status transition (see
// todos/run.Request) — merge is N into 1, so it records no run and starts no
// session.
//
// The agent proposes and gavel writes, the same division triage uses: that is
// what makes the result validatable, so a fixture that does not parse fails
// before anything is written instead of after half the merge has landed.
package merge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// Options configures a merge.
type Options struct {
	// WorkDir is the workspace the merged TODOs belong to. Prompt file overrides
	// and the todos' own path references resolve against it.
	WorkDir string
	// Into is the reference of the TODO to merge into. Empty means the first
	// target supplied.
	Into string
	// DryRun renders and validates the proposal without writing anything.
	DryRun bool
	// Base is the .gavel.yaml ai: spec; Override is todos.merge; Request is the
	// caller's explicit spec (--model/--effort), folded as the top layer.
	Base     api.Spec
	Override verify.PromptSpec
	Request  api.Spec
	Saved    captainconfig.Config
	Runtime  captaincli.AIRuntimeOptions
	// NewAgent builds the agent for the resolved runtime. Nil uses the
	// middleware-wrapped captain agent; tests substitute their own.
	NewAgent func(captaincli.AIRuntimeResolved) (clickyai.Agent, error)
	// Plans resolves a TODO's existing plan. Nil means no plan context, which is
	// what a provider without plan storage supplies.
	Plans func(ctx context.Context, todo *types.TODO) (string, error)
}

// Proposal is the model's merge, decoded and validated. Every field is content
// gavel writes; nothing here is executed by the agent.
type Proposal struct {
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Verification string   `json:"verification,omitempty"`
	Plan         string   `json:"plan,omitempty"`
	Priority     string   `json:"priority,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Summary      string   `json:"summary"`
	Rationale    string   `json:"rationale,omitempty"`
	Excluded     []string `json:"excluded,omitempty"`
}

// Selection is the resolved shape of a merge: which TODO survives, which are
// retired into it, and which the model refused to merge.
type Selection struct {
	Survivor *types.TODO
	Retired  []*types.TODO
	Excluded []*types.TODO
}

// Plan reads each TODO's durable plan through the provider, for the prompt's
// plan context. A provider without plan storage contributes nothing.
func Plan(provider todos.Provider) func(context.Context, *types.TODO) (string, error) {
	plans, ok := provider.(todos.PlanContentProvider)
	if !ok {
		return nil
	}
	return func(ctx context.Context, todo *types.TODO) (string, error) {
		return plans.PlanMarkdown(ctx, todo, types.ModePlan)
	}
}

// Targets validates the selection and orders it: the survivor first, then the
// TODOs retired into it.
//
// Every TODO must belong to one workspace. A cross-workspace merge would write
// a body about another repository's code and link across two databases, and the
// link would resolve in neither.
func Targets(todoList []*types.TODO, into string) (*types.TODO, []*types.TODO, error) {
	if len(todoList) < 2 {
		return nil, nil, fmt.Errorf("merge needs at least two TODOs; got %d", len(todoList))
	}
	workspaces := map[string][]string{}
	for _, todo := range todoList {
		workspaces[strings.TrimSpace(todo.CWD)] = append(workspaces[strings.TrimSpace(todo.CWD)], Ref(todo))
	}
	if len(workspaces) > 1 {
		var groups []string
		for dir, refs := range workspaces {
			groups = append(groups, fmt.Sprintf("%s: %s", dirName(dir), strings.Join(refs, ", ")))
		}
		return nil, nil, fmt.Errorf("every merged TODO must belong to one workspace; got %s", strings.Join(groups, "; "))
	}

	index := 0
	if ref := strings.TrimSpace(into); ref != "" {
		index = -1
		for i, todo := range todoList {
			if matches(todo, ref) {
				index = i
				break
			}
		}
		if index < 0 {
			return nil, nil, fmt.Errorf("--into %q is not one of the TODOs being merged", into)
		}
	}
	survivor := todoList[index]
	retired := make([]*types.TODO, 0, len(todoList)-1)
	for i, todo := range todoList {
		if i != index {
			retired = append(retired, todo)
		}
	}
	return survivor, retired, nil
}

// Propose asks the model for the merged content. It writes nothing.
func Propose(ctx context.Context, survivor *types.TODO, retired []*types.TODO, opts Options) (proposal *Proposal, err error) {
	todoList := append([]*types.TODO{survivor}, retired...)
	data, err := opts.templateData(ctx, todoList)
	if err != nil {
		return nil, err
	}
	runtime, err := opts.resolvePrompt(data)
	if err != nil {
		return nil, fmt.Errorf("render todos merge prompt: %w", err)
	}
	for _, warning := range runtime.Resolution.Warnings {
		logger.Warnf("todos merge: %s", warning)
	}

	newAgent := opts.NewAgent
	if newAgent == nil {
		newAgent = func(resolved captaincli.AIRuntimeResolved) (clickyai.Agent, error) {
			return clickyai.NewAgent(resolved.Config)
		}
	}
	agent, err := newAgent(runtime)
	if err != nil {
		return nil, fmt.Errorf("create todos merge agent: %w", err)
	}
	defer func() {
		if closeErr := agent.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close todos merge agent: %w", closeErr))
		}
	}()

	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name: fmt.Sprintf("todos merge: %d todos", len(todoList)),
		Spec: runtime.Request,
	})
	if err != nil {
		return nil, fmt.Errorf("execute todos merge prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("todos merge prompt returned error: %s", resp.Error)
	}
	proposal = &Proposal{}
	if err := clickyai.DecodeStructured(resp, proposal); err != nil {
		return nil, fmt.Errorf("decode todos merge response: %w", err)
	}
	return proposal, nil
}

// templateData renders the prompt variables: the numbered TODO sections with
// every fixture verbatim (the merge rewrites them, so the command projection
// would hide exactly what is being rewritten), plus each TODO's existing plan.
func (opts Options) templateData(ctx context.Context, todoList []*types.TODO) (map[string]any, error) {
	plans, err := opts.renderPlans(ctx, todoList)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"count": len(todoList),
		"body":  todoprompt.Sections(todoList, opts.WorkDir, true),
		"plans": plans,
	}, nil
}

func (opts Options) renderPlans(ctx context.Context, todoList []*types.TODO) (string, error) {
	if opts.Plans == nil {
		return "", nil
	}
	var rendered strings.Builder
	for _, todo := range todoList {
		plan, err := opts.Plans(ctx, todo)
		if err != nil {
			return "", fmt.Errorf("read the plan of %s: %w", Ref(todo), err)
		}
		if strings.TrimSpace(plan) == "" {
			continue
		}
		fmt.Fprintf(&rendered, "### Plan of %s — %s\n\n%s\n\n", Ref(todo), todo.Title, strings.TrimSpace(plan))
	}
	return rendered.String(), nil
}

// Validate holds the proposal to the contract before a single write happens.
//
// The two "required when a source has one" rules are the point: a merged TODO
// that silently lost the fixture or the plan of one of its sources claims the
// whole scope while being able to prove — and to implement — only part of it.
func (p *Proposal) Validate(survivor *types.TODO, retired []*types.TODO, hasFixture, hasPlan bool, workDir string) (*Selection, error) {
	if p == nil {
		return nil, fmt.Errorf("merge finished without a proposal")
	}
	for field, value := range map[string]string{"title": p.Title, "body": p.Body, "summary": p.Summary} {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("merge proposal is missing %s", field)
		}
	}
	if raw := strings.TrimSpace(p.Priority); raw != "" {
		if err := types.ValidatePriority(types.Priority(raw)); err != nil {
			return nil, fmt.Errorf("merge priority: %w", err)
		}
	}
	selection, err := p.partition(survivor, retired)
	if err != nil {
		return nil, err
	}
	if hasFixture && strings.TrimSpace(p.Verification) == "" {
		return nil, fmt.Errorf("merge proposal dropped the verification fixture: %s has one, so the merged TODO needs one",
			Ref(fixtureSource(append([]*types.TODO{selection.Survivor}, selection.Retired...))))
	}
	if fixture := strings.TrimSpace(p.Verification); fixture != "" {
		if _, err := todos.ParseVerificationMarkdown(todos.VerificationMarkdownOptions{
			Name: "verification", Markdown: fixture, SourceDir: workDir,
		}); err != nil {
			return nil, fmt.Errorf("merge produced an unparseable verification fixture: %w", err)
		}
	}
	if hasPlan && strings.TrimSpace(p.Plan) == "" {
		return nil, fmt.Errorf("merge proposal dropped the plan: a merged TODO has one, so the merged TODO needs one")
	}
	return selection, nil
}

// partition splits the selection by the model's exclusions. An excluded TODO is
// untouched — not retired, not edited — and the survivor cannot be excluded:
// the merge is written onto it.
func (p *Proposal) partition(survivor *types.TODO, retired []*types.TODO) (*Selection, error) {
	selection := &Selection{Survivor: survivor}
	excluded := map[*types.TODO]bool{}
	for _, ref := range p.Excluded {
		if ref = strings.TrimSpace(ref); ref == "" {
			continue
		}
		if matches(survivor, ref) {
			return nil, fmt.Errorf("merge excluded %s, which is the TODO being merged into; re-run with --into naming a different survivor", ref)
		}
		found := false
		for _, todo := range retired {
			if matches(todo, ref) {
				excluded[todo] = true
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("merge excluded %q, which is not one of the TODOs being merged", ref)
		}
	}
	for _, todo := range retired {
		if excluded[todo] {
			selection.Excluded = append(selection.Excluded, todo)
			continue
		}
		selection.Retired = append(selection.Retired, todo)
	}
	if len(selection.Retired) == 0 {
		return nil, fmt.Errorf("merge excluded every other TODO: nothing left to merge into %s", Ref(survivor))
	}
	return selection, nil
}

// Ref names a TODO in a message, preferring the short id because that is what
// the CLI accepts and what a dashboard row shows.
func Ref(todo *types.TODO) string {
	if todo == nil {
		return ""
	}
	if short := strings.TrimSpace(todo.ShortID); short != "" {
		return short
	}
	return todos.TODOReference(todo)
}

func matches(todo *types.TODO, ref string) bool {
	if todo == nil {
		return false
	}
	for _, candidate := range []string{todo.ShortID, todo.ID, todo.Title, todos.TODOReference(todo)} {
		if candidate != "" && strings.EqualFold(strings.TrimSpace(candidate), strings.TrimSpace(ref)) {
			return true
		}
	}
	return false
}

func fixtureSource(todoList []*types.TODO) *types.TODO {
	for _, todo := range todoList {
		if strings.TrimSpace(todo.VerificationMarkdown) != "" {
			return todo
		}
	}
	return nil
}

func dirName(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return "(no workspace)"
	}
	return dir
}
