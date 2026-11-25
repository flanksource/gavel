package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/flanksource/clicky/prompt"
)

// ErrIndexPatchExclusive rejects --index-patch combined with an option that
// chooses its own staging or rewrites history.
var ErrIndexPatchExclusive = errors.New("--index-patch cannot be combined with")

// privateIndexes maps a git root to the private index file an --index-patch
// commit stages into for the length of its run. Commits run in-process (the
// dashboard answers their prompts), so the index is selected per git command
// rather than through the process environment, which concurrent requests share.
var privateIndexes sync.Map

type realIndexKey struct{}

// withRealIndex makes the git helpers on ctx use the repository's own index
// even while an --index-patch commit of the same root is running.
func withRealIndex(ctx context.Context) context.Context {
	return context.WithValue(ctx, realIndexKey{}, true)
}

// gitIndexEnv is the environment of an index-reading git command (or commit
// hook) run in dir: the private index of an --index-patch run whose git root
// holds dir, or nil to inherit the real index.
func gitIndexEnv(dir string) []string {
	var env []string
	privateIndexes.Range(func(key, value any) bool {
		root := key.(string)
		if dir == root || strings.HasPrefix(dir, root+string(filepath.Separator)) {
			env = append(os.Environ(), "GIT_INDEX_FILE="+value.(string))
			return false
		}
		return true
	})
	return env
}

func validateIndexPatchOptions(opts Options) error {
	if opts.IndexPatch == "" {
		return nil
	}
	switch {
	case len(opts.Files) > 0:
		return fmt.Errorf("%w file arguments", ErrIndexPatchExclusive)
	case opts.Interactive || opts.Batch:
		return fmt.Errorf("%w --interactive", ErrIndexPatchExclusive)
	case opts.Fixup != "":
		return fmt.Errorf("%w --fixup", ErrIndexPatchExclusive)
	case opts.Since != "":
		return fmt.Errorf("%w --since", ErrIndexPatchExclusive)
	}
	return nil
}

