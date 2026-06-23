package todos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/todos/types"
)

// fakeGriteStore is an in-memory GriteCacheStore for exercising the cached
// provider without a database.
type fakeGriteStore struct {
	enabled  bool
	issues   map[string]CachedIssue
	cursorTS int64
	syncedAt time.Time
	upserts  int
}

func newFakeGriteStore() *fakeGriteStore {
	return &fakeGriteStore{enabled: true, issues: map[string]CachedIssue{}}
}

func (f *fakeGriteStore) Enabled() bool { return f.enabled }

func (f *fakeGriteStore) LoadCursor(_ context.Context, _ string) (int64, time.Time, error) {
	return f.cursorTS, f.syncedAt, nil
}

func (f *fakeGriteStore) ListIssues(_ context.Context, _ string) ([]CachedIssue, error) {
	out := make([]CachedIssue, 0, len(f.issues))
	for _, ci := range f.issues {
		out = append(out, ci)
	}
	return out, nil
}

func (f *fakeGriteStore) GetIssue(_ context.Context, _, id string) (*CachedIssue, error) {
	if ci, ok := f.issues[id]; ok {
		return &ci, nil
	}
	return nil, nil
}

func (f *fakeGriteStore) UpsertSync(_ context.Context, _ string, issues []CachedIssue, lastEventTS int64) error {
	for _, ci := range issues {
		f.issues[ci.IssueID] = ci
	}
	f.cursorTS = lastEventTS
	f.syncedAt = time.Now()
	f.upserts++
	return nil
}

// recordingRunner stands in for the grite CLI: it records the args of every call
// and, for `export`, writes the next canned export file to disk and returns the
// envelope pointing at it (mirroring grite's real file-based export).
type recordingRunner struct {
	t       *testing.T
	tempDir string
	exports []griteExportFile
	next    int
	calls   [][]string
}

func (r *recordingRunner) run(_ context.Context, _, _ string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, append([]string(nil), args...))
	if len(args) > 0 && args[0] == "export" {
		idx := r.next
		if idx >= len(r.exports) {
			idx = len(r.exports) - 1
		}
		r.next++
		path := filepath.Join(r.tempDir, fmt.Sprintf("export-%d.json", r.next))
		data, err := json.Marshal(r.exports[idx])
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, err
		}
		env, _ := json.Marshal(map[string]any{"ok": true, "data": map[string]string{"output_path": path}})
		return env, nil
	}
	env, _ := json.Marshal(map[string]any{"ok": true, "data": map[string]any{}})
	return env, nil
}

func (r *recordingRunner) exportCount() int {
	n := 0
	for _, c := range r.calls {
		if len(c) > 0 && c[0] == "export" {
			n++
		}
	}
	return n
}

func (r *recordingRunner) sawArgs(sub string) bool {
	for _, c := range r.calls {
		if strings.Contains(strings.Join(c, " "), sub) {
			return true
		}
	}
	return false
}

func evt(id, issueID, kind string, ts int64, body string) griteEvent {
	payload, _ := json.Marshal(griteEventPayload{Body: body})
	return griteEvent{EventID: id, IssueID: issueID, TimestampMS: ts, Kind: map[string]json.RawMessage{kind: payload}}
}

func newCachedProvider(store GriteCacheStore, runner *recordingRunner, ttl time.Duration) *CachedGriteProvider {
	inner := &GriteProvider{WorkDir: "/repo", Binary: "grite", Runner: runner.run}
	return NewCachedGriteProvider(inner, store, "/repo", ttl)
}

