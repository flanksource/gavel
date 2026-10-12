// Package context is gavel's process context: a commons context (logger,
// tracer, debug/trace) that also carries the process-wide services. Import it
// as gavelctx. The root is created once at the CLI entry point; every Context
// derived from it shares the same services.
package context

import (
	gocontext "context"
	"errors"
	"time"

	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/git/gitstate"
)

// ErrNoGitTracker is returned by GitTracker when the process runs without the
// database the git state tracker keeps its rows in.
var ErrNoGitTracker = errors.New("git state tracking requires PostgreSQL; set GAVEL_DB_DSN or configure embedded PostgreSQL with gavel system install --embedded")

type Context struct {
	commonscontext.Context
	services *services
}

type services struct {
	gitTracker *gitstate.Tracker
}

type Option func(*services)

// WithGitTracker sets the process's git state tracker.
func WithGitTracker(t *gitstate.Tracker) Option {
	return func(svc *services) { svc.gitTracker = t }
}

// New creates a root Context over base.
func New(base gocontext.Context, opts ...Option) Context {
	ctx := Context{Context: commonscontext.NewContext(base), services: &services{}}
	for _, opt := range opts {
		opt(ctx.services)
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

// GitTracker is the process-wide git state tracker, ErrNoGitTracker when the
// process has none.
func (c Context) GitTracker() (*gitstate.Tracker, error) {
	if c.services == nil || c.services.gitTracker == nil {
		return nil, ErrNoGitTracker
	}
	return c.services.gitTracker, nil
}
