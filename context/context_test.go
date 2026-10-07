package context_test

import (
	gocontext "context"
	"testing"
	"time"

	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/git/gitstate"
)

type requestKey struct{}

func TestDerivedContextsShareGitState(t *testing.T) {
	root := gavelctx.New(gocontext.Background())
	request, cancel := gocontext.WithCancel(gocontext.WithValue(gocontext.Background(), requestKey{}, "r1"))
	defer cancel()

	wrapped := root.Wrap(request)
	valued := root.WithValue("k", "v")
	timed, cancelTimed := root.WithTimeout(time.Minute)
	defer cancelTimed()

	for name, derived := range map[string]gavelctx.Context{"Wrap": wrapped, "WithValue": valued, "WithTimeout": timed} {
		if derived.GitState() != root.GitState() {
			t.Errorf("%s: GitState() is a different instance from the root's", name)
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

func TestWithGitStateOverridesTheService(t *testing.T) {
	fake := gitstate.New(gitstate.Options{Compute: func(gocontext.Context, string) (gitstate.State, error) {
		return gitstate.State{Base: "fake"}, nil
	}})

	ctx := gavelctx.New(gocontext.Background(), gavelctx.WithGitState(fake))

	if ctx.GitState() != fake {
		t.Fatal("GitState() is not the injected service")
	}
	if ctx.Wrap(gocontext.Background()).GitState() != fake {
		t.Fatal("a wrapped context lost the injected service")
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
