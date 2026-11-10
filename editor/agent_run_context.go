package editor

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/comments"
)

// agentRunContext is what the new-session dialog chooses from for a file's project.
type agentRunContext struct {
	Prompts       []captaincli.PromptSummary `json:"prompts"`
	Presets       []runtimeprofiles.Preset   `json:"presets"`
	DefaultPrompt string                     `json:"defaultPrompt"`
	HostVariables []string                   `json:"hostVariables"`
	Dir           string                     `json:"dir"`
}

// agentRunContextOf serves GET /api/file/{id}/agent-run/context: the prompts a
// run in the file's project resolves (as a run does), and the runtime presets.
func (s *Server) agentRunContextOf(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Read(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	target, err := targetOf(r.Context(), doc.Path)
	if err != nil {
		respond(w, nil, err)
		return
	}
	runContext, err := discoverRunContext(r.Context(), target.Dir)
	respond(w, runContext, err)
}

func discoverRunContext(ctx context.Context, dir string) (agentRunContext, error) {
	prompts, err := captaincli.DiscoverPrompts(captaincli.ContextWithPromptSources(ctx, librarySources), dir, captaincli.PromptListOptions{})
	if err != nil {
		return agentRunContext{}, fmt.Errorf("discover prompts for %s: %w", dir, err)
	}
	catalog, err := runtimeprofiles.NewDefaultCatalog(ctx, runtimeprofiles.DefaultCatalogOptions{Cwd: dir})
	if err != nil {
		return agentRunContext{}, fmt.Errorf("open the runtime preset catalog for %s: %w", dir, err)
	}
	presets, err := catalog.ListPresets(ctx)
	if err != nil {
		return agentRunContext{}, fmt.Errorf("list runtime presets for %s: %w", dir, err)
	}
	return agentRunContext{
		Prompts: nonNil(prompts), Presets: nonNil(presets),
		DefaultPrompt: managedPrompt, HostVariables: hostVariables, Dir: dir,
	}, nil
}

func nonNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

// runPreview is the run a request would start, as clicky-ui's PromptRunEditor shows it.
type runPreview struct {
	Model           string           `json:"model"`
	Mode            string           `json:"mode"`
	Resolution      api.ResolvedSpec `json:"resolution"`
	ValidationError string           `json:"validationError,omitempty"`
	Warnings        []string         `json:"warnings,omitempty"`
	Tools           []previewTool    `json:"tools"`
}

// previewTool is one tool the session gets: from the thread (the editor's
// caller tools) or declared by the prompt.
type previewTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// previewAgent stands in for the agent a run would create, which a preview never claims.
var previewAgent = comments.Actor{Kind: "agent", ID: "gavel-edit:preview", Label: "Editor agent"}

// previewRun serves POST /api/file/{id}/comments/{cid}/run/preview: the run
// body /run would start, rendered without claiming the thread or launching.
func (s *Server) previewRun(w http.ResponseWriter, r *http.Request) {
	options, ok := decodeRunOptions(w, r)
	if !ok {
		return
	}
	preview, err := s.previewThread(r.Context(), runRequest{FileID: r.PathValue("id"), CommentID: r.PathValue("cid"), Options: options})
	var failed sendError
	if errors.As(err, &failed) {
		http.Error(w, failed.Error(), failed.status)
		return
	}
	respond(w, preview, err)
}

func (s *Server) previewThread(ctx context.Context, req runRequest) (runPreview, error) {
	doc, service, comment, err := s.fileComment(ctx, req.FileID, req.CommentID)
	if err != nil {
		return runPreview{}, err
	}
	target, err := targetOf(ctx, doc.Path)
	if err != nil {
		return runPreview{}, err
	}
	invocation, err := sessionInvocation(sessionRequest{Service: service, ThreadID: comment.ID, Target: target, Agent: previewAgent, Message: sessionMessage(comment, doc.Path, req.Options.Instructions), Options: req.Options})
	if err != nil {
		return runPreview{}, err
	}
	if ctx, err = promptRunContext(ctx, target.Dir); err != nil {
		return runPreview{}, err
	}
	rendered, err := s.Agents.render(ctx, invocation)
	var failed sendError
	if errors.As(err, &failed) {
		return runPreview{}, err
	}
	if err != nil {
		return runPreview{}, sendError{http.StatusBadRequest, fmt.Errorf("render prompt %s: %w", invocation.Reference, err)}
	}
	return previewOf(rendered, invocation.Tools), nil
}

func previewOf(rendered captaincli.PromptRenderResult, threadTools []api.ToolDefinition) runPreview {
	tools := []previewTool{}
	for _, tool := range threadTools {
		tools = append(tools, previewTool{Name: tool.Name, Description: tool.Description, Source: "thread"})
	}
	for _, name := range slices.Sorted(maps.Keys(rendered.Input.Tools)) {
		tools = append(tools, previewTool{Name: name, Description: rendered.Input.Tools[name].Description, Source: "prompt"})
	}
	return runPreview{
		Model: rendered.Model, Mode: rendered.Mode, Resolution: rendered.Resolution,
		ValidationError: rendered.ValidationError, Warnings: rendered.Resolution.Warnings, Tools: tools,
	}
}
