package cmux

import (
	"bufio"
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/ai/history"
	"github.com/flanksource/captain/pkg/ai/pricing"
)

// High-level agent states surfaced to the dashboard, derived from the last
// meaningful event in the session log. Mirrors the TS deriveSessionState mapping.
const (
	sessionStateThinking  = "thinking"
	sessionStateWorking   = "working"
	sessionStateAsk       = "ask"
	sessionStateCompleted = "completed"
)

// isAskTool reports whether a tool pauses the turn awaiting the user (Claude
// emits no terminal stop reason for these), so the session is "asking", not
// "working", until they respond.
func isAskTool(tool string) bool {
	switch tool {
	case "AskUserQuestion", "ExitPlanMode":
		return true
	default:
		return false
	}
}

// sessionStateFromLine maps the last event of one session-log line to the agent
// state it represents. Non-conversational lines (tool results, bookkeeping)
// yield no event and return ("", false) so the caller keeps the prior state.
func sessionStateFromLine(line []byte) (string, bool) {
	events, err := history.ParseSessionEvents(line)
	if err != nil || len(events) == 0 {
		return "", false
	}
	last := events[len(events)-1]
	switch last.Kind {
	case history.EventThinking:
		return sessionStateThinking, true
	case history.EventToolUse:
		if isAskTool(last.ToolUse.Tool) {
			return sessionStateAsk, true
		}
		return sessionStateWorking, true
	case history.EventTurnEnd:
		return sessionStateCompleted, true
	case history.EventAssistantText:
		return sessionStateWorking, true
	default:
		return "", false
	}
}

const (
	// sessionStatsTTL bounds how long a cold (disk-derived) stats entry is reused
	// before it is recomputed, even when the log's mtime has not changed.
	sessionStatsTTL = 5 * time.Second
	// sessionStatsMaxLive caps the live map; finished sessions are evicted
	// oldest-first beyond it so a long-lived dashboard does not grow unbounded.
	sessionStatsMaxLive = 128
	sessionStatsMaxLine = 10 * 1024 * 1024
)

// SessionStats is the rolled-up activity of one agent session: its identity
// (agent/model/effort), elapsed time, token usage and derived cost. It backs the
// dashboard's session timer in both the detail pane and the sidebar. Token totals
// are summed across every assistant turn in the session log; cost is derived from
// those totals via captain's pricing registry (best-effort — zero when the model
// is absent from the registry, mirroring the ai-fix context-window lookup).
type SessionStats struct {
	SessionID           string    `json:"sessionId,omitempty"`
	Agent               string    `json:"agent,omitempty"`
	Model               string    `json:"model,omitempty"`
	Effort              string    `json:"effort,omitempty"`
	StartedAt           time.Time `json:"startedAt,omitempty"`
	UpdatedAt           time.Time `json:"updatedAt,omitempty"`
	DurationMs          int64     `json:"durationMs"`
	InputTokens         int       `json:"inputTokens"`
	OutputTokens        int       `json:"outputTokens"`
	CacheReadTokens     int       `json:"cacheReadTokens"`
	CacheCreationTokens int       `json:"cacheCreationTokens"`
	TotalTokens         int       `json:"totalTokens"`
	Turns               int       `json:"turns"`
	CostUSD             float64   `json:"costUsd"`
	InProgress          bool      `json:"inProgress"`
	Found               bool      `json:"found"`
	// State is the high-level agent state from the most recent session-log event
	// (thinking / working / ask / completed); empty before the first event.
	State string `json:"state,omitempty"`
}

// sessionUsageLine is the subset of a Claude session-log entry needed for stats:
// the per-request token usage and model carried on each assistant turn.
type sessionUsageLine struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Model string `json:"model"`
		Usage struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

func (l sessionUsageLine) hasUsage() bool {
	u := l.Message.Usage
	return u.InputTokens > 0 || u.OutputTokens > 0 || u.CacheReadInputTokens > 0 || u.CacheCreationInputTokens > 0
}

// applyUsage folds one assistant entry's usage into the running totals.
func (s *SessionStats) applyUsage(l sessionUsageLine) {
	u := l.Message.Usage
	s.InputTokens += u.InputTokens
	s.OutputTokens += u.OutputTokens
	s.CacheReadTokens += u.CacheReadInputTokens
	s.CacheCreationTokens += u.CacheCreationInputTokens
	s.Turns++
	if l.Message.Model != "" {
		s.Model = l.Message.Model
	}
}

