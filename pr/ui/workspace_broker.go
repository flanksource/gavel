package ui

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

type databaseChangeBroker struct {
	listeners map[uint64]chan workspaceChange
	failures  map[string]error
	next      uint64
	cancel    context.CancelFunc
	workers   sync.WaitGroup
}

var workspaceBrokers = struct {
	sync.Mutex
	pools map[*sql.DB]*databaseChangeBroker
}{pools: map[*sql.DB]*databaseChangeBroker{}}

func subscribeDatabaseChanges(ctx context.Context, pool *sql.DB) (<-chan workspaceChange, func(), error) {
	workspaceBrokers.Lock()
	defer workspaceBrokers.Unlock()
	broker := workspaceBrokers.pools[pool]
	if broker == nil {
		root, cancel := context.WithCancel(context.Background())
		broker = &databaseChangeBroker{listeners: map[uint64]chan workspaceChange{}, failures: map[string]error{}, cancel: cancel}
		listeners := map[string]*captaindb.Listener{}
		for _, channel := range []string{"gavel_workspace_change", "captain_session_change", "captain_session_overview_change", "captain_prompt_run_overview_change"} {
			listener, err := captaindb.Listen(root, pool, channel)
			if err != nil {
				cancel()
				for _, started := range listeners {
					_ = started.Run(root, func(string) error { return nil }, func() error { return nil })
				}
				return nil, nil, err
			}
			listeners[channel] = listener
		}
		workspaceBrokers.pools[pool] = broker
		for channel, listener := range listeners {
			broker.workers.Add(1)
			go func() {
				defer broker.workers.Done()
				broker.run(root, pool, channel, listener)
			}()
		}
	}
	broker.next++
	id := broker.next
	events := make(chan workspaceChange, 1)
	broker.listeners[id] = events
	if len(broker.failures) > 0 {
		failures := make([]error, 0, len(broker.failures))
		for _, err := range broker.failures {
			failures = append(failures, err)
		}
		events <- workspaceChange{Err: errors.Join(failures...)}
	}
	release := func() {
		workspaceBrokers.Lock()
		delete(broker.listeners, id)
		last := len(broker.listeners) == 0
		if last {
			broker.cancel()
			delete(workspaceBrokers.pools, pool)
		}
		workspaceBrokers.Unlock()
		if last {
			broker.workers.Wait()
		}
	}
	return events, sync.OnceFunc(release), nil
}

func (b *databaseChangeBroker) run(ctx context.Context, pool *sql.DB, channel string, listener *captaindb.Listener) {
	for {
		err := listener.Run(ctx, func(payload string) error {
			change := workspaceChange{}
			if channel == "gavel_workspace_change" {
				if err := json.Unmarshal([]byte(payload), &change); err != nil {
					return fmt.Errorf("decode workspace notification: %w", err)
				}
				change.All = change.WorkspaceID == uuid.Nil
			} else {
				id, err := uuid.Parse(payload)
				if err != nil {
					return fmt.Errorf("decode Captain notification: %w", err)
				}
				if channel == "captain_prompt_run_overview_change" {
					change.RunIDs = []uuid.UUID{id}
				} else {
					change.SessionIDs = []uuid.UUID{id}
				}
			}
			b.broadcast(channel, change)
			return nil
		}, func() error { b.broadcast(channel, workspaceChange{Resync: true}); return nil })
		if ctx.Err() != nil {
			return
		}
		if err == nil {
			err = errors.New("workspace listener exited unexpectedly")
		}
		b.broadcast(channel, workspaceChange{Err: err})
		for {
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			replacement, listenErr := captaindb.Listen(ctx, pool, channel)
			if listenErr != nil {
				b.broadcast(channel, workspaceChange{Err: listenErr})
				continue
			}
			listener = replacement
			b.broadcast(channel, workspaceChange{Resync: true})
			break
		}
	}
}

func (b *databaseChangeBroker) broadcast(channel string, change workspaceChange) {
	workspaceBrokers.Lock()
	defer workspaceBrokers.Unlock()
	if change.Err != nil {
		b.failures[channel] = change.Err
	} else if change.Resync {
		delete(b.failures, channel)
	}
	failures := []string{}
	for name, err := range b.failures {
		failures = append(failures, name+": "+err.Error())
	}
	sort.Strings(failures)
	if len(failures) > 0 {
		change.Err = errors.New(strings.Join(failures, "; "))
	}
	for _, events := range b.listeners {
		next := change
		select {
		case previous := <-events:
			next = mergeWorkspaceChanges(previous, next)
		default:
		}
		events <- next
	}
}

func (s *Server) startWorkspaceChanges(ctx context.Context) error {
	d := s.workspaceData
	pool, err := d.db.DB()
	if err != nil {
		return err
	}
	events, release, err := subscribeDatabaseChanges(ctx, pool)
	if err != nil {
		return err
	}
	d.workers.Add(1)
	go func() {
		defer d.workers.Done()
		defer release()
		select {
		case <-ctx.Done():
			return
		case <-d.ready:
		}
		for {
			select {
			case <-ctx.Done():
				return
			case change := <-events:
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
				select {
				case more := <-events:
					change = mergeWorkspaceChanges(change, more)
				default:
				}
				d.mu.Lock()
				d.listenerError = ""
				if change.Err != nil {
					d.listenerError = change.Err.Error()
				}
				if change.Resync {
					d.snapshot.Epoch = uuid.NewString()
				}
				d.mu.Unlock()
				if change.Err != nil {
					d.changed.Notify()
					continue
				}
				var err error
				if change.Catalog || change.Resync || d.snapshotCopy().Error != "" {
					err = s.reloadWorkspaceData(ctx)
				}
				var ids []uuid.UUID
				if err == nil {
					ids, err = d.affectedWorkspaces(ctx, change)
				}
				if err == nil && ids != nil && len(ids) == 0 {
					continue
				}
				if err == nil {
					err = s.reconcileWorkspaceData(ctx, ids...)
				}
				if err == nil {
					err = s.prepareWorkspaceDiffs(ctx, ids...)
				}
				if err == nil {
					err = d.refresh(ctx, ids...)
				}
				if err != nil {
					d.fail(err)
				}
				d.changed.Notify()
			}
		}
	}()
	return nil
}
