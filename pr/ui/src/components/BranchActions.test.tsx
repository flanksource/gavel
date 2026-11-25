import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { GitWorktree, ProjectGit } from '../types';
import type { BranchDeletePlan } from '../types/branchDelete';
import { queryTestWrapper } from './todos/queryTestWrapper';
import { selectedBranch, useBranchActions } from './BranchActions';

const changes = { staged: 0, unstaged: 0, both: 0, untracked: 0, conflict: 0, adds: 0, dels: 0 };
const worktree = (path: string, branch: string, extra: Partial<GitWorktree> = {}): GitWorktree => ({
  path, branch, head: 'h', primary: false, detached: false, prunable: false, changes, ahead: 0, lastCommitAt: '2026-10-01T00:00:00Z', ...extra,
});
const git: ProjectGit = {
  base: 'main', currentBranch: 'release', baseCheckedOut: false, upstream: 'origin/main', upstreamAhead: 0, upstreamBehind: 0,
  worktrees: [worktree('/work/acme', 'release', { primary: true }), worktree('/work/acme-topic', 'topic'), worktree('/work/acme-detached', '', { detached: true })],
  branches: [], computedAt: '2026-10-01T00:00:00Z', generation: 1,
};

describe('selectedBranch', () => {
  it.each([
    { ref: 'wt:/work/acme-topic', expected: { name: 'topic', deleteReason: '' } },
    { ref: 'br:parked', expected: { name: 'parked', deleteReason: '' } },
    { ref: 'br:main', expected: { name: 'main', deleteReason: 'main is the base branch.' } },
    { ref: '', expected: { name: 'release', deleteReason: 'release is checked out in the main checkout; switch it to another branch first.' } },
    { ref: 'wt:/work/acme-detached', expected: null },
  ])('resolves $ref', ({ ref, expected }) => {
    expect(selectedBranch(git, ref)).toEqual(expected);
  });
});

function Harness({ value, onDeleted }: { value: string; onDeleted: () => void }) {
  const { items, dialog } = useBranchActions({ projectName: 'acme', git, value, onDeleted });
  return <>{items.map((item, index) => <button key={index} type="button" disabled={item.disabled} onClick={item.onSelect}>{item.label}</button>)}{dialog}</>;
}

const json = (value: unknown, status = 200) => new Response(JSON.stringify(value), { status, headers: { 'Content-Type': 'application/json' } });

function setup(plan: BranchDeletePlan) {
  const posts: unknown[] = [];
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    if (init?.method === 'POST') {
      posts.push(JSON.parse(String(init.body)));
      return json({ worktreeRemoved: true, branchDeleted: true });
    }
    if (String(input) === '/api/projects/acme/branch/delete?branch=topic') return json(plan);
    return json({});
  }));
  const onDeleted = vi.fn();
  render(<Harness value="wt:/work/acme-topic" onDeleted={onDeleted} />, { wrapper: queryTestWrapper() });
  return { posts, onDeleted };
}

afterEach(() => { cleanup(); vi.unstubAllGlobals(); });

describe('branch delete', () => {
  const plan: BranchDeletePlan = { branch: 'topic', head: 'abc123', base: 'main', worktree: '/work/acme-topic', uncommitted: [], unmerged: [] };

  it('confirms a clean, merged branch with a single Delete click', async () => {
    const { posts, onDeleted } = setup(plan);
    fireEvent.click(screen.getByRole('button', { name: 'Delete topic…' }));
    const dialog = await screen.findByRole('dialog', { name: 'Delete topic' });
    await within(dialog).findByText(/Removes the worktree/);
    fireEvent.click(within(dialog).getByRole('button', { name: 'Delete topic' }));
    await waitFor(() => expect(onDeleted).toHaveBeenCalled());
    expect(posts).toEqual([{ branch: 'topic', expectedHead: 'abc123' }]);
  });

  it('keeps Delete disabled until every loss is acknowledged, then sends both acknowledgements', async () => {
    const { posts, onDeleted } = setup({ ...plan, uncommitted: ['notes.md'], unmerged: [{ sha: 'deadbeefcafe', subject: 'feat: wip' }] });
    fireEvent.click(screen.getByRole('button', { name: 'Delete topic…' }));
    const dialog = await screen.findByRole('dialog', { name: 'Delete topic' });
    expect(await within(dialog).findByText('notes.md')).toBeTruthy();
    expect(within(dialog).getByText('deadbeef feat: wip')).toBeTruthy();
    const confirm = within(dialog).getByRole('button', { name: 'Delete topic' }) as HTMLButtonElement;
    expect(confirm.disabled).toBe(true);
    fireEvent.click(within(dialog).getByRole('checkbox', { name: 'Discard 1 uncommitted file' }));
    expect(confirm.disabled).toBe(true);
    fireEvent.click(within(dialog).getByRole('checkbox', { name: 'Delete 1 unmerged commit' }));
    fireEvent.click(confirm);
    await waitFor(() => expect(onDeleted).toHaveBeenCalled());
    expect(posts).toEqual([{ branch: 'topic', expectedHead: 'abc123', discardUncommitted: true, deleteUnmerged: true }]);
  });
});
