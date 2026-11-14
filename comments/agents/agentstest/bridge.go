// Package agentstest is an in-memory agents.Bridge for tests and test hosts.
package agentstest

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/session/deliver"
)

// Delivery is one message delivered into a session.
type Delivery struct{ SessionID, Text string }

// Bridge serves Sessions as live, records deliveries and launches, and makes
// every launch report a new live session "managed-<n>" in Dir.
type Bridge struct {
	mu       sync.Mutex
	Sessions []captaincli.SessionRecord
	// Dir is the cwd of the sessions launches start.
	Dir       string
	Delivered []Delivery
	Launched  []captaincli.PromptInvocation
	// FailDeliver makes every delivery fail with this message.
	FailDeliver string
	hold        func(stage string)
}

// SetHold calls hold with "live" before every live listing returns and with
// "deliver" before every delivery is made, so a test can pause a delivery
// mid-flight; nil removes it.
func (b *Bridge) SetHold(hold func(stage string)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.hold = hold
}

func (b *Bridge) pause(stage string) {
	b.mu.Lock()
	hold := b.hold
	b.mu.Unlock()
	if hold != nil {
		hold(stage)
	}
}

func (b *Bridge) Live(context.Context) ([]captaincli.SessionRecord, error) {
	b.pause("live")
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]captaincli.SessionRecord(nil), b.Sessions...), nil
}

func (b *Bridge) Deliver(_ context.Context, sessionID, text string) (deliver.Result, error) {
	b.pause("deliver")
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.FailDeliver != "" {
		return deliver.Result{}, errors.New(b.FailDeliver)
	}
	b.Delivered = append(b.Delivered, Delivery{SessionID: sessionID, Text: text})
	return deliver.Result{Channel: deliver.ChannelClaudeSocket, Status: deliver.StatusSent}, nil
}

func (b *Bridge) Launch(_ context.Context, invocation captaincli.PromptInvocation) (captaincli.PromptRunResult, error) {
	b.mu.Lock()
	b.Launched = append(b.Launched, invocation)
	id := fmt.Sprintf("managed-%d", len(b.Launched))
	b.Sessions = append(b.Sessions, captaincli.SessionRecord{ID: id, Source: "claude", Slug: id, Live: &captaincli.SessionLiveWire{CWD: b.Dir}})
	b.mu.Unlock()
	if invocation.OnEvent != nil {
		invocation.OnEvent(0, api.Event{Kind: api.EventSystem, SessionID: id})
	}
	return captaincli.PromptRunResult{RunID: "run-" + id}, nil
}

// Render resolves the prompt with captain, without the process database the in-process bridge opens first.
func (b *Bridge) Render(ctx context.Context, invocation captaincli.PromptInvocation) (captaincli.PromptRenderResult, error) {
	return captaincli.RenderPrompt(ctx, invocation)
}

// Snapshot returns what was delivered and launched so far.
func (b *Bridge) Snapshot() ([]Delivery, []captaincli.PromptInvocation) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]Delivery(nil), b.Delivered...), append([]captaincli.PromptInvocation(nil), b.Launched...)
}
