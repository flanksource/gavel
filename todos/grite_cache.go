package todos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// defaultSyncTTL bounds how often a read re-syncs from grite. Within the TTL a
// read is served straight from the cache DB without spawning grite.
const defaultSyncTTL = 3 * time.Second

// EnvSyncTTL overrides defaultSyncTTL with a Go duration string.
const EnvSyncTTL = "GAVEL_TODOS_SYNC_TTL"

// CachedIssue is the persisted projection of a grite issue plus its accumulated
// event history (JSON), enough to rebuild a *types.TODO identically to a direct
// `grite issue show`.
type CachedIssue struct {
	Repo         string
	IssueID      string
	Title        string
	State        string
	Labels       []string
	CreatedTS    int64
	UpdatedTS    int64
	CommentCount int
	// EventsJSON is the full accumulated event history for the issue, encoded as
	// a JSON array of grite events (deduped by event_id, ascending by timestamp).
	EventsJSON []byte
}

func (ci CachedIssue) issue() griteIssue {
	return griteIssue{
		IssueID:      ci.IssueID,
		Title:        ci.Title,
		State:        ci.State,
		Labels:       ci.Labels,
		CreatedTS:    ci.CreatedTS,
		UpdatedTS:    ci.UpdatedTS,
		CommentCount: ci.CommentCount,
	}
}

func (ci CachedIssue) events() ([]griteEvent, error) {
	if len(ci.EventsJSON) == 0 {
		return nil, nil
	}
	var events []griteEvent
	if err := json.Unmarshal(ci.EventsJSON, &events); err != nil {
		return nil, fmt.Errorf("decode cached events for %s: %w", ci.IssueID, err)
	}
	return events, nil
}

// GriteCacheStore persists grite issue projections in the gavel DB. A disabled
// store reports Enabled()==false, so ResolveGriteProvider falls back to direct grite.
type GriteCacheStore interface {
	Enabled() bool
	LoadCursor(ctx context.Context, repo string) (lastEventTS int64, syncedAt time.Time, err error)
	ListIssues(ctx context.Context, repo string) ([]CachedIssue, error)
	GetIssue(ctx context.Context, repo, issueID string) (*CachedIssue, error)
	UpsertSync(ctx context.Context, repo string, issues []CachedIssue, lastEventTS int64) error
}

// CachedGriteProvider serves reads from the gavel DB and keeps it fresh with a
// TTL-guarded incremental `grite export --since`. Writes pass through to the inner
// GriteProvider (grite stays the source of truth) and then force a sync so the next
// read reflects them.
type CachedGriteProvider struct {
	inner *GriteProvider
	store GriteCacheStore
	repo  string
	ttl   time.Duration
}

func NewCachedGriteProvider(inner *GriteProvider, store GriteCacheStore, repo string, ttl time.Duration) *CachedGriteProvider {
	return &CachedGriteProvider{inner: inner, store: store, repo: cacheRepoKey(repo), ttl: resolveSyncTTL(ttl)}
}

// ResolveGriteProvider returns a DB-backed CachedGriteProvider when store is enabled,
// otherwise a plain GriteProvider that talks to grite directly. Both share the same
// write path, so grite remains the source of truth either way.
func ResolveGriteProvider(dir string, store GriteCacheStore, ttl time.Duration) Provider {
	inner := NewGriteProvider(dir)
	if store != nil && store.Enabled() {
		return NewCachedGriteProvider(inner, store, dir, ttl)
	}
	return inner
}

func resolveSyncTTL(ttl time.Duration) time.Duration {
	if ttl > 0 {
		return ttl
	}
	if v := os.Getenv(EnvSyncTTL); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return defaultSyncTTL
}

// cacheRepoKey normalizes a workspace directory into a stable cache key.
func cacheRepoKey(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		return filepath.Clean(abs)
	}
	return dir
}

func (p *CachedGriteProvider) List(ctx context.Context, filters DiscoveryFilters) (types.TODOS, error) {
	if err := p.ensureSynced(ctx); err != nil {
		return nil, err
	}
	cached, err := p.store.ListIssues(ctx, p.repo)
	if err != nil {
		return nil, err
	}
	var out types.TODOS
	for _, ci := range cached {
		todo := todoFromGriteIssue(ci.issue(), p.inner.WorkDir)
		if filters.Matches(todo) {
			out = append(out, todo)
		}
	}
	out.Sort()
	return out, nil
}

func (p *CachedGriteProvider) Get(ctx context.Context, ref string) (*types.TODO, error) {
	if err := p.ensureSynced(ctx); err != nil {
		return nil, err
	}
	ci, err := p.resolveCachedIssue(ctx, ref)
	if err != nil {
		return nil, err
	}
	events, err := ci.events()
	if err != nil {
		return nil, err
	}
	issue := ci.issue()
	defaults := frontmatterFromGriteIssue(issue, p.inner.WorkDir)
	todo, err := ParseTODOContent(issue.Title, bodyFromGriteEvents(events), p.inner.WorkDir, defaults)
	if err != nil {
		return nil, err
	}
	applyGriteIdentity(todo, issue)
	todo.ProviderEvents = providerEventsFromGriteEvents(events)
	return todo, nil
}

