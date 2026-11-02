import type React from 'react';
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { queryKeys } from '../query';
import type { BranchMergeResult } from '../types';
import { BranchLandActions, branchAIMergeBlockedReason, branchLandBlockedReasons, type BranchLandActionsProps } from './BranchLandActions';

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
    items: { label: ReactNode; onSelect?: () => void; disabled?: boolean; title?: string }[];
    disabled?: boolean;
    loading?: boolean;
    title?: string;
  }) => (
    <div role="group" aria-label={title}>
      <button type="button" onClick={onClick} disabled={disabled || loading}>{label}</button>
      {items.map((item, index) => <button key={index} type="button" title={item.title} onClick={item.onSelect} disabled={disabled || item.disabled}>{item.label}</button>)}
    </div>
  ),
}));
// SessionTail is clicky-ui's (tested there); the double records which session it follows.
vi.mock('@flanksource/clicky-ui/ai', () => ({
  SessionTail: ({ src }: { src: string }) => <span data-testid="session-tail">{src}</span>,
}));
/* oxlint-enable clicky-ui/prefer-clicky-components */

// The run stream is the tab's event hub; this double hands each test the
// listeners the hook registered so it can play Captain's frames.
const streams = vi.hoisted(() => [] as { url: string; listeners: Map<string, (event: Event) => void>; closed: boolean }[]);
vi.mock('../eventHub', () => ({
  openEventStream: (url: string) => {
    const stream = { url, listeners: new Map<string, (event: Event) => void>(), closed: false };
    streams.push(stream);
    return {
      addEventListener: (name: string, listener: (event: Event) => void) => stream.listeners.set(name, listener),
      removeEventListener: (name: string, listener: (event: Event) => void) => {
        if (stream.listeners.get(name) === listener) stream.listeners.delete(name);
      },
      close: () => { stream.closed = true; },
    };
  },
}));

function emit(name: string, data: unknown) {
  const stream = streams.at(-1);
  if (!stream) throw new Error('no run stream was opened');
  act(() => stream.listeners.get(name)?.(new MessageEvent(name, { data: JSON.stringify(data) })));
}

const PROJECT = 'gavel';
const BRANCH = 'feat/git-panel';
const REPO = '/src/gavel';
const WORKTREE = '/src/gavel/.worktrees/git-panel';

const changes = { staged: 0, unstaged: 0, both: 0, untracked: 0, conflict: 0, adds: 0, dels: 0 };
const projectGit = {
  base: 'main', currentBranch: 'main', baseCheckedOut: true, upstreamAhead: 0, upstreamBehind: 0, computedAt: '2026-10-13T00:00:00Z', generation: 1, branches: [],
  worktrees: [
    { path: REPO, branch: 'main', head: 'aaa', primary: true, detached: false, prunable: false, changes, ahead: 0, lastCommitAt: '2026-10-13T00:00:00Z' },
    { path: WORKTREE, branch: BRANCH, head: 'bbb', primary: false, detached: false, prunable: false, changes, ahead: 3, lastCommitAt: '2026-10-13T00:00:00Z' },
  ],
};

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

const json = (status: number, body: unknown) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

// Every POST answers status/body; the project's git state (read for the
// primary checkout) is served separately and never counted as a land request.
function stubLand(status: number, body: unknown) {
  const fetchMock = vi.fn(async (_url: string, _init?: RequestInit) => json(status, body));
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => (
    url === `/api/projects/${PROJECT}/git` ? json(200, projectGit) : fetchMock(url, init)
  )));
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
  streams.length = 0;
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

describe('branchAIMergeBlockedReason', () => {
  it.each([
    ['a refused merge', { merge: 'The main checkout is not on main.', worktreePath: WORKTREE, repo: REPO }, /not on main/],
    ['a branch with no worktree', { merge: '', repo: REPO }, /has none/],
    ['an unloaded primary checkout', { merge: '', worktreePath: WORKTREE }, /Loading/],
    ['a branch ready to rebase', { merge: '', worktreePath: WORKTREE, repo: REPO }, /^$/],
  ])('explains %s', (_label, input, expected) => {
    expect(branchAIMergeBlockedReason(input)).toMatch(expected);
  });
});

