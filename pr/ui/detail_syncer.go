package ui

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
)

type DetailSyncer struct {
	srv        *Server
	cache      *DetailCache
	ghOpts     github.Options
	maxWorkers int
	notify     chan struct{}
}

func NewDetailSyncer(srv *Server, cache *DetailCache, ghOpts github.Options) *DetailSyncer {
	return &DetailSyncer{
		srv:        srv,
		cache:      cache,
		ghOpts:     ghOpts,
		maxWorkers: 3,
		notify:     make(chan struct{}, 1),
	}
}

func (ds *DetailSyncer) Notify() {
	select {
	case ds.notify <- struct{}{}:
	default:
	}
}

func (ds *DetailSyncer) Start(ctx context.Context) {
	go ds.loop(ctx)
}

func (ds *DetailSyncer) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ds.notify:
		}
		ds.syncAll(ctx)
	}
}

func (ds *DetailSyncer) syncAll(ctx context.Context) {
	ds.srv.mu.RLock()
	prs := ds.srv.prs
	ds.srv.mu.RUnlock()

	if len(prs) == 0 {
		return
	}

	activeKeys := make(map[string]bool, len(prs))
	var toSync []github.PRListItem
	for _, pr := range prs {
		key := prKey(pr)
		activeKeys[key] = true
		if ds.cache.IsStale(key, pr.UpdatedAt) {
			toSync = append(toSync, pr)
			ds.cache.SetStatus(key, PRSyncStatus{State: SyncQueued})
		}
	}

	ds.cache.EvictStale(activeKeys)

	if len(toSync) == 0 {
		return
	}

	prioritize(toSync)
	ds.srv.notify()

	sem := make(chan struct{}, ds.maxWorkers)
	for _, pr := range toSync {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !ds.checkRateLimit() {
			logger.Infof("detail syncer: rate limit low, pausing sync cycle")
			return
		}

		sem <- struct{}{}
		go func(item github.PRListItem) {
			defer func() { <-sem }()
			ds.syncPR(item)
		}(pr)

		time.Sleep(200 * time.Millisecond)
	}

	// Wait for remaining workers
	for i := 0; i < ds.maxWorkers; i++ {
		sem <- struct{}{}
	}
}

func (ds *DetailSyncer) syncPR(pr github.PRListItem) {
	key := prKey(pr)
	ds.cache.SetStatus(key, PRSyncStatus{State: SyncSyncing, Phase: "metadata"})
	ds.srv.notify()

	detail := ds.srv.fetchPRDetail(pr.Repo, pr.Number)

	if detail.Error != "" {
		ds.cache.SetStatus(key, PRSyncStatus{
			State:      SyncError,
			LastSynced: time.Now(),
			Error:      detail.Error,
		})
		ds.srv.notify()
		return
	}

	ds.cache.Put(key, detail, pr.UpdatedAt)
	ds.cache.SetStatus(key, PRSyncStatus{
		State:      SyncUpToDate,
		LastSynced: time.Now(),
	})
	ds.srv.notify()
}

func (ds *DetailSyncer) checkRateLimit() bool {
	ds.srv.mu.RLock()
	rl := ds.srv.rateLimit
	ds.srv.mu.RUnlock()
	if rl == nil {
		return true
	}
	return rl.Remaining >= 1500
}

// prioritize sorts PRs: failing checks first, then open, then by recency.
func prioritize(prs []github.PRListItem) {
	sort.Slice(prs, func(i, j int) bool {
		a, b := prs[i], prs[j]
		ap, bp := syncPriority(a), syncPriority(b)
		if ap != bp {
			return ap < bp
		}
		return a.UpdatedAt.After(b.UpdatedAt)
	})
}

func syncPriority(pr github.PRListItem) int {
	if pr.CheckStatus != nil && pr.CheckStatus.Failed > 0 {
		return 0 // highest priority
	}
	if pr.State == "OPEN" && !pr.IsDraft {
		return 1
	}
	if pr.State == "OPEN" {
		return 2 // drafts
	}
	return 3 // merged/closed
}

func (ds *DetailSyncer) GetGHOpts() github.Options {
	return ds.ghOpts
}

// SyncStatusKey returns the cache key for a PR — exported for handleDetail.
func SyncStatusKey(repo string, number int) string {
	return fmt.Sprintf("%s#%d", repo, number)
}