// IndexPatchFile is where a commit queue entry id keeps its staged snapshot:
// under the worktree's own git directory, so it never shows as a change.
func IndexPatchFile(workDir, id string) (string, error) {
	if id == "" || strings.ContainsAny(id, `/\`) || id == "." || id == ".." {
		return "", fmt.Errorf("invalid index patch id %q", id)
	}
	out, err := realGitOutput(workDir, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return "", err
	}
	return filepath.Join(strings.TrimSpace(out), "gavel", "commit-queue", id+".patch"), nil
}

// SnapshotOptions name the staged paths to snapshot and the patch file to
// write them to.
type SnapshotOptions struct {
	WorkDir   string
	Paths     []string
	PatchFile string
}

// SnapshotStagedPatch writes the staged changes of opts.Paths (each must be
// staged) to opts.PatchFile and then removes them from the real index, so a
// later --index-patch commit lands exactly that staging however the index
// changes meanwhile.
func SnapshotStagedPatch(ctx context.Context, opts SnapshotOptions) error {
	if len(opts.Paths) == 0 {
		return errors.New("snapshot staged patch: no paths")
	}
	root, err := resolveGitRoot(opts.WorkDir)
	if err != nil {
		return err
	}
	names, err := realGitOutput(root, append([]string{"--literal-pathspecs", "diff", "--cached", "--name-only", "--no-renames", "--"}, opts.Paths...)...)
	if err != nil {
		return err
	}
	stagedNames := splitLines(names)
	for _, path := range opts.Paths {
		if !slices.ContainsFunc(stagedNames, func(name string) bool { return name == path || strings.HasPrefix(name, path+"/") }) {
			return fmt.Errorf("%q has no staged changes to commit", path)
		}
	}
	patch, err := realGitOutput(root, append([]string{"--literal-pathspecs", "diff", "--cached", "--binary", "--full-index", "--no-renames", "--no-color", "--no-ext-diff", "--"}, opts.Paths...)...)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(opts.PatchFile), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(opts.PatchFile), err)
	}
	if err := os.WriteFile(opts.PatchFile, []byte(patch), 0o644); err != nil {
		return fmt.Errorf("write staged snapshot: %w", err)
	}
	if !prompt.PreferSink(ctx) {
		ctx = withGitLockPolicy(ctx, gitLockPolicy{Mode: CheckModeFail})
	}
	ctx = withRealIndex(ctx)
	if out, err := runGitRetryingLocks(ctx, root, append([]string{"--literal-pathspecs", "reset", "--quiet", "--"}, opts.Paths...)...); err != nil {
		return fmt.Errorf("remove snapshot paths from the index (snapshot kept at %s): %w: %s", opts.PatchFile, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RestoreStagedPatch stages a snapshot back into the real index and deletes
// it: the undo of SnapshotStagedPatch for a commit that was never queued.
func RestoreStagedPatch(ctx context.Context, workDir, patchFile string) error {
	root, err := resolveGitRoot(workDir)
	if err != nil {
		return err
	}
	ctx = withRealIndex(withGitLockPolicy(ctx, gitLockPolicy{Mode: CheckModeFail}))
	if out, err := runGitRetryingLocks(ctx, root, "apply", "--cached", "--binary", patchFile); err != nil {
		return fmt.Errorf("restage snapshot %s: %w: %s", patchFile, err, strings.TrimSpace(string(out)))
	}
	return os.Remove(patchFile)
}

func realGitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

// privateIndex is the index an --index-patch run commits from: HEAD plus the
// snapshot patch, in a file beside the patch.
type privateIndex struct {
	root, file, patch, head string
	paths                   []string
}

func openPrivateIndex(ctx context.Context, root, patch string) (*privateIndex, error) {
	if !filepath.IsAbs(patch) {
		return nil, fmt.Errorf("--index-patch %q must be an absolute path", patch)
	}
	if _, err := os.Stat(patch); err != nil {
		return nil, fmt.Errorf("--index-patch: %w", err)
	}
	head, err := realGitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	index := &privateIndex{root: root, file: patch + ".index", patch: patch, head: strings.TrimSpace(head)}
	if err := os.Remove(index.file); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("remove stale private index: %w", err)
	}
	if _, running := privateIndexes.LoadOrStore(root, index.file); running {
		return nil, fmt.Errorf("another --index-patch commit is running in %s", root)
	}
	if err := index.load(ctx); err != nil {
		return nil, errors.Join(err, index.release())
	}
	return index, nil
}

func (p *privateIndex) load(ctx context.Context) error {
	for _, args := range [][]string{{"read-tree", "HEAD"}, {"apply", "--cached", "--binary", p.patch}} {
		if out, err := runGitRetryingLocks(ctx, p.root, args...); err != nil {
			return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
		}
	}
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--no-renames")
	cmd.Dir, cmd.Env = p.root, gitIndexEnv(p.root)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("list snapshot paths: %w", err)
	}
	p.paths = splitLines(string(out))
	return nil
}

func (p *privateIndex) release() error {
	privateIndexes.Delete(p.root)
	if err := os.Remove(p.file); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove private index: %w", err)
	}
	return nil
}

// close ends the run: the real index takes the new HEAD for every committed
// path, and a fully successful run deletes its patch. A failed run keeps the
// patch so a retry can land it.
func (p *privateIndex) close(ctx context.Context, runErr error) error {
	if err := p.release(); err != nil {
		return errors.Join(runErr, err)
	}
	head, err := realGitOutput(p.root, "rev-parse", "HEAD")
	if err != nil {
		return errors.Join(runErr, err)
	}
	if head = strings.TrimSpace(head); head == p.head {
		return runErr
	}
	committed, err := realGitOutput(p.root, "diff", "--name-only", "--no-renames", p.head, head)
	if err != nil {
		return errors.Join(runErr, err)
	}
	paths := slices.Compact(slices.Sorted(slices.Values(append(splitLines(committed), p.paths...))))
	ctx = withRealIndex(ctx)
	if out, err := runGitRetryingLocks(ctx, p.root, append([]string{"--literal-pathspecs", "reset", "--quiet", "--"}, paths...)...); err != nil {
		return errors.Join(runErr, fmt.Errorf("sync the index to the new HEAD: %w: %s", err, strings.TrimSpace(string(out))))
	}
	if runErr != nil {
		return runErr
	}
	return os.Remove(p.patch)
}
