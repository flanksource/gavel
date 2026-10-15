// Push checks: after a push to hosted gavel, each updated branch runs a push-check job in the
// sandbox, whose log and verdict stream back to the pushing SSH session.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/serve"
	"github.com/flanksource/gavel/tenant"
)

// pushChecks returns the push server's OnPushed hook. A nil dispatcher means no sandbox is
// configured, so a push is only stored.
func pushChecks(dispatcher *sandbox.Dispatcher) func(context.Context, serve.Push) {
	return func(ctx context.Context, push serve.Push) {
		if dispatcher == nil {
			fmt.Fprintln(push.Out, "push received; push checks need a sandbox: start gavel serve with --sandbox")
			return
		}
		for _, ref := range push.Refs {
			branch, ok := strings.CutPrefix(ref.Name, "refs/heads/")
			if !ok || ref.New == "" {
				continue
			}
			fmt.Fprintf(push.Out, "%s: running push check\n", branch)
			user := push.UserID
			run, err := dispatcher.Start(ctx, tenant.Scope{OrgID: push.OrgID, UserID: &user},
				sandbox.RunSpec{Kind: "push-check", RepoID: push.RepoID, Ref: ref.Name, SHA: ref.New}, push.Out)
			if err != nil {
				fmt.Fprintf(push.Out, "%s: push check could not start: %v\n", branch, err)
				continue
			}
			result, err := run.Wait(ctx)
			switch {
			case err != nil:
				fmt.Fprintf(push.Out, "%s: push check did not finish: %v\n", branch, err)
			case result.ExitCode == 0:
				fmt.Fprintf(push.Out, "%s: push check PASSED\n", branch)
			default:
				fmt.Fprintf(push.Out, "%s: push check FAILED (exit code %d)\n", branch, result.ExitCode)
			}
		}
	}
}
