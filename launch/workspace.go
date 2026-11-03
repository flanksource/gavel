package launch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/merge"
	"github.com/flanksource/gavel/internal/safefs"
	"github.com/flanksource/gavel/resultstore"
	"github.com/google/uuid"
)

type WorkspacePolicy struct {
	Root       string
	AllowLocal bool
	LocalRoots []string
	Context    func(context.Context) dbcontext.Context
}

type PreparedWorkspace struct {
	ID          string    `json:"preparationId"`
	Directory   string    `json:"directory"`
	Commit      string    `json:"commit,omitempty"`
	PatchDigest string    `json:"patchDigest,omitempty"`
	Expires     time.Time `json:"expires"`
	Env         []string  `json:"-"`
	Scope       string    `json:"-"`
	cleanup     func() error
}

type workspaceEntry struct {
	PreparedWorkspace
	owner string
	used  bool
}

type Workspaces struct {
	policy     WorkspacePolicy
	mu         sync.Mutex
	entries    map[string]*workspaceEntry
	stop       chan struct{}
	done       chan struct{}
	cleanupErr error
}

func NewWorkspaces(policy WorkspacePolicy) (*Workspaces, error) {
	root, err := filepath.EvalSymlinks(policy.Root)
	if err != nil {
		return nil, fmt.Errorf("launch storage root: %w", err)
	}
	policy.Root, err = filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	for index, allowed := range policy.LocalRoots {
		canonical, err := filepath.Abs(allowed)
		if err != nil {
			return nil, err
		}
		policy.LocalRoots[index], err = filepath.EvalSymlinks(canonical)
		if err != nil {
			return nil, err
		}
	}
	workspaces := &Workspaces{policy: policy, entries: map[string]*workspaceEntry{}, stop: make(chan struct{}), done: make(chan struct{})}
	go workspaces.collect()
	return workspaces, nil
}

func (w *Workspaces) Scope(request Request, owner string) (string, error) {
	if request.Workspace == nil {
		if request.PreparationID == "" {
			return "", errors.New("workspace or preparationId is required")
		}
		return "preparation:" + request.PreparationID, nil
	}
	setup := request.Workspace
	if setup.Checkout != nil {
		if setup.Checkout.URL != "" {
			return setup.Checkout.URL, nil
		}
		if setup.Checkout.Connection != "" {
			return "connection:" + setup.Checkout.Connection, nil
		}
		if setup.Checkout.Path != "" {
			return w.allowed(setup.Checkout.Path)
		}
	}
	return w.allowed(setup.Cwd)
}

func (w *Workspaces) allowed(path string) (string, error) {
	if !w.policy.AllowLocal {
		return "", errors.New("local workspace access is disabled")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(w.policy.Root, path)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	for _, root := range w.policy.LocalRoots {
		canonical, err := filepath.EvalSymlinks(root)
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(canonical, resolved)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return resolved, nil
		}
	}
	return "", fmt.Errorf("workspace path %q is outside enabled roots", path)
}

func (w *Workspaces) Prepare(ctx context.Context, input shell.Setup, owner string) (PreparedWorkspace, error) {
	w.mu.Lock()
	fault := w.cleanupErr
	w.mu.Unlock()
	if fault != nil {
		return PreparedWorkspace{}, fault
	}
	if err := w.expire(); err != nil {
		return PreparedWorkspace{}, err
	}
	setup := merge.Clone(input, merge.Policy{})
	scope, err := w.Scope(Request{Workspace: &input}, owner)
	if err != nil {
		return PreparedWorkspace{}, err
	}
	if err := w.authorize(&setup); err != nil {
		return PreparedWorkspace{}, err
	}
	if setup.BaseDir != "" {
		return PreparedWorkspace{}, errors.New("workspace.baseDir is managed by the launch service")
	}
	id := uuid.NewString()
	base := filepath.Join(w.policy.Root, ".gavel", "workspaces", id)
	if err := os.MkdirAll(base, 0o700); err != nil {
		return PreparedWorkspace{}, err
	}
	remove := func() error { return safefs.SafeRemove(base, safefs.RemoveContext{AllowedRoot: filepath.Dir(base)}) }
	setup.BaseDir = base
	if setup.Checkout != nil && setup.Checkout.Mode != shell.CheckoutNone {
		if setup.Checkout.Worktree == nil {
			setup.Checkout.Worktree = &shell.Worktree{Mode: shell.WorktreeNew}
		}
		if setup.Checkout.Worktree.Mode == shell.WorktreeNew && setup.Checkout.Worktree.Base == "" {
			setup.Checkout.Worktree.Base = setup.Checkout.Ref
			if setup.Checkout.Worktree.Base == "" {
				setup.Checkout.Worktree.Base = "HEAD"
			}
		}
	}
	setupContext := dbcontext.NewContext(ctx)
	if w.policy.Context != nil {
		setupContext = w.policy.Context(ctx)
	}
	result, err := shell.Prepare(setupContext, &setup)
	if err != nil {
		return PreparedWorkspace{}, errors.Join(err, remove())
	}
	prepared := PreparedWorkspace{ID: id, Directory: result.Cwd, Env: result.Env, Scope: scope, Expires: time.Now().Add(15 * time.Minute), cleanup: func() error {
		if result.Cleanup != nil {
			if err := result.Cleanup(); err != nil {
				return err
			}
		}
		return remove()
	}}
	if err := retainPatch(w.policy.Root, id, input.Patch); err != nil {
		return PreparedWorkspace{}, errors.Join(err, prepared.cleanup())
	}
	if commit, ok := result.Extra["commit"].(string); ok {
		prepared.Commit = commit
	}
	if input.Patch != "" {
		digest := sha256.Sum256([]byte(input.Patch))
		prepared.PatchDigest = hex.EncodeToString(digest[:])
	}
	w.mu.Lock()
	w.entries[id] = &workspaceEntry{PreparedWorkspace: prepared, owner: owner}
	w.mu.Unlock()
	return prepared, nil
}

