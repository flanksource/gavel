package prompt

import (
	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// stepEvents are the built-in lifecycle steps' events (todos/lifecycle/todos.yaml),
// for specs that render a prompt the way its step does. The schema contract
// itself is pinned against the lifecycle in todos/lifecycle/host_envelope_schema_ginkgo_test.go.
var stepEvents = map[string][]string{
	"run":        {"completed", "blocked"},
	"plan":       {"new", "updated", "unchanged"},
	"triage":     {"ready", "shape", "investigate", "done", "retired", "merged", "duplicate"},
	"triage.new": {"kept", "duplicate", "merged", "child"},
}

func renderResolvedForTest(todos []*types.TODO, options Options) (captainai.Request, captainai.Config, error) {
	if options.Events == nil {
		options.Events = stepEvents[options.promptName()]
	}
	template, err := templateSource(options)
	if err != nil {
		return captainai.Request{}, captainai.Config{}, err
	}
	resolved, err := (verify.PromptSpec{}).Resolve(verify.PromptResolveOptions{
		DefaultPrompt: template, Data: TemplateData(todos, options), Request: options.Spec,
	})
	if err != nil {
		return captainai.Request{}, captainai.Config{}, err
	}
	options.Spec = resolved.Spec
	return Render(todos, options)
}
