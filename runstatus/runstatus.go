// Package runstatus publishes what a running gavel process is doing — the
// test, lint and fixture progress its live UI shows, and the slots it holds
// or waits for — as the sdk/status contract: a status file rewritten on
// change and, once the run has lasted a while, a socket streaming it.
package runstatus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/flanksource/clicky/sse"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sdk/status"
	"github.com/flanksource/gavel/slots"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type Options struct {
	Cache   string
	WorkDir string
	Command string
	// Interval is how often progress is read. Defaults to 500ms.
	Interval time.Duration
	// SocketAfter is how long the run lasts before its socket opens.
	// Defaults to 30s.
	SocketAfter time.Duration
	// Tasks reads the live task groups; defaults to clicky's global ones.
	Tasks func() []clickytask.TaskSnapshot
}

// Reporter keeps this process's status file, and socket, current.
type Reporter struct {
	options Options
	path    string
	cancel  context.CancelFunc
	done    chan struct{}
	changed sse.Notifier

	mu      sync.Mutex
	current status.Status
	// dirty publishes the next refresh even if progress did not change.
	dirty  bool
	tests  *status.Summary
	err    error
	server *socketServer
}

// Start removes the status of processes that have exited and begins
// publishing this one's.
func Start(ctx context.Context, options Options) (*Reporter, error) {
	if options.Cache == "" {
		return nil, errors.New("runstatus: cache directory is required")
	}
	options.Interval = cmpDuration(options.Interval, 500*time.Millisecond)
	options.SocketAfter = cmpDuration(options.SocketAfter, 30*time.Second)
	if options.Tasks == nil {
		options.Tasks = liveTasks
	}
	if err := os.MkdirAll(status.Dir(options.Cache), 0o700); err != nil {
		return nil, fmt.Errorf("runstatus: %w", err)
	}
	if err := removeStale(options.Cache); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	r := &Reporter{
		options: options, path: status.Path(options.Cache, os.Getpid()), cancel: cancel, done: make(chan struct{}),
		current: status.Status{PID: os.Getpid(), Command: options.Command, WorkDir: options.WorkDir, Started: time.Now().UTC()},
		dirty:   true,
	}
	if err := r.refresh(); err != nil {
		cancel()
		return nil, err
	}
	go r.loop(ctx)
	return r, nil
}

func cmpDuration(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func removeStale(cache string) error {
	stale, err := status.Stale(cache)
	if err != nil {
		return fmt.Errorf("runstatus: %w", err)
	}
	for _, path := range stale {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("runstatus: remove stale %s: %w", path, err)
		}
	}
	return nil
}

// SetTests records the latest test tree a run reported.
func (r *Reporter) SetTests(tests []parsers.Test) {
	if r == nil {
		return
	}
	summary := summaryOf(parsers.Tests(tests).Sum())
	r.mu.Lock()
	r.tests = &summary
	r.mu.Unlock()
}

// Status is what the reporter last published.
func (r *Reporter) Status() status.Status {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.current
}

func (r *Reporter) loop(ctx context.Context) {
	defer close(r.done)
	ticker := time.NewTicker(r.options.Interval)
	defer ticker.Stop()
	socket := time.NewTimer(r.options.SocketAfter)
	defer socket.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-socket.C:
			r.fail(r.serve())
		case <-ticker.C:
		}
		r.fail(r.refresh())
	}
}

// fail keeps the first error for Stop and logs it: a reporter that cannot
// publish must not stop the run it reports on.
func (r *Reporter) fail(err error) {
	if err == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err == nil {
		logger.Warnf("gavel status: %v", err)
		r.err = err
	}
}

// refresh reads progress and publishes it when it changed.
func (r *Reporter) refresh() error {
	groups, running := progress(r.options.Tasks())
	leases := slots.Leases()
	r.mu.Lock()
	next := r.current
	next.Tests, next.Groups, next.Running, next.Slots = r.tests, groups, running, leases
	if !r.dirty && reflect.DeepEqual(next, r.current) {
		r.mu.Unlock()
		return nil
	}
	next.Updated = time.Now().UTC()
	r.current, r.dirty = next, false
	r.mu.Unlock()
	r.changed.Notify()
	return writeAtomic(r.path, next)
}

func writeAtomic(path string, value status.Status) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("runstatus: encode: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("runstatus: %w", err)
	}
	_, err = tmp.Write(data)
	if err = errors.Join(err, tmp.Close()); err == nil {
		err = os.Rename(tmp.Name(), path)
	}
	if err != nil {
		return errors.Join(fmt.Errorf("runstatus: write %s: %w", path, err), os.Remove(tmp.Name()))
	}
	return nil
}

// Stop publishes nothing more, removes the status file and socket, and
// returns the first error publishing hit.
func (r *Reporter) Stop() error {
	if r == nil {
		return nil
	}
	r.cancel()
	<-r.done
	r.mu.Lock()
	err, server := r.err, r.server
	r.mu.Unlock()
	// Its handlers read the status under the lock.
	if server != nil {
		err = errors.Join(err, server.close())
	}
	if removeErr := os.Remove(r.path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		err = errors.Join(err, removeErr)
	}
	return err
}
