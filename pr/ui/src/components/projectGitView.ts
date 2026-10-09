import { branchRef, worktreeRef } from '../projectRef';
import type { GitWorktree, GitWorktreeChanges, ProjectGit } from '../types';
import { ageShort, timeAgoShort } from '../utils';

export function dirtyFileCount({ staged, unstaged, both, untracked, conflict }: GitWorktreeChanges): number {
  return staged + unstaged + both + untracked + conflict;
}

function basename(path: string): string {
  return path.split('/').filter(Boolean).pop() ?? path;
}

/**
 * What an entry holds beyond its base: commits ahead, uncommitted files, and
 * the line counts — of the uncommitted files for a checkout, of the commits
 * for a branch. lines is absent when there is nothing to count.
 */
export interface ProjectRefStats {
  commits: number;
  uncommitted: number;
  lines?: { adds: number; dels: number };
}

function worktreeStats(changes: GitWorktreeChanges, commits: number): ProjectRefStats {
  const uncommitted = dirtyFileCount(changes);
  return uncommitted > 0 ? { commits, uncommitted, lines: { adds: changes.adds, dels: changes.dels } } : { commits, uncommitted };
}

/**
 * One entry of the worktree/branch picker. value is the raw `?ref=` ("" for
 * the main checkout); label is the searchable name.
 */
export interface ProjectRefEntry {
  value: string;
  group: 'Checkout' | 'Worktrees' | 'Branches';
  label: string;
  title: string;
  stats: ProjectRefStats;
  lastCommitAt: string;
  touchedAt?: string;
}

function worktreeEntry(worktree: GitWorktree): ProjectRefEntry {
  return {
    value: worktreeRef(worktree.path),
    group: 'Worktrees',
    // A detached worktree has no branch name to tell it apart, so its directory stands in.
    label: worktree.branch || `(detached) ${basename(worktree.path)}`,
    title: worktree.path,
    stats: worktreeStats(worktree.changes, worktree.ahead),
    lastCommitAt: worktree.lastCommitAt,
    touchedAt: worktree.touchedAt,
  };
}

/**
 * Entries of the worktree/branch picker: the main checkout, every linked
 * worktree, then each local branch that has commits past the base and is not
 * checked out anywhere.
 */
export function projectRefEntries(git: ProjectGit): ProjectRefEntry[] {
  const main = git.worktrees.find(worktree => worktree.primary);
  return [
    {
      value: '',
      group: 'Checkout',
      label: git.currentBranch || '(detached)',
      title: main?.path ?? '',
      stats: main ? worktreeStats(main.changes, 0) : { commits: 0, uncommitted: 0 },
      lastCommitAt: main?.lastCommitAt ?? '',
      touchedAt: main?.touchedAt,
    },
    ...git.worktrees.filter(worktree => !worktree.primary).map(worktreeEntry),
    ...git.branches.filter(branch => branch.worktree === '').map((branch): ProjectRefEntry => ({
      value: branchRef(branch.name),
      group: 'Branches',
      label: branch.name,
      title: branch.name,
      stats: { commits: branch.ahead, uncommitted: 0, lines: { adds: branch.diff.adds, dels: branch.diff.dels } },
      lastCommitAt: branch.lastCommitAt,
    })),
  ];
}

// Go marshals an unset time.Time as year 1; any date at or before the epoch
// is "no date", not a 56-year-old commit.
const age = (iso: string | undefined) => (iso && Date.parse(iso) > 0 ? ageShort(iso) : '');

/** Compact ages of an entry's last commit and newest uncommitted edit. */
export function projectRefAges({ lastCommitAt, touchedAt }: { lastCommitAt: string; touchedAt?: string }): { committed: string; touched: string } {
  return { committed: age(lastCommitAt), touched: age(touchedAt) };
}

/** How old the git state a pane shows is, e.g. "git state 8s ago". */
export function gitStateAge(scannedAt: string): string {
  return `git state ${timeAgoShort(scannedAt)}`;
}

/**
 * When what a pane shows was last scanned: the repo's refs (computedAt) and,
 * for the main checkout or a worktree, the older of that and its own status
 * scan. Other worktrees, polled less often, do not age it.
 */
export function gitStateScannedAt(computedAt: string, worktree?: { statusScannedAt?: string }): string {
  const status = worktree?.statusScannedAt;
  return status && Date.parse(status) < Date.parse(computedAt) ? status : computedAt;
}
