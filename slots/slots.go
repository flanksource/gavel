package slots

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
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sdk/status"
	"github.com/flanksource/gavel/workspace"
	"github.com/gofrs/flock"
)

type Request struct {
	Limits   Limits
	WorkDir  string
	CacheDir string
	Command  string
	// Count is the most slots wanted from each pool, 0 meaning one. Acquire
	// grants as many as are free, at least one, rather than waiting for all.
	Count int
}

type Lease struct {
	locks     []*flock.Flock
	tracked   []int
	count     int
	inherited bool
	once      sync.Once
	err       error
}

// Env is the environment a subprocess started under this lease must carry.
func (l *Lease) Env() map[string]string {
	if l == nil || len(l.locks) == 0 && !l.inherited {
		return nil
	}
	return map[string]string{HeldEnv: "1"}
}

// Count is the number of slots held in each pool; zero when no locks were
// taken (slots disabled, or inherited from a parent's lease).
func (l *Lease) Count() int {
	if l == nil {
		return 0
	}
	return l.count
}

// HeldEnv marks a subprocess started under a lease. A gavel nested inside it
// (a fixture command or linter shim running gavel) runs on the parent's slot:
// taking a second one from the same pools deadlocks once every slot is held
// by a parent waiting on its child.
const HeldEnv = "GAVEL_SLOT_HELD"

type pool struct {
	name, dir string
	// project is the canonical git common directory of a project pool.
	project string
	limit   int
}

// Owner is the record a holder writes into the slot file it locked. It
// outlives the lock, so only a held lock's record names a live holder.
type Owner struct {
	PID     int       `json:"pid"`
	Command string    `json:"command"`
	Project string    `json:"project,omitempty"`
	Limit   int       `json:"limit,omitempty"`
	Started time.Time `json:"started,omitzero"`
}

func Acquire(ctx context.Context, request Request) (*Lease, error) {
	if request.Limits.Global < 0 || request.Limits.Project < 0 {
		return nil, fmt.Errorf("slots: negative limits")
	}
	if request.Count < 0 {
		return nil, fmt.Errorf("slots: negative count %d", request.Count)
	}
	count := max(request.Count, 1)
	if !request.Limits.Enabled() {
		return &Lease{}, nil
	}
	if os.Getenv(HeldEnv) != "" {
		return &Lease{inherited: true}, nil
	}
	pools, err := request.pools(ctx)
	if err != nil {
		return nil, err
	}
	resume := pauseTimeouts(ctx)
	defer resume()
	logged := map[string]bool{}
	queued := &queue{command: request.Command}
	defer queued.leave()
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		lease, waiting, holders, err := tryPools(pools, request.Command, count)
		if err != nil || waiting == nil {
			return lease, err
		}
		queued.wait(*waiting)
		if !logged[waiting.name] {
			logger.Infof("Waiting for slots pool %s; holders: %s", waiting.name, strings.Join(holders, "; "))
			logged[waiting.name] = true
		}
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (r Request) pools(ctx context.Context) ([]pool, error) {
	if r.WorkDir == "" {
		r.WorkDir = "."
	}
	if r.CacheDir == "" {
		dirs, err := workspace.Resolve(r.WorkDir)
		if err != nil {
			return nil, err
		}
		r.CacheDir = dirs.User.Cache
	}
	var pools []pool
	if r.Limits.Project > 0 {
		project, id, err := ProjectDir(ctx, r.WorkDir)
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool{name: "project/" + id, dir: filepath.Join(r.CacheDir, "slots", "projects", id), project: project, limit: r.Limits.Project})
	}
	if r.Limits.Global > 0 {
		pools = append(pools, pool{name: "global", dir: filepath.Join(r.CacheDir, "slots", "global"), limit: r.Limits.Global})
	}
	for _, p := range pools {
		if err := os.MkdirAll(p.dir, 0o700); err != nil {
			return nil, fmt.Errorf("slots %s: %w", p.name, err)
		}
	}
	return pools, nil
}

