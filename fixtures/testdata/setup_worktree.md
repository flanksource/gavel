---
setup:
  checkout:
    mode: local
    path: .
    worktree:
      mode: new
      uncommitted: skip
      ignored: skip
---

# setup: worktree

`setup.checkout` with a `worktree:` prepares a disposable tree once for this
document and runs every fixture in it. `uncommitted: skip` and `ignored: skip`
keep this fixture cheap — it only needs to prove the relocation, not the copy.

The source repository is never mutated: nothing is stashed, moved, or restored.

## The prepared tree is where commands run

`$SETUP_DIR` is the prepared directory, and `GIT_ROOT_DIR`/`ROOT_DIR` are
re-rooted onto it rather than staying on the repository this markdown lives in.

### command: git toplevel is the prepared worktree

```bash
test "$(git rev-parse --show-toplevel)" = "$(printenv SETUP_DIR)"
```

### command: GIT_ROOT_DIR follows the worktree

```bash
test "$(printenv GIT_ROOT_DIR)" = "$(printenv SETUP_DIR)"
```

### command: the worktree is not the checked-out repository

The worktree lands under gavel's per-file cache directory, so it is a different
path from the repository holding this file.

```bash
test "$(pwd)" != "$(git -C "$(printenv SETUP_DIR)" rev-parse --git-common-dir)"
test "$(pwd)" = "$(printenv SETUP_DIR)"
git -C "$(printenv SETUP_DIR)" rev-parse --git-common-dir | grep -qv "^$(printenv SETUP_DIR)/.git$"
```

## The setup template variable describes the prepared tree

| Name             | CLI  | Args           | CEL Validation                |
|------------------|------|----------------|-------------------------------|
| commit is pinned | git  | rev-parse HEAD | stdout.trim() == setup.commit |

### command: setup.worktree is the directory the fixtures run in

```bash
test "{{.setup.worktree}}" = "$(printenv SETUP_DIR)"
test -f "{{.setup.worktree}}/go.mod"
```
