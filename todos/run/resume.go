package run

import (
	"context"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
)

// ModelProvider names the family a spec's model belongs to. It reads the
// resolved descriptor when there is one and derives it from the name otherwise
// — the provider is a resolution result, never something the spec authored.
func ModelProvider(model api.Model) (string, error) {
	if model.Provider != nil {
		return model.Provider.Name, nil
	}
	provider, err := api.ProviderFor(model.Name)
	if err != nil {
		return "", err
	}
	return provider.Name, nil
}

// ProviderKey is ModelProvider for display: empty when the model names no
// family, which a response renders as "no provider".
func ProviderKey(model api.Model) string {
	provider, err := ModelProvider(model)
	if err != nil {
		return ""
	}
	return provider
}

// ContinuesPrior reports whether a run on model can continue prior's
// conversation: a resume hands the prior provider session id to the new agent,
// which only an agent of the same provider family can open. There is nothing
// to continue without a prior run.
func ContinuesPrior(prior *captaindb.PromptRun, model api.Model) (bool, error) {
	if prior == nil {
		return false, nil
	}
	requested, err := ModelProvider(model)
	if err != nil {
		return false, fmt.Errorf("resolve the provider of the requested runtime: %w", err)
	}
	return todos.SameProviderFamily(prior.Runtime.Resolved.Provider, requested)
}

// Revise resolves the options a reviewed plan is revised with. It resumes the
// plan's conversation with the feedback as the next turn when there is a
// session to resume and the requested runtime continues it; otherwise the plan
// step runs afresh — the returned options do not resume — and the caller must
// hand the feedback over in the todo's own prompt (RevisionPrompt).
func Revise(ctx context.Context, c Continuation, feedback string) (Options, error) {
	if PriorSessionID(c.Todo) != "" {
		c.Resume, c.Message = true, feedback
		opts, err := Continue(c)
		if err != nil {
			return Options{}, err
		}
		prepared, err := Resolve(ctx, Request{Provider: c.Provider, Todo: c.Todo, Dir: c.Dir, Options: opts})
		if err != nil {
			return Options{}, err
		}
		continues, err := ContinuesPrior(c.Prior, prepared.Resolution.Spec.Model)
		if err != nil {
			return Options{}, err
		}
		if continues {
			return opts, nil
		}
	}
	c.Resume, c.Message = false, ""
	return Continue(c)
}

// RevisionPrompt is the todo prompt a fresh revision plans with.
func RevisionPrompt(feedback string) string {
	return "Revise the existing plan using this reviewer feedback:\n\n" + feedback
}