// ProjectDir is the canonical git common directory of dir's repository, which
// its linked worktrees share, and the id of its project pool.
func ProjectDir(ctx context.Context, dir string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("slots: git common directory in %s: %w: %s", dir, err, output)
	}
	common, err := filepath.EvalSymlinks(strings.TrimSpace(string(output)))
	if err != nil {
		return "", "", fmt.Errorf("slots: resolve git common directory: %w", err)
	}
	common, err = filepath.Abs(common)
	if err != nil {
		return "", "", err
	}
	hash := sha256.Sum256([]byte(common))
	return common, hex.EncodeToString(hash[:]), nil
}

// Never hold a global slot while waiting for a project's pool, or a busy
// repository could starve every other project. Roll back partial acquisitions.
// Each pool grants up to what the pools before it granted, and earlier pools
// give back their surplus, so every pool ends up holding the same count.
// It returns the pool it must wait for, if any.
func tryPools(pools []pool, command string, count int) (*Lease, *pool, []string, error) {
	taken := make([][]*flock.Flock, 0, len(pools))
	indexes := make([][]int, 0, len(pools))
	for _, p := range pools {
		locks, index, holders, err := p.tryLock(command, count)
		if err != nil || len(locks) == 0 {
			for _, held := range taken {
				err = errors.Join(err, closeLocks(held))
			}
			return nil, &p, holders, err
		}
		taken, indexes = append(taken, locks), append(indexes, index)
		count = len(locks)
	}
	lease := &Lease{count: count}
	var err error
	for i, p := range pools {
		err = errors.Join(err, closeLocks(taken[i][count:]))
		lease.locks = append(lease.locks, taken[i][:count]...)
		for _, index := range indexes[i][:count] {
			lease.tracked = append(lease.tracked, track(p.lease(status.LeaseHeld, index, command)))
		}
	}
	if err != nil {
		return nil, nil, nil, errors.Join(err, lease.Close())
	}
	return lease, nil, nil, nil
}

func closeLocks(locks []*flock.Flock) error {
	var err error
	for _, lock := range locks {
		err = errors.Join(err, lock.Close())
	}
	return err
}

// tryLock takes up to count free slots of p, returning the holders of the
// busy ones it skipped.
func (p pool) tryLock(command string, count int) ([]*flock.Flock, []int, []string, error) {
	var locks []*flock.Flock
	var indexes []int
	var holders []string
	for i := 0; i < p.limit && len(locks) < count; i++ {
		path := filepath.Join(p.dir, fmt.Sprintf("%d.lock", i))
		lock := flock.New(path)
		locked, err := lock.TryLock()
		if err != nil {
			return nil, nil, nil, errors.Join(err, lock.Close(), closeLocks(locks))
		}
		if locked {
			data, err := json.Marshal(Owner{PID: os.Getpid(), Command: command, Project: p.project, Limit: p.limit, Started: time.Now().UTC()})
			if err == nil {
				err = os.WriteFile(path, data, 0o600)
			}
			if err != nil {
				return nil, nil, nil, errors.Join(err, lock.Close(), closeLocks(locks))
			}
			locks, indexes = append(locks, lock), append(indexes, i)
			continue
		}
		data, readErr := os.ReadFile(path)
		if err := errors.Join(readErr, lock.Close()); err != nil {
			return nil, nil, nil, errors.Join(err, closeLocks(locks))
		}
		holders = append(holders, string(data))
	}
	return locks, indexes, holders, nil
}

func (l *Lease) Close() error {
	if l == nil {
		return nil
	}
	l.once.Do(func() {
		for i := len(l.locks) - 1; i >= 0; i-- {
			l.err = errors.Join(l.err, l.locks[i].Close())
		}
		untrack(l.tracked...)
	})
	return l.err
}
