import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { GitBranchInfo, GitWorktree, ProjectGit } from '../types';
import { dirtyFileCount, gitStateAge, projectRefAges, projectRefEntries } from './projectGitView';

const NOW = new Date('2026-10-04T12:00:00Z');
const HOURS_AGO_3 = '2026-10-04T09:00:00Z';
const MINUTES_AGO_5 = '2026-10-04T11:55:00Z';
const DAYS_AGO_2 = '2026-10-02T12:00:00Z';
const GO_ZERO_TIME = '0001-01-01T00:00:00Z';

const primary: GitWorktree = {
  path: '/work/gavel', branch: 'main', head: 'a1', primary: true, detached: false, prunable: false,
  changes: { staged: 0, unstaged: 4, both: 0, untracked: 0, conflict: 0, adds: 9, dels: 9 }, ahead: 0,
  lastCommitAt: DAYS_AGO_2, touchedAt: MINUTES_AGO_5,
};

const linked: GitWorktree = {
  path: '/work/gavel-feat', branch: 'feat/x', head: 'b2', primary: false, detached: false, prunable: false,
  changes: { staged: 1, unstaged: 2, both: 1, untracked: 3, conflict: 1, adds: 10, dels: 4 }, ahead: 2,
  lastCommitAt: HOURS_AGO_3, touchedAt: MINUTES_AGO_5,
};

const clean: GitWorktree = {
  path: '/work/gavel-clean', branch: 'chore/y', head: 'c3', primary: false, detached: false, prunable: false,
  changes: { staged: 0, unstaged: 0, both: 0, untracked: 0, conflict: 0, adds: 0, dels: 0 }, ahead: 1,
  lastCommitAt: DAYS_AGO_2,
};

const detached: GitWorktree = { ...clean, path: '/work/gavel-det', branch: '', head: 'd4', detached: true, ahead: 0 };

const branch = (name: string, worktree: string, ahead: number, adds: number, dels: number): GitBranchInfo => ({
  name, head: 'e5', ahead, behind: 0, worktree, diff: { commits: ahead, files: 1, adds, dels }, lastCommitAt: HOURS_AGO_3,
});

const git: ProjectGit = {
  base: 'main',
  currentBranch: 'main',
  baseCheckedOut: true,
  worktrees: [primary, linked, clean, detached],
  branches: [branch('feat/x', linked.path, 2, 5, 1), branch('spike', '', 1, 7, 0)],
  computedAt: '2026-10-05T08:00:00Z',
  generation: 1,
};

describe('dirtyFileCount', () => {
  it('sums every uncommitted file state but not the line counts', () => {
    expect(dirtyFileCount(linked.changes)).toBe(8);
  });
});

describe('projectRefEntries', () => {
  it('groups the main checkout, linked worktrees and worktree-less branches with commits, uncommitted files and lines (only when there are any)', () => {
    expect(projectRefEntries(git)).toEqual([
      { value: '', group: 'Checkout', label: 'main', title: '/work/gavel', stats: { commits: 0, uncommitted: 4, lines: { adds: 9, dels: 9 } }, lastCommitAt: DAYS_AGO_2, touchedAt: MINUTES_AGO_5 },
      { value: 'wt:/work/gavel-feat', group: 'Worktrees', label: 'feat/x', title: '/work/gavel-feat', stats: { commits: 2, uncommitted: 8, lines: { adds: 10, dels: 4 } }, lastCommitAt: HOURS_AGO_3, touchedAt: MINUTES_AGO_5 },
      { value: 'wt:/work/gavel-clean', group: 'Worktrees', label: 'chore/y', title: '/work/gavel-clean', stats: { commits: 1, uncommitted: 0 }, lastCommitAt: DAYS_AGO_2, touchedAt: undefined },
      { value: 'wt:/work/gavel-det', group: 'Worktrees', label: '(detached) gavel-det', title: '/work/gavel-det', stats: { commits: 0, uncommitted: 0 }, lastCommitAt: DAYS_AGO_2, touchedAt: undefined },
      { value: 'br:spike', group: 'Branches', label: 'spike', title: 'spike', stats: { commits: 1, uncommitted: 0, lines: { adds: 7, dels: 0 } }, lastCommitAt: HOURS_AGO_3 },
    ]);
  });

  it('omits branches that are checked out in a worktree', () => {
    expect(projectRefEntries(git).some(entry => entry.value === 'br:feat/x')).toBe(false);
  });

  it('labels a detached main checkout', () => {
    const [main] = projectRefEntries({ ...git, currentBranch: '', worktrees: [{ ...primary, branch: '' }], branches: [] });
    expect(main.label).toBe('(detached)');
  });
});

describe('projectRefAges', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(NOW);
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows the last commit and last touched ages', () => {
    expect(projectRefAges({ lastCommitAt: HOURS_AGO_3, touchedAt: MINUTES_AGO_5 })).toEqual({ committed: '3h', touched: '5m' });
  });

  it('leaves out a touched age for a clean checkout', () => {
    expect(projectRefAges({ lastCommitAt: DAYS_AGO_2 })).toEqual({ committed: '2d', touched: '' });
  });

  it.each([
    ['2026-10-04T11:59:52Z', 'git state 8s ago'],
    [MINUTES_AGO_5, 'git state 5m ago'],
    ['2026-10-04T11:59:58Z', 'git state just now'],
  ])('labels git state computed at %s as "%s"', (computedAt, label) => {
    expect(gitStateAge(computedAt)).toBe(label);
  });

  it("treats Go's zero time (a prunable worktree) as no commit age", () => {
    expect(projectRefAges({ lastCommitAt: GO_ZERO_TIME })).toEqual({ committed: '', touched: '' });
  });
});