func TestCachedGriteProviderSyncsThenServesFromStore(t *testing.T) {
	store := newFakeGriteStore()
	runner := &recordingRunner{t: t, tempDir: t.TempDir(), exports: []griteExportFile{{
		Issues: []griteIssue{{IssueID: "abc12345", Title: "Do thing", State: "open", Labels: []string{"priority:high", "status:pending"}, UpdatedTS: 100}},
		Events: []griteEvent{evt("e1", "abc12345", "IssueCreated", 100, "the body")},
	}}}
	p := newCachedProvider(store, runner, time.Hour)

	got, err := p.List(context.Background(), DiscoveryFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(got))
	}
	if got[0].ID != "abc12345" || got[0].Status != types.StatusPending || got[0].Priority != types.PriorityHigh {
		t.Fatalf("unexpected mapping: %+v", got[0])
	}
	if runner.exportCount() != 1 || store.upserts != 1 {
		t.Fatalf("expected one export+upsert, got export=%d upserts=%d", runner.exportCount(), store.upserts)
	}
	if store.cursorTS != 100 {
		t.Fatalf("expected cursor advanced to 100, got %d", store.cursorTS)
	}

	// Within the TTL the second read must not re-export.
	if _, err := p.List(context.Background(), DiscoveryFilters{}); err != nil {
		t.Fatalf("second List failed: %v", err)
	}
	if runner.exportCount() != 1 {
		t.Fatalf("expected no re-sync within TTL, got %d exports", runner.exportCount())
	}
}

func TestCachedGriteProviderResyncsAfterTTL(t *testing.T) {
	store := newFakeGriteStore()
	runner := &recordingRunner{t: t, tempDir: t.TempDir(), exports: []griteExportFile{{
		Issues: []griteIssue{{IssueID: "abc12345", Title: "Do thing", State: "open", UpdatedTS: 100}},
		Events: []griteEvent{evt("e1", "abc12345", "IssueCreated", 100, "b")},
	}}}
	p := newCachedProvider(store, runner, time.Nanosecond)

	for i := 0; i < 2; i++ {
		if _, err := p.List(context.Background(), DiscoveryFilters{}); err != nil {
			t.Fatalf("List %d failed: %v", i, err)
		}
	}
	if runner.exportCount() != 2 {
		t.Fatalf("expected re-sync after TTL expiry, got %d exports", runner.exportCount())
	}
}

func TestCachedGriteProviderWritePassThroughForcesSync(t *testing.T) {
	store := newFakeGriteStore()
	runner := &recordingRunner{t: t, tempDir: t.TempDir(), exports: []griteExportFile{{
		Issues: []griteIssue{{IssueID: "abc12345", Title: "Do thing", State: "open", UpdatedTS: 100}},
		Events: []griteEvent{evt("e1", "abc12345", "IssueCreated", 100, "b")},
	}}}
	p := newCachedProvider(store, runner, time.Hour) // long TTL: writes must sync anyway

	if _, err := p.List(context.Background(), DiscoveryFilters{}); err != nil {
		t.Fatalf("warm List failed: %v", err)
	}
	beforeExports := runner.exportCount()

	todo := &types.TODO{ID: "abc12345", Provider: ProviderGrite, ProviderState: "open"}
	if err := p.Delete(context.Background(), todo); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !runner.sawArgs("issue close abc12345") {
		t.Fatalf("expected delete to pass through to grite CLI, calls=%v", runner.calls)
	}
	if runner.exportCount() != beforeExports+1 {
		t.Fatalf("expected write to force a sync within TTL, exports %d -> %d", beforeExports, runner.exportCount())
	}
}

func TestCachedGriteProviderFiltersClosed(t *testing.T) {
	store := newFakeGriteStore()
	runner := &recordingRunner{t: t, tempDir: t.TempDir(), exports: []griteExportFile{{
		Issues: []griteIssue{{IssueID: "done0001", Title: "Closed", State: "closed", UpdatedTS: 50}},
		Events: []griteEvent{evt("e1", "done0001", "IssueCreated", 50, "b")},
	}}}
	p := newCachedProvider(store, runner, time.Hour)

	pending, err := p.List(context.Background(), DiscoveryFilters{IncludeStatuses: []types.Status{types.StatusPending}})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected closed issue excluded from pending filter, got %d", len(pending))
	}
	all, err := p.List(context.Background(), DiscoveryFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 1 || all[0].Status != types.StatusCompleted {
		t.Fatalf("expected closed issue mapped to completed, got %+v", all)
	}
}