// finalize derives the totals, elapsed duration, and cost from the accumulated
// token counts and timestamps. The model is fixed within a session, so summing
// per-turn cost equals computing cost once from the totals.
func (s *SessionStats) finalize() {
	s.TotalTokens = s.InputTokens + s.OutputTokens + s.CacheReadTokens + s.CacheCreationTokens
	if !s.StartedAt.IsZero() && !s.UpdatedAt.IsZero() && s.UpdatedAt.After(s.StartedAt) {
		s.DurationMs = s.UpdatedAt.Sub(s.StartedAt).Milliseconds()
	}
	s.CostUSD = sessionCost(s.Model, s.InputTokens, s.OutputTokens, s.CacheReadTokens, s.CacheCreationTokens)
}

// sessionCost prices the session's tokens via captain's pricing registry. Claude
// session logs report bare model ids (e.g. "claude-opus-4-8") while the registry
// is keyed by OpenRouter ids ("anthropic/<model>"), so both forms are tried.
// An unknown model yields zero cost rather than failing — pricing is optional
// enrichment, not a correctness invariant.
func sessionCost(model string, in, out, cacheRead, cacheWrite int) float64 {
	if model == "" {
		return 0
	}
	for _, id := range []string{model, "anthropic/" + model} {
		if res, err := pricing.CalculateCost(id, in, out, 0, cacheRead, cacheWrite); err == nil {
			return res.TotalCost
		}
	}
	return 0
}

func parseSessionTime(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t
	}
	return time.Time{}
}

// computeSessionStats reads a complete session log from disk and rolls up its
// token usage, model, and first/last timestamps. Used for cold (non-in-progress)
// sessions the live tailers never observed.
func computeSessionStats(path string) (SessionStats, error) {
	f, err := os.Open(path)
	if err != nil {
		return SessionStats{}, err
	}
	defer func() { _ = f.Close() }()

	stats := SessionStats{Found: true, Agent: "claude"}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), sessionStatsMaxLine)
	var first, last time.Time
	for scanner.Scan() {
		if state, ok := sessionStateFromLine(scanner.Bytes()); ok {
			stats.State = state
		}
		var entry sessionUsageLine
		if json.Unmarshal(scanner.Bytes(), &entry) != nil {
			continue
		}
		if ts := parseSessionTime(entry.Timestamp); !ts.IsZero() {
			if first.IsZero() || ts.Before(first) {
				first = ts
			}
			if ts.After(last) {
				last = ts
			}
		}
		if entry.Type == "assistant" && entry.hasUsage() {
			stats.applyUsage(entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return SessionStats{}, err
	}
	stats.StartedAt = first
	stats.UpdatedAt = last
	stats.finalize()
	return stats, nil
}

// SessionAccumulator is the live, in-progress stats for one running session,
// fed a line at a time by the cmux tailer. It is registered in the cache so the
// dashboard reads live totals without re-reading the growing log on every poll.
type SessionAccumulator struct {
	mu    sync.Mutex
	stats SessionStats
}

// AddLine folds one raw session-log line into the running stats: its state (from
// the line's last event) and, for assistant turns, its token usage. Safe to use
// as the tailer's onLine hook; it never retains the slice.
func (a *SessionAccumulator) AddLine(line []byte) {
	state, hasState := sessionStateFromLine(line)

	var entry sessionUsageLine
	hasUsage := json.Unmarshal(line, &entry) == nil && entry.Type == "assistant" && entry.hasUsage()
	if !hasState && !hasUsage {
		return
	}
	a.mu.Lock()
	if hasState {
		a.stats.State = state
	}
	if hasUsage {
		a.stats.applyUsage(entry)
		a.stats.UpdatedAt = time.Now()
	}
	a.mu.Unlock()
}

// Finish marks the session no longer in progress, freezing its elapsed duration.
func (a *SessionAccumulator) Finish() {
	a.mu.Lock()
	a.stats.InProgress = false
	if a.stats.UpdatedAt.IsZero() {
		a.stats.UpdatedAt = time.Now()
	}
	a.mu.Unlock()
}

// snapshot returns a finalized copy. While in progress the elapsed clock runs to
// now so the dashboard timer keeps ticking between polls.
func (a *SessionAccumulator) snapshot() SessionStats {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.stats
	s.Found = true
	if s.InProgress {
		s.UpdatedAt = time.Now()
	}
	s.finalize()
	return s
}

func (a *SessionAccumulator) finished() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return !a.stats.InProgress
}

