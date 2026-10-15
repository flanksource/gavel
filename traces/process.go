package traces

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/flanksource/commons-db/tracing"
	"github.com/flanksource/commons-db/tracing/process"
)

// ProcessTrace is the run-scope capture of both process kinds over one root:
// the streams the Processes tab reads through the query mount.
type ProcessTrace struct {
	host    *Host
	tree    *Capture
	samples *Capture
}

// ProcessRefs name the process capture's streams in the run's store.
type ProcessRefs struct {
	Tree    tracing.TraceRef `json:"tree"`
	Samples tracing.TraceRef `json:"samples"`
}

// StartProcessTrace captures the process tree under rootPID, and its samples.
func (h *Host) StartProcessTrace(rootPID int) (*ProcessTrace, error) {
	params := json.RawMessage(`{"rootPid": ` + strconv.Itoa(rootPID) + `}`)
	tree, err := h.Start(process.TreeKind, params)
	if err != nil {
		return nil, err
	}
	samples, err := h.Start(process.SamplesKind, params)
	if err != nil {
		_, stopErr := tree.Stop(context.Background())
		return nil, errors.Join(err, stopErr)
	}
	return &ProcessTrace{host: h, tree: tree, samples: samples}, nil
}

// Refs are the whole tree and samples streams.
func (p *ProcessTrace) Refs() ProcessRefs {
	return ProcessRefs{Tree: p.tree.Ref(0, 0), Samples: p.samples.Ref(0, 0)}
}

// Session is the tree capture's session, which a client checkpoints through
// the sessions API to dump the run's goroutines onto its root.
func (p *ProcessTrace) Session() string { return p.tree.managed.Session().ID() }

// Checkpoint observes the tree and dumps the root's goroutines onto its node,
// both committed when it returns.
func (p *ProcessTrace) Checkpoint(ctx context.Context) error {
	return p.host.runtime.Checkpoint(ctx, p.tree.managed)
}

// Stop ends both captures; their streams stay in the run's store.
func (p *ProcessTrace) Stop(ctx context.Context) error {
	_, treeErr := p.tree.Stop(ctx)
	_, samplesErr := p.samples.Stop(ctx)
	return errors.Join(treeErr, samplesErr)
}
