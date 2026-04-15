//go:build !unix

package main

import (
	"fmt"
	"net"
	"strconv"
)

// openListener always takes the standalone path on non-Unix platforms: no
// socket inheritance, no lockfile. The fork handoff in `gavel test --ui
// --auto-stop` is Unix-only; non-Unix builds fall back to the foreground
// SIGINT-block behavior and never spawn a child, so the inherited-FD branch
// is unreachable from within the same binary.
func openListener(opts UIServeOptions) (net.Listener, error) {
	if opts.ListenerFD > 0 {
		return nil, fmt.Errorf("--listener-fd is not supported on this platform")
	}
	host := opts.Addr
	if host == "" {
		host = "localhost"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(opts.Port))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("bind %s: %w", addr, err)
	}
	return l, nil
}

// notifyHandoff is a no-op on non-Unix: the fork parent never spawns a child
// on these platforms, so there's no lockfile to signal.
func notifyHandoff(_ UIServeOptions, _ int, _ string) error { return nil }
