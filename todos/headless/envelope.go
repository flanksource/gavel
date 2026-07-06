package headless

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	todopkg "github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
)

// envelope is the mode-erased parse target: run mode fills the base fields,
// plan mode also carries the plan definition.
type envelope struct {
	types.ResultEnvelope
	Plan *types.PlanResult
}

// captureEnvelope obtains the structured result envelope for a finished (or
// timed-out) run: first from the final iteration's assistant text, then — when
// that has none — by resuming the session with a final-result request on a
// fresh context (the run context is already dead after a timeout). A plan run
// without an envelope is a hard error (its plan status/path are unknowable);
// run mode degrades to the transport result with a logged warning, leaving
// EndStatus empty so the degradation is visible to callers.
func (e *Executor) captureEnvelope(ctx *todopkg.ExecutorContext, provider captainai.StreamingProvider, result *todopkg.ExecutionResult, rres *agent.RunResult, todosInGroup []*types.TODO, timedOut bool, schemaJSON json.RawMessage) error {
	if env := e.envelopeFromRun(rres); env != nil {
		applyEnvelope(result, env)
		return nil
	}

	sessionID := runSessionID(rres, todosInGroup)
	if sessionID != "" {
		workDir := groupWorkDir(e.config.WorkDir, todosInGroup)
		template, err := todoprompt.ResolveFinalTemplate(workDir)
		if err != nil {
			return err
		}
		if env := e.requestFinalResult(ctx, provider, sessionID, workDir, timedOut, schemaJSON, template); env != nil {
			applyEnvelope(result, env)
			return nil
		}
	}

	if e.config.Mode == types.ModePlan {
		return fmt.Errorf("plan run produced no result envelope (session %q): plan status and path are unknown", sessionID)
	}
	ctx.Logger.Warnf("%s: no result envelope captured; falling back to the transport result", e.Name())
	return nil
}

// envelopeFromRun parses the envelope out of the final iteration's assistant
// text: the last text event alone first (the final reply is usually the bare
// JSON), then the iteration's full text.
func (e *Executor) envelopeFromRun(rres *agent.RunResult) *envelope {
	if rres == nil || rres.Loop == nil || len(rres.Loop.Iterations) == 0 {
		return nil
	}
	last := rres.Loop.Iterations[len(rres.Loop.Iterations)-1]
	var texts []string
	for _, ev := range last.Events {
		if ev.Kind == captainai.EventText && ev.Text != "" {
			texts = append(texts, ev.Text)
		}
	}
	return e.parseEnvelopeTexts(texts)
}

// requestFinalResult resumes the session with the mode's final-result turn and
// parses the reply. Failures are logged, not fatal — the caller decides what a
// missing envelope means for the mode.
func (e *Executor) requestFinalResult(ctx *todopkg.ExecutorContext, provider captainai.StreamingProvider, sessionID, workDir string, timedOut bool, schemaJSON json.RawMessage, template string) *envelope {
	freq, err := todoprompt.FinalResultRequest(template, sessionID, timedOut, schemaJSON)
	if err != nil {
		ctx.Logger.Warnf("%s: final-result request: %v", e.Name(), err)
		return nil
	}
	freq.Budget = api.Budget{MaxTurns: finalResultMaxTurns}
	freq.SetCwd(workDir)

	fctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), finalResultTimeout)
	defer cancel()

	ctx.Logger.Infof("%s: requesting final result from session %s (timedOut=%t)", e.Name(), sessionID, timedOut)
	events, err := provider.ExecuteStream(fctx, freq)
	if err != nil {
		ctx.Logger.Warnf("%s: final-result resume failed: %v", e.Name(), err)
		return nil
	}
	var texts []string
	for ev := range events {
		if ev.Kind == captainai.EventText && ev.Text != "" {
			texts = append(texts, ev.Text)
		}
	}
	return e.parseEnvelopeTexts(texts)
}

// parseEnvelopeTexts tries the last text alone, then the joined whole.
func (e *Executor) parseEnvelopeTexts(texts []string) *envelope {
	if len(texts) == 0 {
		return nil
	}
	if env := e.parseEnvelope(texts[len(texts)-1]); env != nil {
		return env
	}
	if len(texts) == 1 {
		return nil
	}
	return e.parseEnvelope(strings.Join(texts, "\n"))
}

func (e *Executor) parseEnvelope(text string) *envelope {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if e.config.Mode == types.ModePlan {
		pe, err := captainai.ParseStructured(text, (*types.PlanEnvelope).Validate)
		if err != nil {
			return nil
		}
		return &envelope{ResultEnvelope: pe.ResultEnvelope, Plan: &pe.Plan}
	}
	re, err := captainai.ParseStructured(text, (*types.ResultEnvelope).Validate)
	if err != nil {
		return nil
	}
	return &envelope{ResultEnvelope: *re}
}

func applyEnvelope(result *todopkg.ExecutionResult, env *envelope) {
	result.Summary = env.Summary
	result.EndStatus = env.EndStatus
	result.Questions = env.Questions
	result.Plan = env.Plan
}

// runSessionID prefers the session the Runner observed on the event stream,
// falling back to the todos' recorded prior session.
func runSessionID(rres *agent.RunResult, todosInGroup []*types.TODO) string {
	if rres != nil && rres.SessionID != "" {
		return rres.SessionID
	}
	return priorSessionID(todosInGroup)
}
