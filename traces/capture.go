package traces

import (
	"context"
	"fmt"
	"maps"

	"github.com/flanksource/commons-db/query"
	"github.com/flanksource/commons-db/tracing"
)

// Capture is one running capture: a managed session committing to its own
// stream of the host's store.
type Capture struct {
	host    *Host
	kind    string
	managed *query.ManagedSession
}

// Kind is the trace kind the capture records.
func (c *Capture) Kind() string { return c.kind }

// Stream is the stream the capture commits to, which a host-fed kind such as
// terminal is written through.
func (c *Capture) Stream() string {
	events := c.managed.Session().Snapshot().Events
	if events == nil {
		panic(fmt.Sprintf("%s capture %s records no stream", c.kind, c.managed.Session().ID()))
	}
	return events.Stream
}

// Exports are the environment variables the processes the capture observes
// must inherit, such as HTTPS_PROXY for a recording proxy.
func (c *Capture) Exports() map[string]string {
	return maps.Clone(c.managed.Status().Exports)
}

// Mark commits everything the capture emitted so far and returns the seq its
// stream has committed through, so two marks bound a window exactly.
func (c *Capture) Mark(ctx context.Context) (int64, error) {
	seq, err := c.host.runtime.Mark(ctx, c.managed)
	if err != nil {
		return 0, fmt.Errorf("mark %s trace: %w", c.kind, err)
	}
	return seq, nil
}

// Ref names the capture's rows after seq after, through seq to.
func (c *Capture) Ref(after, to int64) tracing.TraceRef {
	return tracing.TraceRef{Kind: c.kind, Stream: c.Stream(), AfterSeq: after, ToSeq: to}
}

// Ended waits for a capture that ends by itself, such as a terminal whose
// writer closed, to finish committing what it recorded.
func (c *Capture) Ended(ctx context.Context) error {
	select {
	case <-c.managed.Session().Done():
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%s trace did not end: %w", c.kind, ctx.Err())
	}
}

// Stop ends the capture, committing and sealing its stream, and returns the
// seq the stream ended at. A capture that already ended by itself, such as a
// terminal whose writer closed, stops as it ended.
func (c *Capture) Stop(ctx context.Context) (int64, error) {
	stream := c.Stream()
	if _, err := c.managed.Stop(ctx); err != nil {
		return 0, fmt.Errorf("stop %s trace: %w", c.kind, err)
	}
	seq, err := c.host.highSeq(ctx, stream)
	if err != nil {
		return 0, fmt.Errorf("stop %s trace: %w", c.kind, err)
	}
	return seq, nil
}