// resolveCachedIssue locates the cached issue identified by ref. An exact issue
// id wins; otherwise ref is treated as a short-id prefix (e.g. the 8-char
// DisplayID shown by `todo list`), mirroring grite's own `issue show` prefix
// resolution. An ambiguous prefix is reported rather than silently picking one.
func (p *CachedGriteProvider) resolveCachedIssue(ctx context.Context, ref string) (*CachedIssue, error) {
	ci, err := p.store.GetIssue(ctx, p.repo, ref)
	if err != nil {
		return nil, err
	}
	if ci != nil {
		return ci, nil
	}
	issues, err := p.store.ListIssues(ctx, p.repo)
	if err != nil {
		return nil, err
	}
	var matches []CachedIssue
	for _, candidate := range issues {
		if strings.HasPrefix(candidate.IssueID, ref) {
			matches = append(matches, candidate)
		}
	}
	switch len(matches) {
	case 1:
		return &matches[0], nil
	case 0:
		return nil, fmt.Errorf("todo not found: %s", ref)
	default:
		return nil, fmt.Errorf("ambiguous todo id %q matches %d issues: %s", ref, len(matches), strings.Join(ambiguousMatchLabels(matches), "; "))
	}
}

func ambiguousMatchLabels(issues []CachedIssue) []string {
	out := make([]string, len(issues))
	for i, ci := range issues {
		out[i] = fmt.Sprintf("%s %s", shortGriteID(ci.IssueID), ci.Title)
	}
	return out
}

func (p *CachedGriteProvider) Create(ctx context.Context, req CreateRequest) (*types.TODO, error) {
	return p.passThrough(ctx, func() (*types.TODO, error) { return p.inner.Create(ctx, req) })
}

func (p *CachedGriteProvider) Delete(ctx context.Context, todo *types.TODO) error {
	if err := p.inner.Delete(ctx, todo); err != nil {
		return err
	}
	return p.syncNow(ctx)
}

func (p *CachedGriteProvider) Edit(ctx context.Context, todo *types.TODO, edit EditRequest) error {
	if err := p.inner.Edit(ctx, todo, edit); err != nil {
		return err
	}
	return p.syncNow(ctx)
}

func (p *CachedGriteProvider) Comment(ctx context.Context, todo *types.TODO, body string) error {
	if err := p.inner.Comment(ctx, todo, body); err != nil {
		return err
	}
	return p.syncNow(ctx)
}

func (p *CachedGriteProvider) UpdateState(ctx context.Context, todo *types.TODO, updates StateUpdate) error {
	if err := p.inner.UpdateState(ctx, todo, updates); err != nil {
		return err
	}
	return p.syncNow(ctx)
}

func (p *CachedGriteProvider) UpdateLatestFailure(ctx context.Context, todo *types.TODO, result *types.TestResultInfo) error {
	if err := p.inner.UpdateLatestFailure(ctx, todo, result); err != nil {
		return err
	}
	return p.syncNow(ctx)
}

func (p *CachedGriteProvider) SaveAttempt(ctx context.Context, todo *types.TODO, result *ExecutionResult) error {
	if err := p.inner.SaveAttempt(ctx, todo, result); err != nil {
		return err
	}
	return p.syncNow(ctx)
}

func (p *CachedGriteProvider) SaveVerification(ctx context.Context, todo *types.TODO, result *verify.VerifyResult) error {
	if err := p.inner.SaveVerification(ctx, todo, result); err != nil {
		return err
	}
	return p.syncNow(ctx)
}

func (p *CachedGriteProvider) passThrough(ctx context.Context, write func() (*types.TODO, error)) (*types.TODO, error) {
	todo, err := write()
	if err != nil {
		return nil, err
	}
	if err := p.syncNow(ctx); err != nil {
		return nil, err
	}
	return todo, nil
}

// Per-repo locks coalesce concurrent syncs so a burst of stale reads spawns a
// single `grite export` rather than one per request.
var (
	syncLocksMu sync.Mutex
	syncLocks   = map[string]*sync.Mutex{}
)

func repoSyncLock(repo string) *sync.Mutex {
	syncLocksMu.Lock()
	defer syncLocksMu.Unlock()
	m := syncLocks[repo]
	if m == nil {
		m = &sync.Mutex{}
		syncLocks[repo] = m
	}
	return m
}

func (p *CachedGriteProvider) ensureSynced(ctx context.Context) error {
	if fresh, err := p.isFresh(ctx); err != nil || fresh {
		return err
	}
	lock := repoSyncLock(p.repo)
	lock.Lock()
	defer lock.Unlock()
	// Re-check under the lock: another goroutine may have synced while we waited.
	if fresh, err := p.isFresh(ctx); err != nil || fresh {
		return err
	}
	return p.doSync(ctx)
}

