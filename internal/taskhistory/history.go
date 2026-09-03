package taskhistory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	clickytask "github.com/flanksource/clicky/task"
)

const Retention = 30 * 24 * time.Hour

type Record struct {
	Run        clickytask.RunMeta        `json:"run"`
	Snapshots  []clickytask.TaskSnapshot `json:"snapshots"`
	ArchivedAt time.Time                 `json:"archivedAt"`
}

func Archive(root, runID string) error {
	for _, run := range clickytask.Runs(clickytask.RunFilter{}) {
		if run.ID != runID {
			continue
		}
		return Write(root, Record{
			Run:        run,
			Snapshots:  clickytask.SnapshotByID(runID),
			ArchivedAt: time.Now().UTC(),
		})
	}
	return fmt.Errorf("archive task run %q: run not found", runID)
}

func Write(root string, record Record) error {
	if err := validateRecord(record); err != nil {
		return err
	}
	dir := spoolDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create task history spool %s: %w", dir, err)
	}
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal task history %q: %w", record.Run.ID, err)
	}
	path := filepath.Join(dir, record.Run.ID+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write task history %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("publish task history %s: %w", path, err)
	}
	if err := prune(dir, time.Now().UTC()); err != nil {
		return fmt.Errorf("prune task history %s: %w", dir, err)
	}
	return nil
}

func validateRecord(record Record) error {
	if err := validateRunID(record.Run.ID); err != nil {
		return err
	}
	if record.ArchivedAt.IsZero() {
		return errors.New("task history record requires archivedAt")
	}
	if len(record.Snapshots) == 0 {
		return fmt.Errorf("task history record %q has no snapshots", record.Run.ID)
	}
	if _, err := time.Parse(time.RFC3339Nano, record.Run.StartedAt); err != nil {
		return fmt.Errorf("task history record %q has invalid startedAt: %w", record.Run.ID, err)
	}
	return nil
}

// LoadOptions controls which spool files LoadRoots parses.
type LoadOptions struct {
	// Now bounds retention: records archived more than Retention ago are dropped.
	Now time.Time
	// Skip, when set, is consulted for every spool file before it is read. A
	// sweep that already folded a file into the database uses it to avoid
	// re-parsing and re-offering content that cannot have changed.
	Skip func(path string, modTime time.Time) bool
}

func LoadRoots(roots []string, opts LoadOptions) ([]Record, error) {
	if opts.Now.IsZero() {
		return nil, errors.New("task history load requires a reference time")
	}
	seen := map[string]struct{}{}
	var records []Record
	for _, root := range roots {
		if strings.TrimSpace(root) == "" {
			continue
		}
		loaded, err := loadDir(spoolDir(root), opts)
		if err != nil {
			return nil, err
		}
		for _, record := range loaded {
			if _, ok := seen[record.Run.ID]; ok {
				continue
			}
			seen[record.Run.ID] = struct{}{}
			records = append(records, record)
		}
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Run.StartedAt > records[j].Run.StartedAt
	})
	return records, nil
}

func loadDir(dir string, opts LoadOptions) ([]Record, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read task history spool %s: %w", dir, err)
	}
	var records []Record
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if opts.Skip != nil {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("stat task history %s: %w", path, err)
			}
			if opts.Skip(path, info.ModTime()) {
				continue
			}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read task history %s: %w", path, err)
		}
		var record Record
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("parse task history %s: %w", path, err)
		}
		if err := validateRunID(record.Run.ID); err != nil {
			return nil, fmt.Errorf("parse task history %s: %w", path, err)
		}
		if record.ArchivedAt.Add(Retention).After(opts.Now) {
			records = append(records, record)
		}
	}
	return records, nil
}

func prune(dir string, now time.Time) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var record Record
		if err := json.Unmarshal(data, &record); err != nil {
			return err
		}
		if !record.ArchivedAt.Add(Retention).After(now) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRunID(id string) error {
	if id == "" || id == "." || id == ".." || filepath.Base(id) != id || strings.ContainsAny(id, `/\\`) {
		return fmt.Errorf("invalid task run id %q", id)
	}
	return nil
}

func spoolDir(root string) string {
	return filepath.Join(root, ".gavel", "tasks")
}
