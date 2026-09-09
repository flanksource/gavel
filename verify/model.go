package verify

import "github.com/flanksource/captain/pkg/api"

// ModelFor resolves which model one AI operation runs on, highest precedence
// first:
//
//  1. the CLI override (--ai-model / --model), which carries captain's compact
//     grammar, so "api:haiku" outranks a configured agent mode;
//  2. the operation's own spec (commit.message, status.summary, …);
//  3. the base ai: spec;
//  4. whatever ~/.captain.yaml supplies, applied later at resolve time.
//
// It selects but does not resolve — captain resolves the winner against the
// catalog exactly once, on the way to building the provider.
//
// override is an already-expanded api.Model rather than a name or a flag struct:
// a bare name cannot carry the mode or effort the compact grammar just parsed,
// and both flag surfaces (ai.BindFlags' --ai-model and aiflags.ModelFlags)
// expand before they reach here. An empty override contributes nothing.
//
// This is the only implementation of the ladder. It previously existed only
// inside the commit package, so every other AI command (git analyze, git amend,
// status --ai, test outline --ai-summary) skipped .gavel.yaml entirely and built
// its agent from a hardcoded ai.DefaultConfig(), which is why an explicit
// --ai-model was silently discarded.
func ModelFor(base api.Spec, op PromptSpec, override api.Model) api.Model {
	return base.Merge(op.Spec).Model.Merge(override)
}

// ModelFor resolves op's model against this config's ai: base. Callers holding a
// whole GavelConfig use this; the commit package carries the base spec on its own
// Options and calls the free function directly.
func (c GavelConfig) ModelFor(op PromptSpec, override api.Model) api.Model {
	return ModelFor(c.AI, op, override)
}
