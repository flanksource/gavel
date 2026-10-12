package triagenew

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

type Proposal struct {
	Result        types.TriageNewEnvelope `json:"result"`
	SourceID      string                  `json:"sourceId"`
	SourceTitle   string                  `json:"sourceTitle"`
	SourceVersion int64                   `json:"sourceVersion"`
	TargetID      string                  `json:"targetId"`
	TargetTitle   string                  `json:"targetTitle"`
	TargetVersion int64                   `json:"targetVersion"`
}

type Review struct {
	Provider  todos.Provider
	SourceID  string
	Approved  func(context.Context, map[string]any) error
	mu        sync.Mutex
	proposals map[string]Proposal
	approved  map[string]bool
}

func (r *Review) prepare(ctx context.Context, input map[string]any) (any, error) {
	var env types.TriageNewEnvelope
	if err := decode(input, &env); err != nil {
		return nil, err
	}
	if env.Action == "keep" || env.ProposalID != "" {
		return nil, fmt.Errorf("prepare requires a new relationship proposal")
	}
	source, err := r.Provider.Get(ctx, r.SourceID)
	if err != nil {
		return nil, err
	}
	target, err := r.Provider.Get(ctx, env.Target)
	if err != nil {
		return nil, err
	}
	env.Target = target.ID
	writer, ok := r.Provider.(todos.TriageNewProvider)
	if !ok {
		return nil, fmt.Errorf("provider cannot validate triage.new proposals")
	}
	if err := writer.ValidateTriageNew(ctx, source, &env); err != nil {
		return nil, err
	}
	env.ProposalID = uuid.NewString()
	proposal := Proposal{Result: env, SourceID: source.ID, SourceTitle: source.Title, SourceVersion: source.Version, TargetID: target.ID, TargetTitle: target.Title, TargetVersion: target.Version}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.proposals == nil {
		r.proposals = map[string]Proposal{}
		r.approved = map[string]bool{}
	}
	r.proposals[env.ProposalID] = proposal
	return proposal, nil
}

func (r *Review) review(ctx context.Context, input map[string]any) (any, error) {
	var proposal Proposal
	if err := decode(input, &proposal); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.proposals[proposal.Result.ProposalID]
	if !ok || !reflect.DeepEqual(stored, proposal) {
		return nil, fmt.Errorf("review must match the prepared proposal exactly")
	}
	if r.Approved == nil {
		return nil, fmt.Errorf("triage.new requires durable approval")
	}
	if err := r.Approved(ctx, input); err != nil {
		return nil, err
	}
	r.approved[proposal.Result.ProposalID] = true
	return proposal.Result, nil
}

func (r *Review) Validate(ctx context.Context, todo *types.TODO, env *types.TriageNewEnvelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := env.Validate(); err != nil {
		return err
	}
	if env.EndStatus != types.EndCompleted {
		return nil
	}
	writer, ok := r.Provider.(todos.TriageNewProvider)
	if !ok {
		return fmt.Errorf("provider cannot apply triage.new")
	}
	if env.Action != "keep" {
		r.mu.Lock()
		proposal, exists := r.proposals[env.ProposalID]
		approved := r.approved[env.ProposalID]
		r.mu.Unlock()
		if !exists || !approved || !reflect.DeepEqual(proposal.Result, *env) {
			return fmt.Errorf("triage.new relationship has no matching approved proposal")
		}
		for _, snapshot := range []struct {
			id      string
			version int64
		}{{proposal.SourceID, proposal.SourceVersion}, {proposal.TargetID, proposal.TargetVersion}} {
			current, err := r.Provider.Get(ctx, snapshot.id)
			if err != nil {
				return err
			}
			if current.Version != snapshot.version {
				return fmt.Errorf("TODO %s changed since review; triage again", snapshot.id)
			}
			if current.ID == todo.ID {
				*todo = *current
			}
		}
	}
	if err := writer.ValidateTriageNew(ctx, todo, env); err != nil {
		return err
	}
	return nil
}

func (r *Review) Apply(ctx context.Context, todo *types.TODO, env *types.TriageNewEnvelope) error {
	if err := r.Validate(ctx, todo, env); err != nil {
		return err
	}
	if env.EndStatus != types.EndCompleted {
		return nil
	}
	opts := todos.TriageNewApplyOptions{}
	if env.Action != "keep" {
		r.mu.Lock()
		opts.TargetVersion = r.proposals[env.ProposalID].TargetVersion
		r.mu.Unlock()
	}
	return r.Provider.(todos.TriageNewProvider).ApplyTriageNew(ctx, todo, env, opts)
}

func decode(input map[string]any, target any) error {
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func (r *Review) Tools() []api.ToolDefinition {
	return []api.ToolDefinition{
		{Name: "todo_search", Description: "Search open TODOs in this workspace by title or body. Empty query lists the backlog.", DefaultPermission: api.ToolPolicyAllow, InputSchema: map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string"}}}, Handler: r.search},
		{Name: "todo_get", Description: "Read a workspace TODO, including complete description, criteria, fixture and plan.", DefaultPermission: api.ToolPolicyAllow, InputSchema: map[string]any{"type": "object", "properties": map[string]any{"ref": map[string]any{"type": "string"}}, "required": []string{"ref"}}, Handler: r.get},
		{Name: "triage_prepare", Description: "Validate a relationship proposal without applying it and return the complete proposal for human review.", DefaultPermission: api.ToolPolicyAllow, InputSchema: schema(&types.TriageNewEnvelope{}), Handler: r.prepare},
		{Name: "triage_review", Description: "Request human approval of this exact prepared relationship proposal. No changes are applied by this tool.", DefaultPermission: api.ToolPolicyAsk, InputSchema: schema(&Proposal{}), Handler: r.review},
	}
}

func schema(value any) map[string]any {
	raw, err := api.SchemaJSON(value)
	if err != nil {
		panic(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		panic(err)
	}
	return schema
}
