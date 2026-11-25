package fixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/flanksource/clicky/prompt"
)

// Lookup relay across the worker protocol: a form step in an isolated worker
// registers its lookup fields in the worker, but the browser and terminal
// that answer the form ask the parent. The parent forwards every lookup URL a
// relayed prompt's schema names (WorkerLookups.Forward) for as long as the
// prompt is open; a lookup of one is written to the worker's answer
// descriptor as a `lookup` message, which the worker resolves
// (WorkerLookups.Resolve) and answers with a `lookup_result` event. A parent
// that gives up on a lookup sends `lookup_cancel`, which ends the worker's.

const (
	workerEventLookupResult = "lookup_result"
	workerErrorLookupAbsent = "lookup_not_found"
)

// RemoteLookup answers one lookup request, both in the lookup package's wire
// shape.
type RemoteLookup func(ctx context.Context, request json.RawMessage) (json.RawMessage, error)

// LookupRelay is how an isolated run answers its worker's lookup fields; the
// lookup package installs it, as it owns the fields and their wire shape.
type LookupRelay struct {
	// Forward holds, until release, every lookup field schema names, each
	// answered by resolve.
	Forward func(schema json.RawMessage, resolve RemoteLookup) (release func(), err error)
	// Resolve answers a forwarded request inside the worker.
	Resolve RemoteLookup
	// NotFound is the error a lookup of no registered field wraps, which the
	// relay carries across the process boundary.
	NotFound error
}

// WorkerLookups is the run's lookup relay; without it an isolated run's forms
// cannot have lookup fields, since nothing registers them.
var WorkerLookups LookupRelay

// workerLookup is a lookup message on the answer descriptor.
type workerLookup struct {
	Request json.RawMessage `json:"request,omitempty"`
	Cancel  bool            `json:"cancel,omitempty"`
}

// codedError is an error read from the worker protocol that keeps the
// identity of the sentinel its code names.
type codedError struct {
	message string
	cause   error
}

func (e codedError) Error() string { return e.message }
func (e codedError) Unwrap() error { return e.cause }

// lookup resolves the parent's lookup message id inside the worker, sending
// its lookup_result; a cancel message ends the lookup of the same id.
func (w *workerPrompter) lookup(id string, message workerLookup) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if message.Cancel {
		if cancel, ok := w.lookups[id]; ok {
			cancel()
		}
		return nil
	}
	if _, ok := w.lookups[id]; ok {
		return fmt.Errorf("fixture parent sent lookup %s twice", id)
	}
	ctx, cancel := context.WithCancel(w.ctx)
	w.lookups[id] = cancel
	w.lookupWG.Add(1)
	go func() {
		defer w.lookupWG.Done()
		defer cancel()
		event := workerEvent{Version: 1, Kind: workerEventLookupResult, LookupID: id}
		if WorkerLookups.Resolve == nil {
			event.Error = "fixture worker has no lookup relay to resolve lookup " + id
		} else if result, err := WorkerLookups.Resolve(ctx, message.Request); err != nil {
			event.Error = err.Error()
			if WorkerLookups.NotFound != nil && errors.Is(err, WorkerLookups.NotFound) {
				event.ErrorCode = workerErrorLookupAbsent
			}
		} else {
			event.LookupResult = result
		}
		w.mu.Lock()
		delete(w.lookups, id)
		w.mu.Unlock()
		if err := w.send(event); err != nil {
			w.close(fmt.Errorf("answer fixture parent lookup %s: %w", id, err))
		}
	}()
	return nil
}

// stopLookups ends every lookup still running and waits for their results to
// be sent, so none follows the worker's result event.
func (w *workerPrompter) stopLookups() {
	w.mu.Lock()
	for _, cancel := range w.lookups {
		cancel()
	}
	w.mu.Unlock()
	w.lookupWG.Wait()
}

// forward holds the lookup fields p names until the returned release.
func (r *promptRelay) forward(p prompt.Prompt) (func(), error) {
	if WorkerLookups.Forward == nil || len(p.Schema) == 0 {
		return func() {}, nil
	}
	release, err := WorkerLookups.Forward(p.Schema, r.lookup)
	if err != nil {
		return nil, fmt.Errorf("fixture worker prompt %q: %w", p.Title, err)
	}
	return release, nil
}

// lookup asks the worker to resolve request and waits for its lookup_result.
func (r *promptRelay) lookup(ctx context.Context, request json.RawMessage) (json.RawMessage, error) {
	r.mu.Lock()
	r.nextLookup++
	id := strconv.Itoa(r.nextLookup)
	reply := make(chan workerEvent, 1)
	r.lookups[id] = reply
	r.mu.Unlock()
	if err := r.send(workerAnswer{Version: 1, LookupID: id, Lookup: &workerLookup{Request: request}}); err != nil {
		r.abandonLookup(id)
		return nil, err
	}
	select {
	case event := <-reply:
		switch {
		case event.ErrorCode == workerErrorLookupAbsent && WorkerLookups.NotFound != nil:
			return nil, codedError{message: event.Error, cause: WorkerLookups.NotFound}
		case event.Error != "":
			return nil, errors.New(event.Error)
		}
		return event.LookupResult, nil
	case <-ctx.Done():
		r.abandonLookup(id)
		cancelErr := r.send(workerAnswer{Version: 1, LookupID: id, Lookup: &workerLookup{Cancel: true}})
		return nil, errors.Join(context.Cause(ctx), cancelErr)
	case <-r.ctx.Done():
		return nil, fmt.Errorf("lookup %s: the fixture worker has exited", id)
	}
}

func (r *promptRelay) abandonLookup(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.lookups[id]; ok {
		delete(r.lookups, id)
		r.abandonedLookups[id] = true
	}
}

// deliverLookup hands the worker's lookup_result to the lookup waiting on it.
func (r *promptRelay) deliverLookup(event workerEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if reply, ok := r.lookups[event.LookupID]; ok {
		delete(r.lookups, event.LookupID)
		reply <- event
		return nil
	}
	if r.abandonedLookups[event.LookupID] {
		delete(r.abandonedLookups, event.LookupID)
		return nil
	}
	return fmt.Errorf("fixture worker answered lookup %q, which the parent never asked", event.LookupID)
}
