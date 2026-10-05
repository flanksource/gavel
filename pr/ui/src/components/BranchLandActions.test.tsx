import type React from 'react';
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { queryKeys } from '../query';
import type { BranchMergeResult } from '../types';
import { BranchLandActions, branchLandBlockedReasons, type BranchLandActionsProps } from './BranchLandActions';

/* oxlint-disable clicky-ui/prefer-clicky-components --
   These raw elements ARE the test doubles for clicky-ui's Button/SplitButton/InputField:
   the real SplitButton mounts floating-ui, which crashes under vitest. The SplitButton
   double renders its menu items inline so the secondary entries are reachable. */
vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, variant: _variant, size: _size, loading, disabled, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string; loading?: boolean }) => (
    <button type="button" disabled={disabled || loading} {...props}>{children}</button>
  ),
  InputField: ({ value, onChange, disabled, ...props }: { value?: string; onChange?: (value: string) => void; disabled?: boolean; 'aria-label'?: string; placeholder?: string }) => (
    <input value={value} disabled={disabled} onChange={event => onChange?.(event.currentTarget.value)} {...props} />
  ),
  SplitButton: ({ label, onClick, items, disabled, loading, title }: {
    label: ReactNode;
    onClick: () => void;
    items: { label: ReactNode; onSelect?: () => void }[];
    disabled?: boolean;
    loading?: boolean;
    title?: string;
  }) => (
    <div role="group" aria-label={title}>
      <button type="button" onClick={onClick} disabled={disabled || loading}>{label}</button>
      {items.map((item, index) => <button key={index} type="button" onClick={item.onSelect} disabled={disabled}>{item.label}</button>)}
    </div>
  ),
}));
/* oxlint-enable clicky-ui/prefer-clicky-components */

const PROJECT = 'gavel';
const BRANCH = 'feat/git-panel';

const ready: BranchLandActionsProps = {
  projectName: PROJECT,
  branch: BRANCH,
  base: 'main',
  ahead: 3,
  dirtyCount: 0,
  baseCheckedOut: true,
};

const mergeResult: BranchMergeResult = {
  targetBranch: 'main', landedSha: '1a2b3c4d5e6f1a2b3c4d5e6f1a2b3c4d5e6f1a2b', mode: 'squash', commits: 3, worktreeRemoved: true, branchDeleted: true,
};

function renderActions(props: Partial<BranchLandActionsProps> = {}) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const invalidate = vi.spyOn(client, 'invalidateQueries');
  const onMerged = vi.fn();
  const { unmount } = render(
    <QueryClientProvider client={client}>
      <BranchLandActions {...ready} onMerged={onMerged} {...props} />
    </QueryClientProvider>,
  );
  return { invalidate, onMerged, unmount };
}

function stubLand(status: number, body: unknown) {
  const fetchMock = vi.fn(async () => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }));
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

