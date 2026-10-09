package gitstate

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

const (
	// DefaultHot is how often a worktree someone is looking at, or an agent is
	// working in, has its status rescanned.
	DefaultHot = 5 * time.Second
	// DefaultIdle is how soon every other worktree's status, and every repo's
	// refs, are rescanned after a scan that found a change; each scan finding
	// nothing doubles it, up to DefaultMaxIdle. .git metadata events rescan
	// sooner.
	DefaultIdle = time.Minute
	// DefaultMaxIdle caps the idle interval of a long-unchanged key.
	DefaultMaxIdle = 30 * time.Minute
	// DefaultDebounce coalesces a burst of .git metadata events into one scan.
	DefaultDebounce = 250 * time.Millisecond
	// DefaultConcurrency bounds the scans running at once.
	DefaultConcurrency = 4
)

// Options configures a Tracker. Zero durations and concurrency take the
// defaults.
type Options struct {
	Store *Store
	// Context bounds every background scan: the process's root context.
	Context     context.Context
	Hot         time.Duration
	Idle        time.Duration
	MaxIdle     time.Duration
	Debounce    time.Duration
	Concurrency int
}

// Tracker keeps the git_* rows of every repository it has been asked about
// current. Reads go to the Store; git runs only in scans, which a ref or
// index change under .git, a Touch after a server-side mutation, or the
// per-worktree hot/idle cadence schedules. Scans of one key coalesce: a
// request while one is queued joins it, a request while one runs queues at
// most one more.
type Tracker struct {
	store              *Store
	ctx                context.Context
	hot, idle, maxIdle time.Duration
	debounce           time.Duration
	sem                chan struct{}
	resolve            singleflight.Group
	watcher            *metadataWatcher

	mu      sync.Mutex
	repos   map[string]*trackedRepo // by root dir
	byDir   map[string]*trackedRepo // by the directory a caller named
	jobs    map[jobKey]*jobState
	focus   map[string]time.Time // worktree path → lease expiry
	active  []string             // directories agents are working in
	started bool
	// unlocatable remembers directories that are not in a git repository.
	unlocatable map[string]locateFailure
}

type locateFailure struct {
	err   error
	until time.Time
}

type trackedRepo struct {
	id        uuid.UUID
	root      string
	commonDir string
	// worktrees are the live, non-prunable worktree paths of the last ref
	// scan; status/refs are when each is next due.
	worktrees []string
	status    map[string]*pollSchedule
	refs      pollSchedule
}

// jobKey names a scan: the repo's refs when worktree is "", else one
// worktree's status.
type jobKey struct {
	root     string
	worktree string
}

// jobState is a key's background scan: running from enqueue until its drain
// returns, started while it holds a slot, pending when requested again after
// it started.
type jobState struct {
	running, started, pending bool
}

func NewTracker(opts Options) *Tracker {
	t := &Tracker{
		store: opts.Store, ctx: opts.Context, hot: opts.Hot, idle: opts.Idle, maxIdle: opts.MaxIdle, debounce: opts.Debounce,
		repos: map[string]*trackedRepo{}, byDir: map[string]*trackedRepo{}, jobs: map[jobKey]*jobState{},
		focus: map[string]time.Time{}, unlocatable: map[string]locateFailure{},
	}
	if t.ctx == nil {
		t.ctx = context.Background()
	}
	if t.hot == 0 {
		t.hot = DefaultHot
	}
	if t.idle == 0 {
		t.idle = DefaultIdle
	}
	if t.maxIdle == 0 {
		t.maxIdle = DefaultMaxIdle
	}
	if t.debounce == 0 {
		t.debounce = DefaultDebounce
	}
	concurrency := opts.Concurrency
	if concurrency == 0 {
		concurrency = DefaultConcurrency
	}
	t.sem = make(chan struct{}, concurrency)
	return t
}

// Store is the tracker's store, for reads that need no scheduling.
func (t *Tracker) Store() *Store {
	return t.store
}

