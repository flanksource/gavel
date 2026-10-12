import type React from 'react';
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoRunLanding, TodoRunWorkspace, TodoSessionAttempt } from '../../types';
import type { LatestRunWorkspace } from './runWorkspace';
import { todoQueryKeys } from './todoQueries';
import { TodoRunLandActions } from './TodoRunLandActions';

/* oxlint-disable clicky-ui/prefer-clicky-components --
   These raw <button>s ARE the test doubles for clicky-ui's Button/SplitButton: the
   real SplitButton mounts floating-ui, which crashes under vitest. The SplitButton
   double renders its menu items inline so the draft-PR entry is reachable. */
vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, variant: _variant, size: _size, loading, disabled, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string; loading?: boolean }) => (
    <button type="button" disabled={disabled || loading} {...props}>{children}</button>
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

const DIR = '/repo';
const REF = 'todo-1';
const SETUP = '5e7a0000aaaa1111bbbb2222cccc3333dddd4444';
const HEAD = 'd0b5c966aaaa1111bbbb2222cccc3333dddd4444';
const LANDED = '1a2b3c4d5e6f1a2b3c4d5e6f1a2b3c4d5e6f1a2b';

const committed: TodoRunWorkspace = {
  worktree: { repo: DIR, branch: 'shell/289107d9', base: 'ba5e', setup: SETUP, head: HEAD, removed: true },
  commits: [{ sha: HEAD, message: 'feat: record the workspace' }],
};

function latestRun(workspace: TodoRunWorkspace, landing?: TodoRunLanding): LatestRunWorkspace {
  const attempt: TodoSessionAttempt = {
    promptRunId: 'run-2', ordinal: 2, step: 'run', requested: {}, resolved: {},
    status: 'completed', processActive: false, state: 'succeeded', phase: 'finished',
    queuedAt: '2026-09-14T10:00:00Z', admissionSessionId: 'admission-2',
    createdAt: '2026-09-14T10:00:00Z', updatedAt: '2026-09-14T10:01:00Z',
    verification: null, workspace, landing,
  };
  return { attempt, workspace };
}

function renderActions(latest: LatestRunWorkspace) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const invalidate = vi.spyOn(client, 'invalidateQueries');
  render(
    <QueryClientProvider client={client}>
      <TodoRunLandActions dir={DIR} todoRef={REF} latest={latest} />
    </QueryClientProvider>,
  );
  return { invalidate };
}

function stubLand(status: number, body: unknown) {
  const fetchMock = vi.fn(async () => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }));
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

function postedBody(fetchMock: ReturnType<typeof stubLand>) {
  const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
  expect(url).toBe('/api/todos/land');
  expect(init.method).toBe('POST');
  return JSON.parse(String(init.body));
}

const mergeLanding: TodoRunLanding = {
  promptRunId: 'run-2', via: 'merge', targetBranch: 'main', landedSha: LANDED, commitCount: 1, landedAt: '2026-09-14T11:00:00Z',
};

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('TodoRunLandActions', () => {
  it('posts a merge landing and refreshes the session detail and todo list', async () => {
    const fetchMock = stubLand(200, { landing: mergeLanding });
    const { invalidate } = renderActions(latestRun(committed));

    fireEvent.click(screen.getByRole('button', { name: /Merge into current branch/ }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(postedBody(fetchMock)).toEqual({ ref: REF, dir: DIR, via: 'merge' });
    await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionDetail(DIR, REF) }));
    expect(invalidate).toHaveBeenCalledWith({ queryKey: todoQueryKeys.list(DIR), exact: true });
    expect(await screen.findByText(/Merged into/)).toBeTruthy();
    expect(screen.queryByRole('button', { name: /Merge into current branch/ })).toBeNull();
  });

  it('posts a draft PR landing from the Create PR menu', async () => {
    const fetchMock = stubLand(200, {
      landing: { ...mergeLanding, via: 'pr', prNumber: 42, prUrl: 'https://github.com/acme/widgets/pull/42' },
    });
    renderActions(latestRun(committed));

    fireEvent.click(screen.getByRole('button', { name: /Create draft PR/ }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(postedBody(fetchMock)).toEqual({ ref: REF, dir: DIR, via: 'pr', draft: true });
  });

  it('shows a recorded PR landing with its link, target and sha, and offers no actions', () => {
    renderActions(latestRun(committed, {
      ...mergeLanding, via: 'pr', prNumber: 42, prUrl: 'https://github.com/acme/widgets/pull/42',
    }));

    const link = screen.getByRole('link', { name: /#42/ });
    expect(link.getAttribute('href')).toBe('https://github.com/acme/widgets/pull/42');
    expect(screen.getByText('main')).toBeTruthy();
    expect(screen.getByText('1a2b3c4')).toBeTruthy();
    expect(screen.queryByRole('button')).toBeNull();
  });

  it('renders a 409 conflict inline and keeps the actions', async () => {
    stubLand(409, { error: 'cherry-pick of d0b5c96 onto main conflicts in: a.txt, b.txt' });
    renderActions(latestRun(committed));

    fireEvent.click(screen.getByRole('button', { name: /Merge into current branch/ }));

    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toContain('conflicts in: a.txt, b.txt');
    expect(screen.getByRole('button', { name: /Merge into current branch/ })).toBeTruthy();
  });

  it('renders the cleanup failure of a recorded landing and still refreshes the detail', async () => {
    stubLand(500, { landing: mergeLanding, error: 'landed but could not remove worktree /repo/.worktrees/shell-289107d9' });
    const { invalidate } = renderActions(latestRun(committed));

    fireEvent.click(screen.getByRole('button', { name: /Merge into current branch/ }));

    expect((await screen.findByRole('alert')).textContent).toContain('could not remove worktree');
    expect(invalidate).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionDetail(DIR, REF) });
  });

  it.each([
    ['a kept worktree holding uncommitted paths', {
      worktree: { ...committed.worktree, removed: false, kept: true, keptReason: 'commit failed', dirty: ['go.mod'] },
      commits: committed.commits,
    }, /1 uncommitted path/],
    ['a run with no commits past setup', { worktree: { ...committed.worktree, head: SETUP }, commits: [] }, /no commits to land/],
  ])('disables landing for %s, saying why', (_case, workspace: TodoRunWorkspace, reason) => {
    const fetchMock = stubLand(200, { landing: mergeLanding });
    renderActions(latestRun(workspace));

    const merge = screen.getByRole('button', { name: /Merge into current branch/ }) as HTMLButtonElement;
    expect(merge.disabled).toBe(true);
    expect((screen.getByRole('button', { name: /^Create PR$/ }) as HTMLButtonElement).disabled).toBe(true);
    expect(merge.closest('[title]')?.getAttribute('title')).toMatch(reason);
    fireEvent.click(merge);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
