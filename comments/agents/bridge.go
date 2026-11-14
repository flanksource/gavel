// Package agents delivers comment threads to Claude and Codex sessions through
// Captain: into a live session, or into a new managed session started from a prompt.
package agents

import (
	"context"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/session/deliver"
	"github.com/flanksource/gavel/internal/database"
	promptlibrary "github.com/flanksource/gavel/prompts/library"
)

// Bridge reaches Captain's sessions; Captain is the in-process implementation.
type Bridge interface {
	Live(context.Context) ([]captaincli.SessionRecord, error)
	Deliver(ctx context.Context, sessionID, text string) (deliver.Result, error)
	// Launch starts a managed session from a prompt, as captaincli.LaunchPrompt.
	Launch(context.Context, captaincli.PromptInvocation) (captaincli.PromptRunResult, error)
	// Render resolves a session's prompt without starting it, as captaincli.RenderPrompt.
	Render(context.Context, captaincli.PromptInvocation) (captaincli.PromptRenderResult, error)
}

// PromptSources scopes Captain's prompt resolution to gavel's prompt library.
var PromptSources = captaincli.PromptSources{FS: []captaincli.PromptFS{{Label: promptlibrary.Label, FS: promptlibrary.FS}}}

// OpenCaptain opens gavel's process database first, which hands its pool to
// captain: a captain call that ran first would open a second pool on its own,
// and captain then refuses gavel's, failing every later comments request.
func OpenCaptain(ctx context.Context) error {
	_, err := database.Require(ctx, "Agent sessions")
	return err
}

// Captain is the Bridge to Captain in this process, over gavel's database.
type Captain struct{}

var _ Bridge = Captain{}

func (Captain) Live(ctx context.Context) ([]captaincli.SessionRecord, error) {
	if err := OpenCaptain(ctx); err != nil {
		return nil, err
	}
	result, err := captaincli.RunSessionLive(ctx, captaincli.SessionLiveOptions{Source: "all", All: true, Limit: 100})
	return result.Sessions, err
}

func (Captain) Deliver(ctx context.Context, sessionID, text string) (deliver.Result, error) {
	if err := OpenCaptain(ctx); err != nil {
		return deliver.Result{}, err
	}
	return captaincli.RunSessionDeliver(ctx, captaincli.SessionDeliverOptions{ID: sessionID, Text: text})
}

func (Captain) Launch(ctx context.Context, invocation captaincli.PromptInvocation) (captaincli.PromptRunResult, error) {
	if err := OpenCaptain(ctx); err != nil {
		return captaincli.PromptRunResult{}, err
	}
	return captaincli.LaunchPrompt(ctx, invocation)
}

func (Captain) Render(ctx context.Context, invocation captaincli.PromptInvocation) (captaincli.PromptRenderResult, error) {
	if err := OpenCaptain(ctx); err != nil {
		return captaincli.PromptRenderResult{}, err
	}
	return captaincli.RenderPrompt(ctx, invocation)
}
