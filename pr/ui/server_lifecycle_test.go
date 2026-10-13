// Tests that a Server's background loops stop when its context ends, so evicting
// an organization's Server leaves no goroutines behind.
package ui

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/require"
)

// serverLoops counts goroutines currently running one of the Server's background loops.
func serverLoops() int {
	buf := make([]byte, 1<<22)
	stacks := string(buf[:runtime.Stack(buf, true)])
	count := 0
	for _, stack := range strings.Split(stacks, "\n\n") {
		for _, loop := range []string{"(*Server).authProbeLoop", "(*Server).procMetricsLoop", "(*Server).taskHistoryImportLoop"} {
			if strings.Contains(stack, loop) {
				count++
			}
		}
	}
	return count
}

func TestServerLoopsStopWithItsContext(t *testing.T) {
	before := serverLoops()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	noGitHub := github.Options{TokenSource: func() (string, error) { return "", errors.New("no GitHub in this test") }}
	NewServer(gavelctx.New(ctx), time.Minute, noGitHub, SearchConfig{})
	require.Eventually(t, func() bool { return serverLoops() == before+3 }, 5*time.Second, 10*time.Millisecond)
	cancel()
	require.Eventually(t, func() bool { return serverLoops() == before }, 5*time.Second, 10*time.Millisecond,
		"loops still running after the context ended")
}
