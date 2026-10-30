package resultstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

type fileBackend struct{ dir string }

func OpenSnapshots(dir string) (Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return New(&fileBackend{dir: dir}, fileLock(dir)), nil
}

func fileLock(dir string) func(func() error) error {
	var mu sync.Mutex
	return func(fn func() error) (result error) {
		mu.Lock()
		defer mu.Unlock()
		file, err := os.OpenFile(filepath.Join(dir, ".launch.lock"), os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			return err
		}
		defer file.Close()
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
			return err
		}
		defer func() { result = errors.Join(result, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)) }()
		return fn()
	}
}

func (b *fileBackend) path(id string) (string, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("invalid execution id: %w", err)
	}
	return filepath.Join(b.dir, "run-"+id+".json"), nil
}

func (b *fileBackend) Get(ctx context.Context, id string) (Run, error) {
	if err := ctx.Err(); err != nil {
		return Run{}, err
	}
	path, err := b.path(id)
	if err != nil {
		return Run{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Run{}, ErrNotFound
	}
	if err != nil {
		return Run{}, err
	}
	var run Run
	if err := json.Unmarshal(data, &run); err != nil {
		return Run{}, fmt.Errorf("decode %s: %w", path, err)
	}
	if run.Metadata.ExecutionID != id {
		return Run{}, errors.New("execution identity does not match its stored path")
	}
	return run, nil
}

func (b *fileBackend) Put(ctx context.Context, run Run) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := b.path(run.Metadata.ExecutionID)
	if err != nil {
		return err
	}
	data, err := json.Marshal(run)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(b.dir, ".launch-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	_, writeErr := temporary.Write(data)
	err = errors.Join(writeErr, temporary.Sync(), temporary.Close())
	if err != nil {
		return err
	}
	if err := os.Rename(temporary.Name(), path); err != nil {
		return err
	}
	directory, err := os.Open(b.dir)
	if err != nil {
		return err
	}
	return errors.Join(directory.Sync(), directory.Close())
}

func (b *fileBackend) All(ctx context.Context) ([]Run, error) {
	entries, err := os.ReadDir(b.dir)
	if err != nil {
		return nil, err
	}
	var runs []Run
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "run-") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), "run-"), ".json")
		if _, err := uuid.Parse(id); err != nil {
			continue
		}
		run, err := b.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

func (*fileBackend) Close() error { return nil }
