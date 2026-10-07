// Package context is gavel's process context: a commons context (logger,
// tracer, debug/trace) that also carries the process-wide services. Import it
// as gavelctx. The root is created once at the CLI entry point; every Context
// derived from it shares the same services.
package context

import (
	gocontext "context"
	"time"

	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/git/gitstate"
)

type Context struct {
	commonscontext.Context
	services *services
}

type services struct {
	gitState *gitstate.Service
}

type Option func(*services)

// WithGitState replaces the default git state service, e.g. with one built on
// a fake Compute.
func WithGitState(s *gitstate.Service) Option {
	return func(svc *services) { svc.gitState = s }
}

// New creates a root Context over base. Services it builds run their
// background work on base, so they stop when base is canceled.
func New(base gocontext.Context, opts ...Option) Context {
	ctx := Context{Context: commonscontext.NewContext(base), services: &services{}}
	for _, opt := range opts {
		opt(ctx.services)
	}
	if ctx.services.gitState == nil {
		ctx.services.gitState = gitstate.New(gitstate.Options{Context: base})
	}
	return ctx
}

// IsZero reports whether c was never created by New.
func (c Context) IsZero() bool {
	return c.services == nil
}

// Wrap rebinds cancellation and values to ctx (e.g. a request's context),
// keeping c's logger, tracer and services.
func (c Context) Wrap(ctx gocontext.Context) Context {
	inner := c.Context
	inner.Context = ctx
	return Context{Context: inner, services: c.services}
}

func (c Context) WithValue(key, val any) Context {
	return Context{Context: c.Context.WithValue(key, val), services: c.services}
}

func (c Context) WithTimeout(timeout time.Duration) (Context, gocontext.CancelFunc) {
	inner, cancel := c.Context.WithTimeout(timeout)
	return Context{Context: inner, services: c.services}, cancel
}

// GitState is the process-wide memoized git state service.
func (c Context) GitState() *gitstate.Service {
	if c.services == nil {
		panic("gavel context has no services: create it with context.New")
	}
	return c.services.gitState
}