function posted(fetchMock: ReturnType<typeof stubLand>, path: string) {
  const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
  expect(url).toBe(`/api/projects/${PROJECT}/branch/${path}`);
  expect(init.method).toBe('POST');
  return JSON.parse(String(init.body));
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('branchLandBlockedReasons', () => {
  it.each([
    ['a branch with nothing past main', { ahead: 0 }, { merge: /no commits ahead of main/, pr: /no commits ahead of main/ }],
    ['a worktree with uncommitted changes', { dirtyCount: 2, worktreePath: '/work/gavel-feat' }, { merge: /2 uncommitted changes.*\/work\/gavel-feat/, pr: /2 uncommitted changes/ }],
    ['a main checkout that is off the base branch', { baseCheckedOut: false }, { merge: /not on main/, pr: '' }],
    ['a branch ready to land', {}, { merge: '', pr: '' }],
  ])('explains %s', (_label, overrides, expected) => {
    const reasons = branchLandBlockedReasons({ ...ready, ...overrides });

    for (const via of ['merge', 'pr'] as const) {
      const want = expected[via];
      if (want === '') expect(reasons[via]).toBe('');
      else expect(reasons[via]).toMatch(want);
    }
  });

  it('pluralises one uncommitted change', () => {
    expect(branchLandBlockedReasons({ ...ready, dirtyCount: 1 }).merge).toMatch(/1 uncommitted change[^s]/);
  });
});

describe('BranchLandActions', () => {
  it('squash-merges by default without a message and refreshes git, status and summary', async () => {
    const fetchMock = stubLand(200, mergeResult);
    const { invalidate, onMerged } = renderActions();

    fireEvent.click(screen.getByRole('button', { name: /Squash & merge into main/ }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(posted(fetchMock, 'merge')).toEqual({ branch: BRANCH, mode: 'squash' });
    await waitFor(() => expect(onMerged).toHaveBeenCalledWith(mergeResult));
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projectGitSummary() });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projectGit(PROJECT) });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatusScope(PROJECT) });
  });

  it('reports the merge even when the merged branch view unmounts before the response', async () => {
    const fetchMock = stubLand(200, mergeResult);
    const { onMerged, unmount } = renderActions();

    fireEvent.click(screen.getByRole('button', { name: /Squash & merge into main/ }));
    unmount();

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(onMerged).toHaveBeenCalledWith(mergeResult));
  });

  it('sends the trimmed squash message', async () => {
    const fetchMock = stubLand(200, mergeResult);
    renderActions();

    fireEvent.change(screen.getByRole('textbox', { name: 'Squash commit message' }), { target: { value: '  feat: git panel  ' } });
    fireEvent.click(screen.getByRole('button', { name: /Squash & merge into main/ }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(posted(fetchMock, 'merge')).toEqual({ branch: BRANCH, mode: 'squash', message: 'feat: git panel' });
  });

  it('merges incrementally from the menu and never sends the squash message', async () => {
    const fetchMock = stubLand(200, { ...mergeResult, mode: 'incremental' });
    renderActions();

    fireEvent.change(screen.getByRole('textbox', { name: 'Squash commit message' }), { target: { value: 'ignored' } });
    fireEvent.click(screen.getByRole('button', { name: /Merge each commit/ }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(posted(fetchMock, 'merge')).toEqual({ branch: BRANCH, mode: 'incremental' });
  });

  it.each([
    ['Create PR', false],
    ['Create draft PR', true],
  ])('opens a PR from "%s" and links it', async (name, draft) => {
    const fetchMock = stubLand(200, { number: 42, url: 'https://github.com/acme/widgets/pull/42', topicBranch: 'gavel/feat-git-panel' });
    const { onMerged } = renderActions();

    fireEvent.click(screen.getByRole('button', { name: new RegExp(`^${name}$`) }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(posted(fetchMock, 'pr')).toEqual({ branch: BRANCH, draft });
    const link = await screen.findByRole('link', { name: /#42/ });
    expect(link.getAttribute('href')).toBe('https://github.com/acme/widgets/pull/42');
    expect(onMerged).not.toHaveBeenCalled();
  });

  it('lists the conflicting paths of a 409 and keeps the actions', async () => {
    stubLand(409, { error: 'squash of feat/git-panel onto main conflicts', conflicts: ['a.txt', 'dir/b.txt'] });
    renderActions();

    fireEvent.click(screen.getByRole('button', { name: /Squash & merge into main/ }));

    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toContain('conflicts');
    expect(Array.from(alert.querySelectorAll('li')).map(item => item.textContent)).toEqual(['a.txt', 'dir/b.txt']);
    expect(screen.getByRole('button', { name: /Squash & merge into main/ })).toBeTruthy();
  });

  it('shows a 409 without conflicts as a plain error', async () => {
    stubLand(409, { error: 'worktree is dirty' });
    renderActions();

    fireEvent.click(screen.getByRole('button', { name: /Create draft PR/ }));

    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toContain('worktree is dirty');
    expect(alert.querySelector('li')).toBeNull();
  });

  it.each([
    ['ahead', { ahead: 0 }, /no commits ahead/],
    ['dirty', { dirtyCount: 3 }, /3 uncommitted changes/],
  ])('disables merge and PR for a branch that is %s-blocked, saying why', (_label, overrides, reason) => {
    const fetchMock = stubLand(200, mergeResult);
    renderActions(overrides);

    const merge = screen.getByRole('button', { name: /Squash & merge into main/ }) as HTMLButtonElement;
    expect(merge.disabled).toBe(true);
    expect((screen.getByRole('button', { name: /^Create PR$/ }) as HTMLButtonElement).disabled).toBe(true);
    expect(merge.closest('[title]')?.getAttribute('title')).toMatch(reason);
    fireEvent.click(merge);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('disables only the merge when the main checkout is off the base branch', () => {
    renderActions({ baseCheckedOut: false });

    expect((screen.getByRole('button', { name: /Squash & merge into main/ }) as HTMLButtonElement).disabled).toBe(true);
    expect((screen.getByRole('button', { name: /Merge each commit/ }) as HTMLButtonElement).disabled).toBe(true);
    expect((screen.getByRole('button', { name: /^Create PR$/ }) as HTMLButtonElement).disabled).toBe(false);
    expect(screen.getByRole('button', { name: /Squash & merge into main/ }).closest('[title]')?.getAttribute('title')).toMatch(/not on main/);
  });
});
