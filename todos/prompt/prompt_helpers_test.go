package prompt

import (
	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

func renderResolvedForTest(todos []*types.TODO, options Options) (captainai.Request, captainai.Config, error) {
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
