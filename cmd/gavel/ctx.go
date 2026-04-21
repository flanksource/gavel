package main

import (
	"context"

	commonsContext "github.com/flanksource/commons/context"
)

// mergeContexts derives a child context that is cancelled when either parent
// fires. parent may be nil, in which case primary is returned unchanged.
func mergeContexts(primary commonsContext.Context, parent context.Context) commonsContext.Context {
	if parent == nil {
		return primary
	}
	child, cancel := context.WithCancel(primary)
	go func() {
		select {
		case <-parent.Done():
			cancel()
		case <-child.Done():
		}
	}()
	return commonsContext.NewContext(child)
}
