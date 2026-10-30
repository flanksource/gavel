// Package workspace resolves, once per invocation, every directory gavel
// reads or writes: the invocation dir, the git root, the dir holding the
// repo-local .gavel state, and the user-level gavel dirs. The result travels in
// the context so every command and library agrees on the same locations.
package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/utils"
)

// StateDirName is the repo-local state directory: snapshots, run history,
// traces, task history, process state and installed lint tools.
const StateDirName = ".gavel"

// ErrNoDirs is returned by From when the context was never given Dirs.
var ErrNoDirs = errors.New("workspace: context carries no resolved directories")

type Dirs struct {
	// Invocation is the absolute directory the command runs in: --cwd, a
	// command's own --work-dir, or the process working directory.
	Invocation string
	// GitRoot is the root of the working tree holding Invocation; "" outside one.
	GitRoot string
	// StateRoot is the directory that holds .gavel: the nearest one at or above
	// Invocation within GitRoot, else GitRoot, else Invocation.
	StateRoot string
	User      UserDirs
}

type UserDirs struct {
	Home   string // ~
	Gavel  string // ~/.gavel
	Config string // ~/.config/gavel
	Cache  string // <user cache dir>/gavel
}

// Resolve resolves every directory for an invocation in dir.
func Resolve(dir string) (Dirs, error) {
	if dir == "" {
		return Dirs{}, errors.New("workspace.Resolve: dir is required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Dirs{}, fmt.Errorf("resolve %s: %w", dir, err)
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return Dirs{}, fmt.Errorf("working directory %s: %w", abs, err)
	}
	if err := root.Close(); err != nil {
		return Dirs{}, err
	}
	user, err := resolveUserDirs()
	if err != nil {
		return Dirs{}, err
	}
	gitRoot := utils.FindGitRoot(abs)
	stateRoot, err := findStateRoot(abs, gitRoot)
	if err != nil {
		return Dirs{}, err
	}
	return Dirs{Invocation: abs, GitRoot: gitRoot, StateRoot: stateRoot, User: user}, nil
}

// findStateRoot walks up from dir to the first directory holding a .gavel
// directory, never past gitRoot, so a user-level ~/.gavel above the repo is
// never taken for repo state.
func findStateRoot(dir, gitRoot string) (string, error) {
	if gitRoot == "" {
		return dir, nil
	}
	root, err := os.OpenRoot(gitRoot)
	if err != nil {
		return "", err
	}
	defer root.Close()
	for current := dir; ; current = filepath.Dir(current) {
		rel, err := filepath.Rel(gitRoot, filepath.Join(current, StateDirName))
		if err != nil {
			return "", err
		}
		info, err := root.Stat(rel)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("workspace state %s: %w", rel, err)
		}
		if err == nil && info.IsDir() {
			return current, nil
		}
		if current == gitRoot || filepath.Dir(current) == current {
			return gitRoot, nil
		}
	}
}

func resolveUserDirs() (UserDirs, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return UserDirs{}, fmt.Errorf("resolve home directory: %w", err)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return UserDirs{}, fmt.Errorf("resolve user cache directory: %w", err)
	}
	return UserDirs{
		Home:   home,
		Gavel:  filepath.Join(home, StateDirName),
		Config: filepath.Join(home, ".config", "gavel"),
		Cache:  filepath.Join(cache, "gavel"),
	}, nil
}

// Root is the git root, or the invocation dir outside a working tree.
func (d Dirs) Root() string {
	if d.GitRoot != "" {
		return d.GitRoot
	}
	return d.Invocation
}

// State joins parts under the repo-local .gavel directory.
func (d Dirs) State(parts ...string) string {
	return filepath.Join(append([]string{d.StateRoot, StateDirName}, parts...)...)
}

type contextKey struct{}

func With(ctx context.Context, d Dirs) context.Context {
	return context.WithValue(ctx, contextKey{}, d)
}

// From returns the Dirs the entry point resolved, ErrNoDirs when it resolved none.
func From(ctx context.Context) (Dirs, error) {
	if ctx == nil {
		return Dirs{}, ErrNoDirs
	}
	d, ok := ctx.Value(contextKey{}).(Dirs)
	if !ok {
		return Dirs{}, ErrNoDirs
	}
	return d, nil
}