func TestCachedGriteProviderGetRebuildsEvents(t *testing.T) {
	store := newFakeGriteStore()
	runner := &recordingRunner{t: t, tempDir: t.TempDir(), exports: []griteExportFile{{
		Issues: []griteIssue{{IssueID: "abc12345", Title: "Do thing", State: "open", UpdatedTS: 100}},
		Events: []griteEvent{evt("e1", "abc12345", "IssueCreated", 100, "the body")},
	}}}
	p := newCachedProvider(store, runner, time.Hour)

	todo, err := p.Get(context.Background(), "abc12345")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if todo.ID != "abc12345" {
		t.Fatalf("unexpected id: %s", todo.ID)
	}
	if len(todo.ProviderEvents) != 1 || todo.ProviderEvents[0].Body != "the body" || todo.ProviderEvents[0].Kind != "IssueCreated" {
		t.Fatalf("expected rebuilt event history, got %+v", todo.ProviderEvents)
	}
}

func TestCachedGriteProviderGetResolvesShortID(t *testing.T) {
	const fullID = "81b6f0520b0f9f1780bc964a172283ee"
	store := newFakeGriteStore()
	runner := &recordingRunner{t: t, tempDir: t.TempDir(), exports: []griteExportFile{{
		Issues: []griteIssue{{IssueID: fullID, Title: "Do thing", State: "open", UpdatedTS: 100}},
		Events: []griteEvent{evt("e1", fullID, "IssueCreated", 100, "the body")},
	}}}
	p := newCachedProvider(store, runner, time.Hour)

	// The 8-char DisplayID shown by `todo list` must resolve to the full issue.
	todo, err := p.Get(context.Background(), shortGriteID(fullID))
	if err != nil {
		t.Fatalf("Get by short id failed: %v", err)
	}
	if todo.ID != fullID {
		t.Fatalf("expected short id resolved to %s, got %s", fullID, todo.ID)
	}
	if len(todo.ProviderEvents) != 1 || todo.ProviderEvents[0].Body != "the body" {
		t.Fatalf("expected rebuilt event history, got %+v", todo.ProviderEvents)
	}
}

func TestCachedGriteProviderGetAmbiguousShortIDErrors(t *testing.T) {
	store := newFakeGriteStore()
	runner := &recordingRunner{t: t, tempDir: t.TempDir(), exports: []griteExportFile{{
		Issues: []griteIssue{
			{IssueID: "abcd111100000000", Title: "First", State: "open", UpdatedTS: 100},
			{IssueID: "abcd222200000000", Title: "Second", State: "open", UpdatedTS: 100},
		},
		Events: []griteEvent{
			evt("e1", "abcd111100000000", "IssueCreated", 100, "a"),
			evt("e2", "abcd222200000000", "IssueCreated", 100, "b"),
		},
	}}}
	p := newCachedProvider(store, runner, time.Hour)

	_, err := p.Get(context.Background(), "abcd")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguous short id error, got %v", err)
	}
}

func TestCachedGriteProviderGetUnknownRefErrors(t *testing.T) {
	store := newFakeGriteStore()
	runner := &recordingRunner{t: t, tempDir: t.TempDir(), exports: []griteExportFile{{
		Issues: []griteIssue{{IssueID: "abc12345abc12345", Title: "Do thing", State: "open", UpdatedTS: 100}},
		Events: []griteEvent{evt("e1", "abc12345abc12345", "IssueCreated", 100, "b")},
	}}}
	p := newCachedProvider(store, runner, time.Hour)

	_, err := p.Get(context.Background(), "ffffffff")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestMergeExportAccumulatesEvents(t *testing.T) {
	priorEvents, _ := json.Marshal([]griteEvent{evt("e1", "x", "CommentAdded", 100, "a")})
	existing := []CachedIssue{{IssueID: "x", EventsJSON: priorEvents}}
	export := griteExportFile{
		Issues: []griteIssue{{IssueID: "x", UpdatedTS: 200}},
		Events: []griteEvent{
			evt("e1", "x", "CommentAdded", 100, "a"), // duplicate of prior, must dedup
			evt("e2", "x", "CommentAdded", 200, "b"),
		},
	}

	issues, cursor := mergeExport("/repo", 100, existing, export)
	if cursor != 200 {
		t.Fatalf("expected cursor 200, got %d", cursor)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	var merged []griteEvent
	if err := json.Unmarshal(issues[0].EventsJSON, &merged); err != nil {
		t.Fatalf("decode merged events: %v", err)
	}
	if len(merged) != 2 || merged[0].EventID != "e1" || merged[1].EventID != "e2" {
		t.Fatalf("expected deduped, time-ordered [e1,e2], got %+v", merged)
	}
}
