package editor

import (
	"context"
	"fmt"
	"net/http"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/comments/agents"
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
	prompts, err := captaincli.DiscoverPrompts(captaincli.ContextWithPromptSources(ctx, agents.PromptSources), dir, captaincli.PromptListOptions{})
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
		DefaultPrompt: managedPrompt, HostVariables: agents.HostVariables, Dir: dir,
	}, nil
}

func nonNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
