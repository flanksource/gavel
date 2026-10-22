package snapshots

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	rpchttp "github.com/flanksource/clicky/rpc/http"
	"github.com/flanksource/gavel/blob"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

const (
	Dir         = ".gavel"
	PointerLast = "last"
)

type Pointer struct {
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Uncommitted string `json:"uncommitted,omitempty"`
}

// Save writes snap as sha-<id>.json in workDir's snapshot store, where <id> is the current HEAD
// sha, optionally suffixed with an uncommitted-state checksum. It also refreshes the last.json and
// <branch>.json pointers. It returns where the snapshot was written (see Location).
func Save(workDir string, snap *testui.Snapshot) (string, error) {
	if snap == nil {
		return "", errors.New("snapshots.Save: snapshot is nil")
	}
	if workDir == "" {
		return "", errors.New("snapshots.Save: workDir is required")
	}

	sha, uncommitted, err := SnapshotID(context.Background(), workDir)
	if err != nil {
		return "", err
	}
	if sha == "" {
		return "", errors.New("snapshots.Save: git HEAD sha unavailable")
	}

	store, err := OpenStore(workDir)
	if err != nil {
		return "", err
	}
	fileName := snapshotFileName(sha, uncommitted)
	if err := putJSON(store, fileName, snap); err != nil {
		return "", err
	}

	pointer := &Pointer{
		Path:        filepath.Join(Dir, fileName),
		SHA:         sha,
		Uncommitted: uncommitted,
	}
	if err := putJSON(store, PointerLast+".json", pointer); err != nil {
		return "", err
	}
	if branch := branchPointerName(workDir); branch != "" {
		if err := putJSON(store, branch+".json", pointer); err != nil {
			return "", err
		}
	}
	return Location(workDir, fileName), nil
}

// LoadPointer reads the <name>.json pointer from workDir's snapshot store. A missing pointer
// returns (nil, nil) — not an error.
func LoadPointer(workDir, name string) (*Pointer, error) {
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	return loadPointer(store, name)
}

func loadPointer(store blob.Store, name string) (*Pointer, error) {
	data, err := readKey(store, name+".json")
	if errors.Is(err, blob.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read pointer %s.json: %w", name, err)
	}
	var p Pointer
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("decode pointer %s.json: %w", name, err)
	}
	return &p, nil
}

// LoadByPointer reads and decodes the snapshot referenced by p. A missing
// target is an error since the pointer exists but the referenced data is
// gone — signals a corrupt cache.
func LoadByPointer(workDir string, p *Pointer) (*testui.Snapshot, error) {
	if p == nil {
		return nil, errors.New("snapshots.LoadByPointer: pointer is nil")
	}
	key, err := pointerKey(p)
	if err != nil {
		return nil, err
	}
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	data, err := readKey(store, key)
	if err != nil {
		return nil, err
	}
	var snap testui.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("decode snapshot %s: %w", Location(workDir, key), err)
	}
	return &snap, nil
}

// ResolveLast returns a local file holding the snapshot last.json points at. It errors loudly
// when no pointer exists so callers (e.g. `gavel test --failed`) fail fast instead of silently
// widening to a full run. When snapshots live in a bucket, the snapshot is copied to the path
// it would have under workDir's .gavel directory.
func ResolveLast(workDir string) (string, error) {
	if workDir == "" {
		return "", errors.New("snapshots.ResolveLast: workDir is required")
	}
	store, err := OpenStore(workDir)
	if err != nil {
		return "", err
	}
	ptr, err := loadPointer(store, PointerLast)
	if err != nil {
		return "", err
	}
	if ptr == nil {
		return "", fmt.Errorf("no previous run found at %s — run `gavel test` once to create it", Location(workDir, PointerLast+".json"))
	}
	key, err := pointerKey(ptr)
	if err != nil {
		return "", err
	}
	data, err := readKey(store, key)
	if err != nil {
		return "", fmt.Errorf("last-run snapshot missing at %s: %w", Location(workDir, key), err)
	}
	path := filepath.Join(workDir, Dir, key)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("copy the last-run snapshot to %s: %w", path, err)
	}
	return path, nil
}

// SnapshotID returns the HEAD sha and, when the worktree is dirty, an 8-char
// hex checksum of the uncommitted state. The checksum hashes the combined
// output of `git diff --cached`, `git diff`, and the contents of every
// untracked file (NUL-separated). This is stable across identical dirty
// states but changes whenever the uncommitted content changes.
func SnapshotID(ctx context.Context, workDir string) (string, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	sha, err := gitHead(ctx, workDir)
	if err != nil {
		return "", "", err
	}

	h := sha256.New()
	written := false

	if out, err := gitOutput(ctx, workDir, "diff", "--cached"); err == nil && len(out) > 0 {
		h.Write(out)
		written = true
	}
	if out, err := gitOutput(ctx, workDir, "diff"); err == nil && len(out) > 0 {
		h.Write(out)
		written = true
	}
	if out, err := gitOutput(ctx, workDir, "ls-files", "--others", "--exclude-standard", "-z"); err == nil && len(out) > 0 {
		for _, name := range splitNUL(out) {
			// Skip the snapshot cache itself — its files are created as a
			// side-effect of Save and would otherwise poison the checksum on
			// the next call.
			if strings.HasPrefix(name, Dir+"/") || name == Dir {
				continue
			}
			full := filepath.Join(workDir, name)
			stopFile := rpchttp.Track(ctx, "file")
			data, rerr := os.ReadFile(full)
			stopFile()
			if rerr == nil {
				h.Write([]byte(name))
				h.Write([]byte{0})
				h.Write(data)
				written = true
			}
		}
	}

	if !written {
		return sha, "", nil
	}
	return sha, hex.EncodeToString(h.Sum(nil))[:8], nil
}

func gitHead(ctx context.Context, workDir string) (string, error) {
	out, err := gitOutput(ctx, workDir, "rev-parse", "HEAD")
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func gitOutput(ctx context.Context, workDir string, args ...string) ([]byte, error) {
	stopGit := rpchttp.Track(ctx, "git")
	defer stopGit()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func splitNUL(raw []byte) []string {
	s := strings.TrimRight(string(raw), "\x00")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x00")
}

func snapshotFileName(sha, uncommitted string) string {
	if uncommitted == "" {
		return fmt.Sprintf("sha-%s.json", sha)
	}
	return fmt.Sprintf("sha-%s-%s.json", sha, uncommitted)
}

func branchPointerName(workDir string) string {
	out, err := gitOutput(context.Background(), workDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	name := strings.TrimSpace(string(out))
	if name == "" || name == "HEAD" {
		return "detached"
	}
	return SanitiseBranch(name)
}

// SanitiseBranch turns a branch name into a safe file stem by replacing any
// path separator with '-'.
func SanitiseBranch(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, string(os.PathSeparator), "-")
	name = strings.Trim(name, "-")
	if name == "" {
		return "detached"
	}
	return name
}

// WriteSnapshot writes snap to an explicit path, without the git-sha naming
// and pointer bookkeeping Save performs. Used for snapshots that are not a
// local run — e.g. the merged CI results `--pr` caches at .gavel/pr-<n>.json.
func WriteSnapshot(path string, snap *testui.Snapshot) error {
	if snap == nil {
		return errors.New("snapshots.WriteSnapshot: snapshot is nil")
	}
	return writeJSON(path, snap)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