func (p *CachedGriteProvider) syncNow(ctx context.Context) error {
	lock := repoSyncLock(p.repo)
	lock.Lock()
	defer lock.Unlock()
	return p.doSync(ctx)
}

func (p *CachedGriteProvider) isFresh(ctx context.Context) (bool, error) {
	_, syncedAt, err := p.store.LoadCursor(ctx, p.repo)
	if err != nil {
		return false, err
	}
	return !syncedAt.IsZero() && time.Since(syncedAt) < p.ttl, nil
}

func (p *CachedGriteProvider) doSync(ctx context.Context) error {
	cursor, _, err := p.store.LoadCursor(ctx, p.repo)
	if err != nil {
		return err
	}
	export, err := p.exportSince(ctx, cursor)
	if err != nil {
		return err
	}
	existing, err := p.store.ListIssues(ctx, p.repo)
	if err != nil {
		return err
	}
	issues, newCursor := mergeExport(p.repo, cursor, existing, export)
	return p.store.UpsertSync(ctx, p.repo, issues, newCursor)
}

// griteExportResult is the envelope payload of `grite export --json`; the issue
// and event data is written to OutputPath rather than stdout.
type griteExportResult struct {
	OutputPath string `json:"output_path"`
	EventCount int    `json:"event_count"`
}

// griteExportFile is the on-disk export written by grite: a full snapshot of all
// issues plus the events since the requested cursor.
type griteExportFile struct {
	Issues []griteIssue `json:"issues"`
	Events []griteEvent `json:"events"`
}

func (p *CachedGriteProvider) exportSince(ctx context.Context, cursor int64) (griteExportFile, error) {
	raw, err := p.inner.run(ctx, "export", "--format", "json", "--json", "--since", strconv.FormatInt(cursor, 10))
	if err != nil {
		return griteExportFile{}, err
	}
	res, err := decodeGrite[griteExportResult](raw)
	if err != nil {
		return griteExportFile{}, err
	}
	if res.OutputPath == "" {
		return griteExportFile{}, fmt.Errorf("grite export returned no output path")
	}
	data, err := os.ReadFile(res.OutputPath)
	if err != nil {
		return griteExportFile{}, fmt.Errorf("read grite export %s: %w", res.OutputPath, err)
	}
	var file griteExportFile
	if err := json.Unmarshal(data, &file); err != nil {
		return griteExportFile{}, fmt.Errorf("decode grite export %s: %w", res.OutputPath, err)
	}
	return file, nil
}

// mergeExport applies a (full issue snapshot + incremental event delta) export onto
// the existing cache: every issue is re-projected from the snapshot and its event
// history is the union of prior events and the new delta (deduped, time-ordered).
// The returned cursor is the max timestamp observed so the next export resumes there.
func mergeExport(repo string, cursor int64, existing []CachedIssue, export griteExportFile) ([]CachedIssue, int64) {
	prior := make(map[string][]griteEvent, len(existing))
	for _, ci := range existing {
		if ev, err := ci.events(); err == nil {
			prior[ci.IssueID] = ev
		}
	}
	newByIssue := make(map[string][]griteEvent, len(export.Events))
	newCursor := cursor
	for _, ev := range export.Events {
		newByIssue[ev.IssueID] = append(newByIssue[ev.IssueID], ev)
		if ev.TimestampMS > newCursor {
			newCursor = ev.TimestampMS
		}
	}
	out := make([]CachedIssue, 0, len(export.Issues))
	for _, issue := range export.Issues {
		merged := mergeEvents(prior[issue.IssueID], newByIssue[issue.IssueID])
		eventsJSON, _ := json.Marshal(merged)
		out = append(out, CachedIssue{
			Repo:         repo,
			IssueID:      issue.IssueID,
			Title:        issue.Title,
			State:        issue.State,
			Labels:       append([]string(nil), issue.Labels...),
			CreatedTS:    issue.CreatedTS,
			UpdatedTS:    issue.UpdatedTS,
			CommentCount: issue.CommentCount,
			EventsJSON:   eventsJSON,
		})
		if issue.UpdatedTS > newCursor {
			newCursor = issue.UpdatedTS
		}
	}
	return out, newCursor
}

func mergeEvents(prior, incoming []griteEvent) []griteEvent {
	seen := make(map[string]bool, len(prior)+len(incoming))
	merged := make([]griteEvent, 0, len(prior)+len(incoming))
	for _, batch := range [][]griteEvent{prior, incoming} {
		for _, ev := range batch {
			if seen[ev.EventID] {
				continue
			}
			seen[ev.EventID] = true
			merged = append(merged, ev)
		}
	}
	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].TimestampMS < merged[j].TimestampMS
	})
	return merged
}
