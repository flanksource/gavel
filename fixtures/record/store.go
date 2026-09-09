package record

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Dir is the subdirectory recordings are written to, relative to the fixture
// run's working directory. It is deliberately a level below .gavel so that
// snapshots' `.gavel/run-*.json` glob keeps matching exactly the run snapshots,
// and so the repo's existing `.gitignore` entry for `.gavel/` already covers it.
const Dir = ".gavel/recordings"

// filePrefix marks a recording artifact, mirroring snapshots.PerRunPrefix.
const filePrefix = "rec-"

// timestampLayout is filesystem-safe (no colons) but still chronologically
// sortable. Kept byte-identical to snapshots.PerRunTimestampLayout, which
// package fixtures may not import (see the note on record's package doc).
const timestampLayout = "2006-01-02T15-04-05Z"

// Ext is the file extension for a kind's artifact. HTTP and client recordings
// share `.har` so a browser's devtools opens either one directly.
func (k Kind) Ext() string {
	switch k {
	case KindANSI:
		return "cast.json"
	case KindSQL:
		return "sql.jsonl"
	default:
		return "har"
	}
}

// Format names the on-disk encoding, reported on Result so a viewer knows how
// to render an artifact without sniffing it.
func (k Kind) Format() string {
	switch k {
	case KindANSI:
		return "asciinema-v2"
	case KindSQL:
		return "jsonl"
	default:
		return "har-1.2"
	}
}

// Store allocates artifact paths for one fixture run. All recordings from a run
// share its start timestamp, which is what lets Prune retire whole runs rather
// than individual files.
type Store struct {
	dir     string
	started time.Time

	mu sync.Mutex
}

// NewStore returns a Store writing under workDir/.gavel/recordings. The
// directory is created lazily on the first Create, so a run whose fixtures
// declare no `record:` leaves no trace.
func NewStore(workDir string, started time.Time) *Store {
	return &Store{dir: filepath.Join(workDir, Dir), started: started.UTC()}
}

// Dir returns the directory artifacts are written to.
func (s *Store) Dir() string { return s.dir }

// Create opens a new artifact file for label (usually the fixture name) and
// returns it alongside the Result describing it. The caller writes the payload,
// closes the file, and fills in the counts.
func (s *Store) Create(label string, kind Kind) (*os.File, Result, error) {
	file, path, err := s.create(label, kind, kind.Ext())
	if err != nil {
		return nil, Result{}, err
	}
	return file, Result{
		Kind:   kind,
		ID:     strings.TrimSuffix(filepath.Base(path), "."+kind.Ext()),
		Path:   path,
		Format: kind.Format(),
	}, nil
}

// CreateSidecar opens a companion file for a recording — today the generated CA
// certificate a mitm proxy's children have to trust. It is named like the
// recordings so Prune retires it with the run instead of leaving it behind, but
// it is not itself a Result: nothing asserts on it.
func (s *Store) CreateSidecar(label string, kind Kind, ext string) (*os.File, string, error) {
	return s.create(label, kind, ext)
}

func (s *Store) create(label string, kind Kind, ext string) (*os.File, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return nil, "", fmt.Errorf("create recording dir %s: %w", s.dir, err)
	}

	stem := fmt.Sprintf("%s%s-%s-%s", filePrefix, s.started.Format(timestampLayout), slug(label), kind)
	path, err := uniquePath(s.dir, stem, ext)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, "", fmt.Errorf("create recording %s: %w", path, err)
	}
	return file, path, nil
}

// uniquePath suffixes -2, -3, … when the name is taken, so two fixtures with
// the same name in one run do not overwrite each other. Mirrors
// snapshots.uniqueRunPath, which lives behind a package fixtures may not
// import.
func uniquePath(dir, stem, ext string) (string, error) {
	for i := 1; i < 1000; i++ {
		name := stem
		if i > 1 {
			name = fmt.Sprintf("%s-%d", stem, i)
		}
		path := filepath.Join(dir, name+"."+ext)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return path, nil
			}
			return "", fmt.Errorf("stat %s: %w", path, err)
		}
	}
	return "", fmt.Errorf("record: no free filename for %s in %s", stem, dir)
}

// slug reduces a fixture name to lowercase alphanumerics and dashes so it is
// filename-safe and still readable in a directory listing.
func slug(label string) string {
	var b strings.Builder
	lastDash := true // suppresses a leading dash
	for _, r := range strings.ToLower(strings.TrimSpace(label)) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// Prune keeps the newest `keep` runs' recordings and deletes the rest. Casts and
// HARs are large and .gavel has no garbage collection of its own, so without
// this a repo accumulates every recording ever made.
//
// Retention is by run, not by file: one run emits several artifacts, and
// keeping a file count would leave a run half-deleted.
func Prune(workDir string, keep int) error {
	if keep < 0 {
		return fmt.Errorf("record.Prune: keep must not be negative, got %d", keep)
	}
	dir := filepath.Join(workDir, Dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read recording dir %s: %w", dir, err)
	}

	byRun := map[string][]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), filePrefix) {
			continue
		}
		byRun[runStamp(entry.Name())] = append(byRun[runStamp(entry.Name())], entry.Name())
	}

	stamps := make([]string, 0, len(byRun))
	for stamp := range byRun {
		stamps = append(stamps, stamp)
	}
	// The layout sorts chronologically as a string, newest last.
	sort.Sort(sort.Reverse(sort.StringSlice(stamps)))

	for _, stamp := range stamps[min(keep, len(stamps)):] {
		for _, name := range byRun[stamp] {
			if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("prune recording %s: %w", name, err)
			}
		}
	}
	return nil
}

// runStamp extracts the timestamp segment shared by every artifact of one run.
// A name that does not carry one groups under "" and is pruned first, which is
// the right fate for a stray file in this directory.
func runStamp(name string) string {
	rest := strings.TrimPrefix(name, filePrefix)
	if len(rest) < len(timestampLayout) {
		return ""
	}
	stamp := rest[:len(timestampLayout)]
	if _, err := time.Parse(timestampLayout, stamp); err != nil {
		return ""
	}
	return stamp
}