// Start begins watching .git metadata and the hot/idle cadence. Repos are
// added as callers Track them. It stops when the tracker's context ends.
func (t *Tracker) Start() error {
	watcher, err := newMetadataWatcher(t)
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.watcher, t.started = watcher, true
	repos := make([]*trackedRepo, 0, len(t.repos))
	for _, repo := range t.repos {
		repos = append(repos, repo)
	}
	t.mu.Unlock()
	for _, repo := range repos {
		if err := watcher.add(repo); err != nil {
			return err
		}
	}
	go watcher.run(t.ctx)
	go t.cadence()
	return nil
}

// Track registers the repository containing dir and returns its id. The first
// Track of a repository in this process scans it before returning, so a
// read that follows sees rows no older than the process.
func (t *Tracker) Track(ctx context.Context, dir string) (uuid.UUID, error) {
	dir = filepath.Clean(dir)
	t.mu.Lock()
	repo := t.byDir[dir]
	t.mu.Unlock()
	if repo != nil {
		return repo.id, nil
	}
	v, err, _ := t.resolve.Do(dir, func() (any, error) {
		return t.register(ctx, dir)
	})
	if err != nil {
		return uuid.Nil, err
	}
	return v.(*trackedRepo).id, nil
}

func (t *Tracker) register(ctx context.Context, dir string) (*trackedRepo, error) {
	t.mu.Lock()
	failed, ok := t.unlocatable[dir]
	t.mu.Unlock()
	if ok && time.Now().Before(failed.until) {
		return nil, failed.err
	}
	loc, err := locateRepo(ctx, dir)
	if err != nil {
		// Remembered for a while: a directory that is not a git work tree would
		// otherwise run git again on every read that names it.
		t.mu.Lock()
		t.unlocatable[dir] = locateFailure{err: err, until: time.Now().Add(t.idle)}
		t.mu.Unlock()
		return nil, err
	}
	t.mu.Lock()
	repo, known := t.repos[loc.RootDir]
	t.mu.Unlock()
	if !known {
		id, err := t.store.EnsureRepo(ctx, loc.RootDir, loc.CommonDir)
		if err != nil {
			return nil, err
		}
		repo = &trackedRepo{id: id, root: loc.RootDir, commonDir: loc.CommonDir, status: map[string]*pollSchedule{}}
		// A failed scan is recorded on the repo's rows (State.Error,
		// Worktree.StatusError) for readers to show; the repo is tracked
		// regardless, so the cadence and watcher retry it instead of every read.
		if err := t.scanRepo(ctx, repo); err != nil {
			logger.Warnf("git state: first scan of %s: %v", loc.RootDir, err)
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if existing, ok := t.repos[loc.RootDir]; ok {
		repo = existing
	} else {
		t.repos[loc.RootDir] = repo
		trackedRepos.Set(float64(len(t.repos)))
		if t.started {
			if err := t.watcher.add(repo); err != nil {
				return nil, err
			}
		}
	}
	t.byDir[dir] = repo
	return repo, nil
}

// scanRepo scans refs and then the status of every worktree, waiting for
// another process's scan of the same key instead of skipping it.
func (t *Tracker) scanRepo(ctx context.Context, repo *trackedRepo) error {
	if _, err := t.runRefs(ctx, repo, true); err != nil {
		return err
	}
	var errs []error
	for _, path := range t.worktreesOf(repo) {
		if err := t.runStatus(ctx, repo, path, true); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Touch schedules an immediate rescan of the refs of the repository owning
// path and of the worktree at or above path: what a server-side git mutation
// calls once it has changed either. A path no tracked repository owns has no
// rows to refresh.
func (t *Tracker) Touch(path string) {
	repo, worktree := t.owner(path)
	if repo == nil {
		logger.Debugf("git state: touch of untracked path %s", path)
		return
	}
	t.enqueue(jobKey{root: repo.root}, sourceTouch)
	if worktree != "" {
		t.enqueue(jobKey{root: repo.root, worktree: worktree}, sourceTouch)
	}
}

// Focus marks the worktree at path as being looked at until ttl passes, so its
// status is rescanned at the hot cadence; renewing the lease extends it.
func (t *Tracker) Focus(path string, ttl time.Duration) {
	repo, worktree := t.owner(path)
	if repo == nil || worktree == "" {
		return
	}
	t.mu.Lock()
	expired := t.focus[worktree].Before(time.Now())
	t.focus[worktree] = time.Now().Add(ttl)
	t.mu.Unlock()
	if expired {
		t.enqueue(jobKey{root: repo.root, worktree: worktree}, sourceFocus)
	}
}

// SetActive replaces the directories agents are working in; the worktrees
// containing them are rescanned at the hot cadence.
func (t *Tracker) SetActive(dirs []string) {
	cleaned := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if dir != "" {
			cleaned = append(cleaned, filepath.Clean(dir))
		}
	}
	t.mu.Lock()
	t.active = cleaned
	t.mu.Unlock()
}

// owner finds the tracked repo owning path and the live worktree containing
// it, the deepest one when worktrees nest.
func (t *Tracker) owner(path string) (*trackedRepo, string) {
	path = filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	var (
		owner *trackedRepo
		best  string
	)
	for _, repo := range t.repos {
		for _, wt := range repo.worktrees {
			if pathWithin(path, wt) && len(wt) > len(best) {
				owner, best = repo, wt
			}
		}
		if owner == nil && pathWithin(path, repo.commonDir) {
			owner = repo
		}
	}
	return owner, best
}

func (t *Tracker) worktreesOf(repo *trackedRepo) []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), repo.worktrees...)
}

