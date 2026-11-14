package slots

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/sdk/status"
	"github.com/gofrs/flock"
)

// PoolStatus is one pool as every gavel process on the machine uses it.
type PoolStatus struct {
	// Name is "global" or "project/<id>".
	Name    string `json:"name"`
	Project string `json:"project,omitempty"`
	// Limit is the largest limit a current holder or waiter took the pool
	// with, else the number of slot files it has.
	Limit   int      `json:"limit"`
	Held    []Slot   `json:"held"`
	Waiting []Waiter `json:"waiting"`
}

type Slot struct {
	Index int `json:"index"`
	// Owner is zero for the instant between a holder locking the slot and
	// recording itself.
	Owner Owner `json:"owner"`
}

type Waiter struct {
	PID   int          `json:"pid"`
	Lease status.Lease `json:"lease"`
}

// Inspect reads every pool under cache. A slot counts as held only while its
// lock is: the probe takes a shared lock for an instant, so an Acquire racing
// it may skip that slot for one retry.
func Inspect(cache string) ([]PoolStatus, error) {
	dirs, err := poolDirs(cache)
	if err != nil {
		return nil, err
	}
	statuses, err := status.List(cache)
	if err != nil {
		return nil, err
	}
	pools := map[string]*PoolStatus{}
	for name, dir := range dirs {
		pool, err := inspectPool(name, dir)
		if err != nil {
			return nil, err
		}
		pools[name] = &pool
	}
	for _, s := range statuses {
		for _, lease := range s.Slots {
			if lease.State != status.LeaseWaiting {
				continue
			}
			pool, ok := pools[lease.Pool]
			if !ok {
				pool = &PoolStatus{Name: lease.Pool}
				pools[lease.Pool] = pool
			}
			pool.Waiting = append(pool.Waiting, Waiter{PID: s.PID, Lease: lease})
			pool.Project = cmp.Or(pool.Project, lease.Project)
			pool.Limit = max(pool.Limit, lease.Limit)
		}
	}
	return sortPools(pools), nil
}

func poolDirs(cache string) (map[string]string, error) {
	root := filepath.Join(cache, "slots")
	dirs := map[string]string{}
	if _, err := os.Stat(filepath.Join(root, "global")); err == nil {
		dirs["global"] = filepath.Join(root, "global")
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	projects, err := os.ReadDir(filepath.Join(root, "projects"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	for _, project := range projects {
		if project.IsDir() {
			dirs["project/"+project.Name()] = filepath.Join(root, "projects", project.Name())
		}
	}
	return dirs, nil
}

func inspectPool(name, dir string) (PoolStatus, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return PoolStatus{}, err
	}
	pool := PoolStatus{Name: name}
	heldLimit := 0
	for _, entry := range entries {
		stem, ok := strings.CutSuffix(entry.Name(), ".lock")
		index, err := strconv.Atoi(stem)
		if !ok || err != nil {
			continue
		}
		pool.Limit = max(pool.Limit, index+1)
		path := filepath.Join(dir, entry.Name())
		owner, held, err := probe(path)
		if err != nil {
			return PoolStatus{}, err
		}
		if held {
			pool.Held = append(pool.Held, Slot{Index: index, Owner: owner})
			pool.Project = cmp.Or(pool.Project, owner.Project)
			heldLimit = max(heldLimit, owner.Limit)
		}
	}
	slices.SortFunc(pool.Held, func(a, b Slot) int { return a.Index - b.Index })
	pool.Limit = cmp.Or(heldLimit, pool.Limit)
	return pool, nil
}

// probe reports whether another process holds the slot file, and its owner record.
func probe(path string) (Owner, bool, error) {
	lock := flock.New(path)
	free, err := lock.TryRLock()
	if err != nil {
		return Owner{}, false, errors.Join(fmt.Errorf("probe slot %s: %w", path, err), lock.Close())
	}
	if free {
		return Owner{}, false, lock.Close()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Owner{}, false, errors.Join(err, lock.Close())
	}
	var owner Owner
	if len(data) > 0 {
		if err := json.Unmarshal(data, &owner); err != nil {
			return Owner{}, false, errors.Join(fmt.Errorf("decode slot owner %s: %w", path, err), lock.Close())
		}
	}
	return owner, true, lock.Close()
}

func sortPools(pools map[string]*PoolStatus) []PoolStatus {
	sorted := make([]PoolStatus, 0, len(pools))
	for _, pool := range pools {
		sorted = append(sorted, *pool)
	}
	rank := func(p PoolStatus) int {
		if p.Name == "global" {
			return 0
		}
		return 1
	}
	slices.SortFunc(sorted, func(a, b PoolStatus) int {
		return cmp.Or(cmp.Compare(rank(a), rank(b)), cmp.Compare(a.Project, b.Project), cmp.Compare(a.Name, b.Name))
	})
	return sorted
}
