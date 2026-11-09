package ui

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons/logger"
	"github.com/google/uuid"
)

type workspaceChange struct {
	Catalog bool `json:"catalog"`
}

func (d *workspaceData) refresh(ctx context.Context) error {
	var rows []struct {
		WorkspaceID uuid.UUID
		Fingerprint string
	}
	err := d.db.WithContext(ctx).Raw(`SELECT i.workspace_id,
 md5(string_agg(concat_ws('|', i.id::text, i.version::text, i.updated_at::text,
 v.execution_state, v.approval_state, v.plan_revision::text,
 (v.session_summary - 'updatedAt' - 'durationMs')::text), ',' ORDER BY i.id)) AS fingerprint
 FROM todo_issues i JOIN todo_issue_overview v ON v.issue_id=i.id
 GROUP BY i.workspace_id`).Scan(&rows).Error
	if err != nil {
		return fmt.Errorf("read workspace generations: %w", err)
	}
	fingerprints := map[uuid.UUID]string{}
	for _, row := range rows {
		fingerprints[row.WorkspaceID] = row.Fingerprint
	}
	d.mu.Lock()
	changed := false
	next := map[string]workspaceRevision{}
	for dir, entry := range d.entries {
		if entry.Workspace == nil {
			continue
		}
		id := entry.Workspace.ID
		fingerprint := fingerprints[id]
		previous, exists := d.snapshot.Workspaces[dir]
		if !exists || previous.Fingerprint != fingerprint {
			previous.Revision++
			changed = true
		}
		previous.ID = id
		previous.Dir = dir
		previous.Fingerprint = fingerprint
		next[dir] = previous
	}
	if len(next) != len(d.snapshot.Workspaces) {
		changed = true
	}
	d.snapshot.Workspaces = next
	d.mu.Unlock()
	if changed {
		d.changed.Notify()
	}
	return nil
}

type databaseChangeBroker struct {
	listeners map[uint64]chan bool
	next      uint64
	cancel    context.CancelFunc
}

var workspaceBrokers = struct {
	sync.Mutex
	pools map[*sql.DB]*databaseChangeBroker
}{pools: map[*sql.DB]*databaseChangeBroker{}}

func subscribeDatabaseChanges(ctx context.Context, pool *sql.DB) (<-chan bool, error) {
	workspaceBrokers.Lock()
	defer workspaceBrokers.Unlock()
	broker := workspaceBrokers.pools[pool]
	if broker == nil {
		root, cancel := context.WithCancel(context.Background())
		broker = &databaseChangeBroker{listeners: map[uint64]chan bool{}, cancel: cancel}
		var listeners []*captaindb.Listener
		for _, channel := range []string{"gavel_workspace_change", "captain_session_change", "captain_session_overview_change", "captain_prompt_run_overview_change"} {
			listener, err := captaindb.Listen(root, pool, channel)
			if err != nil {
				cancel()
				for _, started := range listeners {
					go started.Run(root, func(string) error { return nil }, func() error { return nil })
				}
				return nil, err
			}
			listeners = append(listeners, listener)
		}
		workspaceBrokers.pools[pool] = broker
		for _, listener := range listeners {
			go func(listener *captaindb.Listener) {
				err := listener.Run(root, func(payload string) error {
					var change workspaceChange
					_ = json.Unmarshal([]byte(payload), &change)
					broker.broadcast(change.Catalog)
					return nil
				}, func() error { broker.broadcast(true); return nil })
				if err != nil && root.Err() == nil {
					logger.Errorf("workspace change listener stopped: %v", err)
					broker.broadcast(true)
				}
			}(listener)
		}
	}
	broker.next++
	id := broker.next
	events := make(chan bool, 1)
	broker.listeners[id] = events
	go func() {
		<-ctx.Done()
		workspaceBrokers.Lock()
		defer workspaceBrokers.Unlock()
		delete(broker.listeners, id)
		if len(broker.listeners) == 0 {
			broker.cancel()
			delete(workspaceBrokers.pools, pool)
		}
	}()
	return events, nil
}

func (b *databaseChangeBroker) broadcast(catalog bool) {
	workspaceBrokers.Lock()
	defer workspaceBrokers.Unlock()
	for _, events := range b.listeners {
		select {
		case previous := <-events:
			catalog = catalog || previous
		default:
		}
		events <- catalog
	}
}

func (s *Server) startWorkspaceChanges(ctx context.Context) error {
	pool, err := s.workspaceData.db.DB()
	if err != nil {
		return err
	}
	events, err := subscribeDatabaseChanges(ctx, pool)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case catalog := <-events:
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
				select {
				case more := <-events:
					catalog = catalog || more
				default:
				}
				var err error
				if catalog {
					err = s.reloadWorkspaceData(ctx)
				} else {
					err = s.workspaceData.refresh(ctx)
				}
				if err != nil {
					s.workspaceData.fail(err)
				}
			}
		}
	}()
	return nil
}
