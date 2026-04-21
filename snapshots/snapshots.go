package snapshots

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// Save writes snap to .gavel/sha-<id>.json where <id> is the current HEAD sha,
// optionally suffixed with an uncommitted-state checksum. It also refreshes
// .gavel/last.json and .gavel/<branch>.json pointer files. Returns the absolute
// path to the written snapshot.
func Save(workDir string, snap *testui.Snapshot) (string, error) {
	if snap == nil {
		return "", errors.New("snapshots.Save: snapshot is nil")
	}
	if workDir == "" {
		return "", errors.New("snapshots.Save: workDir is required")
	}

	sha, uncommitted, err := SnapshotID(workDir)
	if err != nil {
		return "", err
	}
	if sha == "" {
		return "", errors.New("snapshots.Save: git HEAD sha unavailable")
	}

	dir := filepath.Join(workDir, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}

	fileName := snapshotFileName(sha, uncommitted)
	snapPath := filepath.Join(dir, fileName)
	if err := writeJSON(snapPath, snap); err != nil {
		return "", err
	}

	pointer := &Pointer{
		Path:        filepath.Join(Dir, fileName),
		SHA:         sha,
		Uncommitted: uncommitted,
	}
	if err := writePointer(workDir, PointerLast, pointer); err != nil {
		return "", err
	}
	if branch := branchPointerName(workDir); branch != "" {
		if err := writePointer(workDir, branch, pointer); err != nil {
			return "", err
		}
	}
	return snapPath, nil
}

// LoadPointer reads .gavel/<name>.json and returns the pointer. Missing
// pointer (no .gavel/ or no pointer file) returns (nil, nil) — not an error.
func LoadPointer(workDir, name string) (*Pointer, error) {
	path := filepath.Join(workDir, Dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read pointer %s: %w", path, err)
	}
	var p Pointer
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("decode pointer %s: %w", path, err)
	}
	return &p, nil
}

// LoadByPointer reads and decodes the snapshot referenced by p. A missing
// target file is treated as an error since the pointer exists but the
// referenced data is gone — signals a corrupt cache.
func LoadByPointer(workDir string, p *Pointer) (*testui.Snapshot, error) {
	if p == nil {
		return nil, errors.New("snapshots.LoadByPointer: pointer is nil")
	}
	path := p.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(workDir, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot %s: %w", path, err)
	}
	var snap testui.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("decode snapshot %s: %w", path, err)
	}
	return &snap, nil
}

// SnapshotID returns the HEAD sha and, when the worktree is dirty, an 8-char
// hex checksum of the uncommitted state. The checksum hashes the combined
// output of `git diff --cached`, `git diff`, and the contents of every
// untracked file (NUL-separated). This is stable across identical dirty
// states but changes whenever the uncommitted content changes.
func SnapshotID(workDir string) (string, string, error) {
	sha, err := gitHead(workDir)
	if err != nil {
		return "", "", err
	}

	h := sha256.New()
	written := false

	if out, err := gitOutput(workDir, "diff", "--cached"); err == nil && len(out) > 0 {
		h.Write(out)
		written = true
	}
	if out, err := gitOutput(workDir, "diff"); err == nil && len(out) > 0 {
		h.Write(out)
		written = true
	}
	if out, err := gitOutput(workDir, "ls-files", "--others", "--exclude-standard", "-z"); err == nil && len(out) > 0 {
		for _, name := range splitNUL(out) {
			// Skip the snapshot cache itself — its files are created as a
			// side-effect of Save and would otherwise poison the checksum on
			// the next call.
			if strings.HasPrefix(name, Dir+"/") || name == Dir {
				continue
			}
			full := filepath.Join(workDir, name)
			if data, rerr := os.ReadFile(full); rerr == nil {
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

func gitHead(workDir string) (string, error) {
	out, err := gitOutput(workDir, "rev-parse", "HEAD")
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func gitOutput(workDir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
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
	out, err := gitOutput(workDir, "rev-parse", "--abbrev-ref", "HEAD")
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

func writePointer(workDir, name string, p *Pointer) error {
	path := filepath.Join(workDir, Dir, name+".json")
	return writeJSON(path, p)
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
