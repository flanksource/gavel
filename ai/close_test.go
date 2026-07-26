package ai

import (
	"context"
	"errors"
	"testing"

	"github.com/flanksource/captain/pkg/api"
)

// plainProvider has nothing to release — CloseProvider must treat it as a no-op
// rather than failing.
type plainProvider struct{}

func (p *plainProvider) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, errors.New("not implemented")
}
func (p *plainProvider) GetModel() string        { return "test-model" }
func (p *plainProvider) GetBackend() api.Backend { return "test" }

// closableProvider stands in for the process-backed backends (claude-agent,
// cmux, codex-appserver) whose supervised child only stops on Close.
type closableProvider struct {
	plainProvider
	closed int
	err    error
}

func (p *closableProvider) Close() error {
	p.closed++
	return p.err
}

// wrappedProvider mirrors the middleware wrapper NewProvider returns: it
// implements Provider itself and exposes the inner provider via Unwrap, so
// Close is only reachable through the unwrap chain.
type wrappedProvider struct {
	inner api.Provider
}

func (w *wrappedProvider) Execute(ctx context.Context, spec api.Spec) (*api.Response, error) {
	return w.inner.Execute(ctx, spec)
}
func (w *wrappedProvider) GetModel() string        { return w.inner.GetModel() }
func (w *wrappedProvider) GetBackend() api.Backend { return w.inner.GetBackend() }
func (w *wrappedProvider) Unwrap() api.Provider    { return w.inner }

func TestCloseProvider_ReachesCloserThroughWrappers(t *testing.T) {
	inner := &closableProvider{}
	// Two layers, because gavel's NewProvider stacks logging on top of caching.
	if err := CloseProvider(&wrappedProvider{inner: &wrappedProvider{inner: inner}}); err != nil {
		t.Fatalf("CloseProvider returned %v, want nil", err)
	}
	if inner.closed != 1 {
		t.Errorf("inner provider closed %d times, want 1", inner.closed)
	}
}

func TestCloseProvider_NoCloserIsNoOp(t *testing.T) {
	if err := CloseProvider(&wrappedProvider{inner: &plainProvider{}}); err != nil {
		t.Errorf("CloseProvider on a provider without Close returned %v, want nil", err)
	}
}

// TestCloseProvider_ReachesRealMiddlewareStack is the regression guard for the
// hang: NewProvider hands back captain's middleware wrapper, so Close is only
// reachable by unwrapping. A wrapper that stops forwarding Unwrap would leave
// the claude-agent tsx child running and block process exit.
func TestCloseProvider_ReachesRealMiddlewareStack(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Model = api.Model{Name: "claude-sonnet-5", Backend: api.BackendClaudeAgent}

	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider(claude-agent) failed: %v", err)
	}
	if _, ok := api.ProviderAs[api.CloseableProvider](provider); !ok {
		t.Fatal("claude-agent provider is not closeable through the middleware stack; its supervised tsx child would outlive the command")
	}
	if err := CloseProvider(provider); err != nil {
		t.Errorf("CloseProvider returned %v, want nil", err)
	}
}

func TestCloseProvider_PropagatesCloseError(t *testing.T) {
	want := errors.New("shutdown rpc failed")
	inner := &closableProvider{err: want}
	if err := CloseProvider(&wrappedProvider{inner: inner}); !errors.Is(err, want) {
		t.Errorf("CloseProvider returned %v, want %v", err, want)
	}
}