describe('BranchLandActions AI merge', () => {
  const RUN = 'run-1';
  const SESSION = 'session-1';
  const aiMergeItem = () => screen.getByRole('button', { name: /Merge with AI/ }) as HTMLButtonElement;

  function startedRun(fetchMock: ReturnType<typeof stubLand>, call = 0) {
    const [url, init] = fetchMock.mock.calls[call] as unknown as [string, RequestInit];
    expect(url).toBe('/api/v1/prompt/run?args=branch-merge');
    expect(init.method).toBe('POST');
    return JSON.parse(String(init.body)).variables;
  }

  it('runs the branch-merge prompt from the menu, tails its session and reports the landed merge', async () => {
    const fetchMock = stubLand(200, { runId: RUN, status: 'running' });
    const { onMerged } = renderActions({ worktreePath: WORKTREE });

    await waitFor(() => expect(aiMergeItem().disabled).toBe(false));
    fireEvent.change(screen.getByRole('textbox', { name: 'Squash commit message' }), { target: { value: ' feat: git panel ' } });
    fireEvent.click(aiMergeItem());

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(startedRun(fetchMock)).toEqual({
      repo: REPO, worktree: WORKTREE, branch: BRANCH, base: 'main', mode: 'squash', message: 'feat: git panel', conflicts: [],
    });
    await waitFor(() => expect(streams.at(-1)?.url).toBe(`/api/captain/prompt/runs/${RUN}/stream`));
    expect(screen.getByText(/Starting the AI merge/)).toBeTruthy();

    emit('run', { runId: RUN, sessionId: SESSION });
    expect(screen.getByTestId('session-tail').textContent).toBe(`/api/captain/sessions/${SESSION}`);
    expect((screen.getByRole('button', { name: /Squash & merge into main/ }) as HTMLButtonElement).disabled).toBe(true);

    emit('done', { runId: RUN, sessionId: SESSION, success: true });
    expect(onMerged).toHaveBeenCalledWith({
      targetBranch: 'main', landedSha: '', mode: 'squash', commits: 3, worktreeRemoved: true, branchDeleted: true,
    });
    expect(streams.at(-1)?.closed).toBe(true);
    expect(screen.queryByTestId('session-tail')).toBeNull();
  });

  it('offers to resolve a conflicting land with AI in the mode that failed, listing the conflicts', async () => {
    const fetchMock = stubLand(409, { error: 'cherry-pick of feat/git-panel onto main conflicts', conflicts: ['a.txt'] });
    renderActions({ worktreePath: WORKTREE });

    fireEvent.click(screen.getByRole('button', { name: /Merge each commit/ }));
    const resolve = await screen.findByRole('button', { name: /Resolve with AI & merge each commit/ });
    await waitFor(() => expect((resolve as HTMLButtonElement).disabled).toBe(false));
    fetchMock.mockImplementationOnce(async () => json(200, { runId: RUN, status: 'running' }));
    fireEvent.click(resolve);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(startedRun(fetchMock, 1)).toMatchObject({ mode: 'incremental', conflicts: ['a.txt'] });
    await waitFor(() => expect(screen.queryByRole('button', { name: /Resolve with AI/ })).toBeNull());
  });

  it('shows a failed run inline and reports no merge', async () => {
    stubLand(200, { runId: RUN, status: 'running' });
    const { onMerged } = renderActions({ worktreePath: WORKTREE });

    await waitFor(() => expect(aiMergeItem().disabled).toBe(false));
    fireEvent.click(aiMergeItem());
    await waitFor(() => expect(streams).toHaveLength(1));
    emit('error', { runId: RUN, success: false, error: 'workflow.after[0] "git merge --squash" failed' });

    expect((await screen.findByRole('alert')).textContent).toContain('AI merge failed: workflow.after[0]');
    expect(onMerged).not.toHaveBeenCalled();
    expect(streams[0]?.closed).toBe(true);
  });

  it('fails a run Captain has forgotten once the stream drops, instead of waiting forever', async () => {
    const fetchMock = stubLand(200, { runId: RUN, status: 'running' });
    renderActions({ worktreePath: WORKTREE });

    await waitFor(() => expect(aiMergeItem().disabled).toBe(false));
    fireEvent.click(aiMergeItem());
    await waitFor(() => expect(streams).toHaveLength(1));
    fetchMock.mockImplementationOnce(async () => new Response('unknown run', { status: 404 }));
    act(() => streams[0]?.listeners.get('error')?.(new Event('error')));

    expect((await screen.findByRole('alert')).textContent).toContain('no longer knows this merge run');
    expect(fetchMock.mock.calls.at(-1)?.[0]).toBe(`/api/captain/prompt/runs/${RUN}`);
    expect(streams[0]?.closed).toBe(true);
  });

  it('disables the AI merge for a branch without a worktree, saying why', async () => {
    const fetchMock = stubLand(200, { runId: RUN });
    renderActions();

    await waitFor(() => expect(aiMergeItem().getAttribute('title')).toMatch(/has none/));
    expect(aiMergeItem().disabled).toBe(true);
    fireEvent.click(aiMergeItem());
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