// runRefs scans the repo's refs and reports whether its rows changed. When
// another process holds the scan lock and wait is unset, it scans nothing:
// that process writes the rows.
func (t *Tracker) runRefs(ctx context.Context, repo *trackedRepo, wait bool) (bool, error) {
	var (
		result  refsResult
		scanErr error
	)
	start := time.Now()
	ran, err := t.store.Locked(ctx, "git_refs:"+repo.id.String(), wait, func(tx *Store) error {
		result, scanErr = scanRefs(ctx, tx, repo.id, repo.root)
		return nil
	})
	observeSince(refsScanDuration.WithLabelValues(scanResult(ran, result.Changed, errors.Join(err, scanErr))), start)
	t.mu.Lock()
	repo.refs.record(time.Now(), result.Changed, t.idle, t.maxIdle)
	t.mu.Unlock()
	if err != nil || !ran {
		return false, err
	}
	if scanErr != nil {
		if recordErr := t.store.RecordError(ctx, repo.id, scanErr); recordErr != nil {
			return false, errors.Join(scanErr, recordErr)
		}
		return false, fmt.Errorf("scan refs of %s: %w", repo.root, scanErr)
	}
	live := make([]string, 0, len(result.Worktrees))
	for _, wt := range result.Worktrees {
		if !wt.Prunable {
			live = append(live, filepath.Clean(wt.Path))
		}
	}
	t.mu.Lock()
	repo.worktrees = live
	t.mu.Unlock()
	return result.Changed, nil
}

func (t *Tracker) runStatus(ctx context.Context, repo *trackedRepo, path string, wait bool) error {
	config := currentPollConfig()
	var changed bool
	start := time.Now()
	ran, err := t.store.Locked(ctx, "git_status:"+repo.id.String()+":"+path, wait, func(tx *Store) error {
		var err error
		changed, err = scanStatus(ctx, tx, repo.id, path, config.gitArgs())
		return err
	})
	observeSince(statusScanDuration.WithLabelValues(
		scanResult(ran, changed, err), strconv.FormatBool(config.FSMonitor), strconv.FormatBool(config.UntrackedCache),
	), start)
	t.mu.Lock()
	schedule := repo.status[path]
	if schedule == nil {
		schedule = &pollSchedule{}
		repo.status[path] = schedule
	}
	schedule.record(time.Now(), changed, t.idle, t.maxIdle)
	t.mu.Unlock()
	if err == nil {
		return nil
	}
	err = fmt.Errorf("scan status of %s: %w", path, err)
	if recordErr := t.store.RecordStatusError(ctx, repo.id, path, err); recordErr != nil {
		return errors.Join(err, recordErr)
	}
	return err
}

