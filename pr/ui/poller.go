package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
)

type Poller struct {
	srv         *Server
	searchFn    func(since time.Time) (github.PRSearchResults, *github.RateLimit, error)
	interval    time.Duration
	fullRefresh time.Duration
	known       map[string]github.PRListItem
	lastFetch   time.Time
	lastFull    time.Time
}

func NewPoller(srv *Server, searchFn func(since time.Time) (github.PRSearchResults, *github.RateLimit, error), interval time.Duration) *Poller {
	return &Poller{
		srv:         srv,
		searchFn:    searchFn,
		interval:    interval,
		fullRefresh: 5 * time.Minute,
		known:       make(map[string]github.PRListItem),
	}
}

func prKey(item github.PRListItem) string {
	return fmt.Sprintf("%s#%d", item.Repo, item.Number)
}

func (p *Poller) Start(ctx context.Context) {
	go p.loop(ctx)
}

func (p *Poller) loop(ctx context.Context) {
	p.fetchFull()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.srv.RefreshCh():
			p.fetchFull()
		case <-ticker.C:
			if p.srv.IsPaused() {
				continue
			}
			if time.Since(p.lastFull) >= p.fullRefresh {
				p.fetchFull()
			} else {
				p.fetchIncremental()
			}
		}
	}
}

func (p *Poller) fetchFull() {
	logger.Debugf("pr poller: full fetch")
	results, rl, err := p.searchFn(time.Time{})
	p.srv.SetRateLimit(rl)
	if err != nil {
		logger.Warnf("pr poller: full fetch failed: %v", err)
		p.srv.SetError(err)
		return
	}

	p.known = make(map[string]github.PRListItem, len(results))
	for _, item := range results {
		p.known[prKey(item)] = item
	}
	p.lastFetch = time.Now()
	p.lastFull = time.Now()
	p.srv.SetResults(p.knownAsList(), false)
}

func (p *Poller) fetchIncremental() {
	since := p.lastFetch.Add(-30 * time.Second)
	logger.Debugf("pr poller: incremental fetch since %s", since.Format(time.RFC3339))

	results, rl, err := p.searchFn(since)
	p.srv.SetRateLimit(rl)
	if err != nil {
		logger.Warnf("pr poller: incremental fetch failed: %v", err)
		p.srv.SetError(err)
		return
	}

	for _, item := range results {
		p.known[prKey(item)] = item
	}
	p.lastFetch = time.Now()
	p.srv.SetResults(p.knownAsList(), true)
}

func (p *Poller) knownAsList() github.PRSearchResults {
	result := make(github.PRSearchResults, 0, len(p.known))
	for _, item := range p.known {
		result = append(result, item)
	}
	return result
}
