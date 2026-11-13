package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/comments"
)

// RunPreview is the run a target would start, as clicky-ui's PromptRunEditor shows it.
type RunPreview struct {
	Model           string           `json:"model"`
	Mode            string           `json:"mode"`
	Resolution      api.ResolvedSpec `json:"resolution"`
	ValidationError string           `json:"validationError,omitempty"`
	Warnings        []string         `json:"warnings,omitempty"`
	Tools           []PreviewTool    `json:"tools"`
}

// PreviewTool is one tool the session gets: from the thread (caller tools) or declared by the prompt.
type PreviewTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// previewAgent stands in for the agent a run would create, which a preview never claims.
var previewAgent = comments.Actor{Kind: "agent", ID: "gavel:preview", Label: "Gavel agent"}

// Preview renders the run target would start for comment, without claiming or launching.
func (d delivery) Preview(ctx context.Context, comment comments.Comment, target comments.DeliverTarget) (json.RawMessage, error) {
	options, err := decodeRunOptions(target.RunOptions)
	if err != nil {
		return nil, err
	}
	dir, err := d.Workdir(ctx, comment.Target)
	if err != nil {
		return nil, err
	}
	invocation, err := d.invocation(sessionRequest{Comment: comment, Dir: dir, Agent: previewAgent, Prompt: target.Prompt, Options: options})
	if err != nil {
		return nil, err
	}
	if ctx, err = promptRunContext(ctx, dir); err != nil {
		return nil, err
	}
	rendered, err := d.Bridge.Render(ctx, invocation)
	if err != nil {
		return nil, fmt.Errorf("%w: render prompt %s: %v", comments.ErrInvalid, invocation.Reference, err)
	}
	return json.Marshal(previewOf(rendered, invocation.Tools))
}

func previewOf(rendered captaincli.PromptRenderResult, threadTools []api.ToolDefinition) RunPreview {
	tools := []PreviewTool{}
	for _, tool := range threadTools {
		tools = append(tools, PreviewTool{Name: tool.Name, Description: tool.Description, Source: "thread"})
	}
	for _, name := range slices.Sorted(maps.Keys(rendered.Input.Tools)) {
		tools = append(tools, PreviewTool{Name: name, Description: rendered.Input.Tools[name].Description, Source: "prompt"})
	}
	return RunPreview{
		Model: rendered.Model, Mode: rendered.Mode, Resolution: rendered.Resolution,
		ValidationError: rendered.ValidationError, Warnings: rendered.Resolution.Warnings, Tools: tools,
	}
}
