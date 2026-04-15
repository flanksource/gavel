//go:build !unix

package main

import (
	"fmt"
	"net"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// handoffDetachedUI is not supported on non-Unix platforms. The inherited-FD
// + flock handoff relies on SysProcAttr.Setsid, cmd.ExtraFiles, and the
// golang.org/x/sys/unix flock wrappers, none of which have Windows
// equivalents we want to pull in right now. Return an error so the caller
// falls back to the SIGINT-block behavior.
func handoffDetachedUI(
	_ net.Listener,
	_ []parsers.Test,
	_ []*linters.LinterResult,
	_ time.Duration,
	_ time.Duration,
) error {
	return fmt.Errorf("detached UI handoff is not supported on this platform")
}