func (a *SessionAccumulator) startedAt() time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stats.StartedAt
}

// SessionStatsCache holds the live in-progress sessions (pushed by the tailers)
// and a cold cache of disk-derived stats for sessions no tailer is watching.
type SessionStatsCache struct {
	mu   sync.Mutex
	live map[string]*SessionAccumulator
	cold map[string]coldStatsEntry
}

type coldStatsEntry struct {
	stats      SessionStats
	mtime      time.Time
	computedAt time.Time
}

func NewSessionStatsCache() *SessionStatsCache {
	return &SessionStatsCache{
		live: map[string]*SessionAccumulator{},
		cold: map[string]coldStatsEntry{},
	}
}

// Begin registers a live session and returns the accumulator the tailer feeds.
// agent/model/effort are the run's configured values; model is later refined to
// the concrete id reported in the log as turns arrive.
func (c *SessionStatsCache) Begin(sessionID, agent, model, effort string, start time.Time) *SessionAccumulator {
	acc := &SessionAccumulator{stats: SessionStats{
		SessionID:  sessionID,
		Agent:      agent,
		Model:      model,
		Effort:     effort,
		StartedAt:  start,
		UpdatedAt:  start,
		InProgress: true,
	}}
	c.mu.Lock()
	c.live[sessionID] = acc
	c.evictLiveLocked()
	c.mu.Unlock()
	return acc
}

// Get returns the stats for a session: the live accumulator when one is present,
// otherwise a cold read of the on-disk log cached by mtime + TTL. A missing log
// (session never produced output) is the normal "not found" state, returned as a
// zero SessionStats with Found=false rather than an error.
func (c *SessionStatsCache) Get(sessionID, path string) (SessionStats, error) {
	c.mu.Lock()
	acc := c.live[sessionID]
	c.mu.Unlock()
	if acc != nil {
		return acc.snapshot(), nil
	}
	return c.coldStats(sessionID, path)
}

func (c *SessionStatsCache) coldStats(sessionID, path string) (SessionStats, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SessionStats{}, nil
		}
		return SessionStats{}, err
	}

	c.mu.Lock()
	if entry, ok := c.cold[path]; ok && entry.mtime.Equal(info.ModTime()) && time.Since(entry.computedAt) < sessionStatsTTL {
		c.mu.Unlock()
		return entry.stats, nil
	}
	c.mu.Unlock()

	stats, err := computeSessionStats(path)
	if err != nil {
		return SessionStats{}, err
	}
	stats.SessionID = sessionID

	c.mu.Lock()
	c.cold[path] = coldStatsEntry{stats: stats, mtime: info.ModTime(), computedAt: time.Now()}
	c.mu.Unlock()
	return stats, nil
}

// evictLiveLocked drops the oldest finished sessions once the live map exceeds
// the cap; in-progress sessions are never evicted. Caller holds c.mu.
func (c *SessionStatsCache) evictLiveLocked() {
	if len(c.live) <= sessionStatsMaxLive {
		return
	}
	type aged struct {
		id    string
		start time.Time
	}
	finished := make([]aged, 0, len(c.live))
	for id, acc := range c.live {
		if acc.finished() {
			finished = append(finished, aged{id: id, start: acc.startedAt()})
		}
	}
	sort.Slice(finished, func(i, j int) bool { return finished[i].start.Before(finished[j].start) })
	for _, f := range finished {
		if len(c.live) <= sessionStatsMaxLive {
			break
		}
		delete(c.live, f.id)
	}
}

// defaultSessionStats is the process-wide cache: the cmux executor (running
// in-process under the dashboard) feeds live sessions into it and the dashboard's
// stats endpoint reads from it.
var defaultSessionStats = NewSessionStatsCache()

// GlobalSessionStats returns the process-wide session stats cache.
func GlobalSessionStats() *SessionStatsCache { return defaultSessionStats }
