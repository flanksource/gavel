package context_test

import (
	gocontext "context"
	"errors"
	"testing"
	"time"

	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/git/gitstate"
)

type requestKey struct{}

func TestDerivedContextsShareGitTracker(t *testing.T) {
	tracker := gitstate.NewTracker(gitstate.Options{})
	root := gavelctx.New(gocontext.Background(), gavelctx.WithGitTracker(tracker))
	request, cancel := gocontext.WithCancel(gocontext.WithValue(gocontext.Background(), requestKey{}, "r1"))
	defer cancel()

	wrapped := root.Wrap(request)
	valued := root.WithValue("k", "v")
	timed, cancelTimed := root.WithTimeout(time.Minute)
	defer cancelTimed()

	for name, derived := range map[string]gavelctx.Context{"Wrap": wrapped, "WithValue": valued, "WithTimeout": timed} {
		if got, err := derived.GitTracker(); err != nil || got != tracker {
			t.Errorf("%s: GitTracker() = %p, %v; want the root's %p", name, got, err, tracker)
		}
	}
	if wrapped.Value(requestKey{}) != "r1" {
		t.Errorf("Wrap: request value = %v, want r1", wrapped.Value(requestKey{}))
	}
	cancel()
	if wrapped.Err() == nil {
		t.Error("Wrap: canceling the request did not cancel the wrapped context")
	}
	if root.Err() != nil {
		t.Errorf("Wrap: canceling the request canceled the root: %v", root.Err())
	}
}

func TestGitTrackerMissingIsAnError(t *testing.T) {
	for name, ctx := range map[string]gavelctx.Context{"zero": {}, "New without tracker": gavelctx.New(gocontext.Background())} {
		if _, err := ctx.GitTracker(); !errors.Is(err, gavelctx.ErrNoGitTracker) {
			t.Errorf("%s: GitTracker() error = %v, want ErrNoGitTracker", name, err)
		}
	}
}

func TestZeroContextIsZero(t *testing.T) {
	if !(gavelctx.Context{}).IsZero() {
		t.Error("zero Context: IsZero() = false")
	}
	if gavelctx.New(gocontext.Background()).IsZero() {
		t.Error("New Context: IsZero() = true")
	}
}