func (w *Workspaces) authorize(setup *shell.Setup) error {
	paths := []*string{}
	if setup.Cwd != "" {
		paths = append(paths, &setup.Cwd)
	}
	if setup.Checkout == nil || setup.Checkout.Mode == shell.CheckoutNone {
		if setup.Cwd == "" {
			return errors.New("workspace requires cwd or a Git checkout")
		}
	} else {
		checkout := setup.Checkout
		if checkout.URL != "" {
			remote, err := url.Parse(checkout.URL)
			if err != nil || remote.Host == "" || (remote.Scheme != "https" && remote.Scheme != "ssh") {
				return errors.New("remote checkout URL must use https or ssh; use checkout.path for an enabled local repository")
			}
		}
		if checkout.Worktree != nil && (checkout.Worktree.Mode != shell.WorktreeNew || checkout.Worktree.Path != "" || checkout.Worktree.Keep) {
			return errors.New("launch checkout requires an owned new worktree without path or keep overrides")
		}
		if checkout.Path != "" {
			paths = append(paths, &checkout.Path)
		}
		if checkout.Path == "" && checkout.URL == "" && checkout.Connection == "" {
			return errors.New("checkout requires path, url or connection")
		}
		if checkout.Path != "" && checkout.URL != "" {
			return errors.New("checkout path and url are mutually exclusive")
		}
		if checkout.Worktree != nil && checkout.Worktree.Path != "" {
			paths = append(paths, &checkout.Worktree.Path)
		}
	}
	for i := range setup.DotEnv {
		paths = append(paths, &setup.DotEnv[i])
	}
	for _, path := range paths {
		resolved, err := w.allowed(*path)
		if err != nil {
			return err
		}
		*path = resolved
	}
	return nil
}

func (w *Workspaces) Get(id, owner string, claim bool) (PreparedWorkspace, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	entry, ok := w.entries[id]
	if !ok || entry.owner != owner {
		return PreparedWorkspace{}, resultstore.ErrNotFound
	}
	if time.Now().After(entry.Expires) && !entry.used {
		return PreparedWorkspace{}, ErrPreparationExpired
	}
	if claim && entry.used {
		return PreparedWorkspace{}, fmt.Errorf("%w: workspace preparation is already in use", resultstore.ErrConflict)
	}
	if claim {
		entry.used = true
	}
	entry.Expires = time.Now().Add(15 * time.Minute)
	return entry.PreparedWorkspace, nil
}

func (w *Workspaces) Release(id string) error {
	w.mu.Lock()
	entry := w.entries[id]
	delete(w.entries, id)
	w.mu.Unlock()
	if entry == nil {
		return errors.New("workspace preparation not found")
	}
	if entry.cleanup != nil {
		return entry.cleanup()
	}
	return nil
}

func (w *Workspaces) expire() error {
	w.mu.Lock()
	var expired []*workspaceEntry
	for id, entry := range w.entries {
		if !entry.used && time.Now().After(entry.Expires) {
			expired = append(expired, entry)
			delete(w.entries, id)
		}
	}
	w.mu.Unlock()
	var failures []error
	for _, entry := range expired {
		if entry.cleanup != nil {
			failures = append(failures, entry.cleanup())
		}
	}
	return errors.Join(failures...)
}

func (w *Workspaces) Close() error {
	close(w.stop)
	<-w.done
	w.mu.Lock()
	ids := make([]string, 0, len(w.entries))
	for id := range w.entries {
		ids = append(ids, id)
	}
	w.mu.Unlock()
	var failures []error
	for _, id := range ids {
		failures = append(failures, w.Release(id))
	}
	return errors.Join(append(failures, w.cleanupErr)...)
}

func (w *Workspaces) collect() {
	defer close(w.done)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			if err := w.expire(); err != nil {
				w.mu.Lock()
				w.cleanupErr = errors.Join(w.cleanupErr, err)
				w.mu.Unlock()
				logger.Errorf("expire launch preparation: %v", err)
			}
		}
	}
}
