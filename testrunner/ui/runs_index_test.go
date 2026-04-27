package testui_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func writeTestSnapshot(t *testing.T, path string, snap testui.Snapshot, mtime time.Time) {
	t.Helper()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}

func writeTestPointer(t *testing.T, path string, pointerPath, sha string) {
	t.Helper()
	body := map[string]string{"path": pointerPath, "sha": sha}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write pointer %s: %v", path, err)
	}
}

func TestRunsIndexEndpointListsSnapshotsAndPointers(t *testing.T) {
	root := t.TempDir()
	gavelDir := filepath.Join(root, ".gavel")
	if err := os.MkdirAll(gavelDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-30 * time.Minute)

	snapOld := testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: older, Ended: older.Add(time.Minute)},
		Git:      &testui.SnapshotGit{SHA: "abcdef1234567890"},
		Tests: []parsers.Test{
			{Name: "TestA", Passed: true},
			{Name: "TestB", Failed: true},
		},
	}
	snapNew := testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: newer, Ended: newer.Add(2 * time.Minute)},
		Git:      &testui.SnapshotGit{SHA: "1111111deadbeef"},
		Tests: []parsers.Test{
			{Name: "TestC", Passed: true},
			{Name: "TestD", Skipped: true},
			{Name: "TestE", Passed: true},
		},
	}

	oldPath := filepath.Join(gavelDir, "sha-abcdef1234567890.json")
	newPath := filepath.Join(gavelDir, "sha-1111111deadbeef.json")
	writeTestSnapshot(t, oldPath, snapOld, older)
	writeTestSnapshot(t, newPath, snapNew, newer)
	// Pointer that resolves to the new snapshot.
	writeTestPointer(t, filepath.Join(gavelDir, "last.json"),
		filepath.Join(".gavel", "sha-1111111deadbeef.json"), "1111111deadbeef")

	srv := testui.NewServer()
	srv.SetGavelDir(root)
	handler := srv.Handler()

	resp := doRequest(t, handler, http.MethodGet, "/api/runs", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", resp.Code, resp.Body.String())
	}

	var entries []testui.RunIndexEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3: %+v", len(entries), entries)
	}

	// Pointer first.
	if entries[0].Pointer != "last" {
		t.Fatalf("entries[0].Pointer = %q, want last", entries[0].Pointer)
	}
	if entries[0].SHA != "1111111" {
		t.Fatalf("entries[0].SHA = %q, want 1111111", entries[0].SHA)
	}
	if entries[0].Counts == nil || entries[0].Counts.Total != 3 {
		t.Fatalf("entries[0].Counts = %+v, want Total=3", entries[0].Counts)
	}

	// Snapshots after, newest first by mtime.
	if entries[1].Pointer != "" {
		t.Fatalf("entries[1].Pointer = %q, want empty", entries[1].Pointer)
	}
	if entries[1].Name != "sha-1111111deadbeef" {
		t.Fatalf("entries[1].Name = %q, want sha-1111111deadbeef", entries[1].Name)
	}
	if entries[2].Name != "sha-abcdef1234567890" {
		t.Fatalf("entries[2].Name = %q, want sha-abcdef1234567890", entries[2].Name)
	}
	if entries[2].Counts == nil || entries[2].Counts.Failed != 1 || entries[2].Counts.Passed != 1 {
		t.Fatalf("entries[2].Counts = %+v, want Failed=1 Passed=1", entries[2].Counts)
	}
}

func TestRunsIndexEndpoint404WhenDirNotSet(t *testing.T) {
	srv := testui.NewServer()
	handler := srv.Handler()
	resp := doRequest(t, handler, http.MethodGet, "/api/runs", nil)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.Code)
	}
}

func TestRunSnapshotEndpointResolvesPointer(t *testing.T) {
	root := t.TempDir()
	gavelDir := filepath.Join(root, ".gavel")
	if err := os.MkdirAll(gavelDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	snap := testui.Snapshot{
		Tests: []parsers.Test{{Name: "TestX", Passed: true}},
	}
	target := filepath.Join(gavelDir, "sha-deadbeef.json")
	writeTestSnapshot(t, target, snap, time.Now())
	writeTestPointer(t, filepath.Join(gavelDir, "last.json"),
		filepath.Join(".gavel", "sha-deadbeef.json"), "deadbeef")

	srv := testui.NewServer()
	srv.SetGavelDir(root)
	handler := srv.Handler()

	resp := doRequest(t, handler, http.MethodGet, "/api/runs/last.json", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", resp.Code, resp.Body.String())
	}
	var got testui.Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Tests) != 1 || got.Tests[0].Name != "TestX" {
		t.Fatalf("got tests %+v, want [TestX]", got.Tests)
	}
}

func TestRunSnapshotEndpointRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	gavelDir := filepath.Join(root, ".gavel")
	if err := os.MkdirAll(gavelDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	srv := testui.NewServer()
	srv.SetGavelDir(root)
	handler := srv.Handler()

	for _, name := range []string{"..", "../etc/passwd", "foo/bar.json"} {
		resp := doRequest(t, handler, http.MethodGet, "/api/runs/"+name, nil)
		if resp.Code == http.StatusOK {
			t.Fatalf("traversal %q returned 200; body=%s", name, resp.Body.String())
		}
	}
}