// enqueue runs the scan named key in the background. A request while that
// scan still waits for a slot joins it, since it has yet to read anything;
// one while it runs schedules it once more after.
func (t *Tracker) enqueue(key jobKey, source string) {
	scansEnqueued.WithLabelValues(jobKind(key), source).Inc()
	t.mu.Lock()
	state := t.jobs[key]
	if state == nil {
		state = &jobState{}
		t.jobs[key] = state
	}
	if state.running {
		state.pending = state.pending || state.started
		t.mu.Unlock()
		return
	}
	state.running = true
	t.mu.Unlock()
	go t.drain(key, state)
}

func (t *Tracker) drain(key jobKey, state *jobState) {
	for {
		queued := time.Now()
		select {
		case t.sem <- struct{}{}:
		case <-t.ctx.Done():
			return
		}
		observeSince(scanQueueWait.WithLabelValues(jobKind(key)), queued)
		t.mu.Lock()
		state.started = true
		t.mu.Unlock()
		scansInFlight.Inc()
		t.run(key)
		scansInFlight.Dec()
		<-t.sem
		t.mu.Lock()
		state.started = false
		if !state.pending {
			state.running = false
			t.mu.Unlock()
			return
		}
		state.pending = false
		t.mu.Unlock()
	}
}

func (t *Tracker) run(key jobKey) {
	t.mu.Lock()
	repo := t.repos[key.root]
	t.mu.Unlock()
	if repo == nil {
		return
	}
	var err error
	if key.worktree == "" {
		var changed bool
		if changed, err = t.runRefs(t.ctx, repo, false); changed {
			// A moved HEAD changes what is uncommitted; a new worktree has no
			// status yet.
			for _, path := range t.worktreesOf(repo) {
				t.enqueue(jobKey{root: repo.root, worktree: path}, sourceRefs)
			}
		}
	} else {
		err = t.runStatus(t.ctx, repo, key.worktree, false)
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Warnf("git state: %v", err)
	}
}

// cadence schedules the scans whose interval has passed: a worktree's status
// every hot interval while it is focused or an agent works in it, and a repo's
// refs while one of its worktrees is focused; every other key once its
// backed-off idle interval ends.
func (t *Tracker) cadence() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-t.ctx.Done():
			return
		case now := <-ticker.C:
			for _, key := range t.due(now) {
				t.enqueue(key, sourceCadence)
			}
		}
	}
}

func (t *Tracker) due(now time.Time) []jobKey {
	t.mu.Lock()
	defer t.mu.Unlock()
	var keys []jobKey
	counts := map[string]int{cadenceHot: 0, cadenceIdle: 0, cadenceBackoff: 0}
	for _, repo := range t.repos {
		refsDue := !now.Before(repo.refs.next)
		if t.focusedLocked(repo, now) {
			refsDue = now.Sub(repo.refs.last) >= t.hot
		}
		if refsDue {
			keys = append(keys, jobKey{root: repo.root})
		}
		for _, path := range repo.worktrees {
			schedule := repo.status[path]
			if schedule == nil {
				schedule = &pollSchedule{}
			}
			var due bool
			switch {
			case t.hotLocked(path, now):
				counts[cadenceHot]++
				due = now.Sub(schedule.last) >= t.hot
			case schedule.streak > 0:
				counts[cadenceBackoff]++
				due = !now.Before(schedule.next)
			default:
				counts[cadenceIdle]++
				due = !now.Before(schedule.next)
			}
			if due {
				keys = append(keys, jobKey{root: repo.root, worktree: path})
			}
		}
	}
	for cadence, count := range counts {
		trackedWorktrees.WithLabelValues(cadence).Set(float64(count))
	}
	return keys
}

func (t *Tracker) hotLocked(worktree string, now time.Time) bool {
	if t.focus[worktree].After(now) {
		return true
	}
	for _, dir := range t.active {
		if pathWithin(dir, worktree) {
			return true
		}
	}
	return false
}

// focusedLocked reports whether someone is looking at one of repo's worktrees.
func (t *Tracker) focusedLocked(repo *trackedRepo, now time.Time) bool {
	for _, path := range repo.worktrees {
		if t.focus[path].After(now) {
			return true
		}
	}
	return false
}

// pathWithin reports whether path is root or below it.
func pathWithin(path, root string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}
