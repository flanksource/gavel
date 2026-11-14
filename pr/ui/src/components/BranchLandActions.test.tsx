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
   double renders its menu items inline so the secondary entries are reachable; the Modal
   double renders in place instead of in a portal. */
vi.mock('@flanksource/clicky-ui/components', async (importOriginal) => ({
  SegmentedControl: (await importOriginal<typeof import('@flanksource/clicky-ui/components')>()).SegmentedControl,
  Modal: ({ children, footer, title }: { children: ReactNode; footer: ReactNode; title: string }) => (
    // oxlint-disable-next-line clicky-ui/no-adhoc-overlay -- this IS the stub standing in for clicky's Modal.
    <div role="dialog" aria-label={title}>{children}{footer}</div>
  ),
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
      {/* react-doctor-disable-next-line react-doctor/no-array-index-as-key -- test stub renders a fixed menu list */}
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

// The merge is confirmed in a dialog that picks the method; these drive it.
function openMergeDialog() {
  fireEvent.click(screen.getByRole('button', { name: /Merge into main/ }));
  return screen.getByRole('dialog', { name: `Merge ${BRANCH} into main` });
}
const confirmButton = (name: RegExp) => screen.getByRole('button', { name }) as HTMLButtonElement;

describe('BranchLandActions', () => {
  it('squash-merges from the dialog by default without a message, closes it and refreshes git, status and summary', async () => {
    const fetchMock = stubLand(200, mergeResult);
    const { invalidate, onMerged } = renderActions();

    openMergeDialog();
    expect(fetchMock).not.toHaveBeenCalled();
    fireEvent.click(confirmButton(/^Squash & merge$/));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(posted(fetchMock, 'merge')).toEqual({ branch: BRANCH, mode: 'squash' });
    await waitFor(() => expect(onMerged).toHaveBeenCalledWith(mergeResult));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projectGitSummary() });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projectGit(PROJECT) });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatusScope(PROJECT) });
  });

  it('reports the merge even when the merged branch view unmounts before the response', async () => {
    const fetchMock = stubLand(200, mergeResult);
    const { onMerged, unmount } = renderActions();

    openMergeDialog();
    fireEvent.click(confirmButton(/^Squash & merge$/));
    unmount();

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(onMerged).toHaveBeenCalledWith(mergeResult));
  });

  it('sends the trimmed squash message typed in the dialog', async () => {
    const fetchMock = stubLand(200, mergeResult);
    renderActions();

    openMergeDialog();
    fireEvent.change(screen.getByRole('textbox', { name: 'Squash commit message' }), { target: { value: '  feat: git panel  ' } });
    fireEvent.click(confirmButton(/^Squash & merge$/));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(posted(fetchMock, 'merge')).toEqual({ branch: BRANCH, mode: 'squash', message: 'feat: git panel' });
  });

  it('picks each-commit in the dialog, hides the squash message and never sends it', async () => {
    const fetchMock = stubLand(200, { ...mergeResult, mode: 'incremental' });
    renderActions();

    const dialog = openMergeDialog();
    expect(Array.from(dialog.querySelectorAll('[role="radio"]')).map(radio => radio.textContent)).toEqual(['Squash', 'Each commit', 'With AI']);
    fireEvent.change(screen.getByRole('textbox', { name: 'Squash commit message' }), { target: { value: 'ignored' } });
    fireEvent.click(screen.getByRole('radio', { name: /Each commit/ }));
    expect(screen.queryByRole('textbox', { name: 'Squash commit message' })).toBeNull();
    fireEvent.click(confirmButton(/^Merge each commit$/));

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

  it('keeps the dialog open on a 409, listing the conflicting paths', async () => {
    stubLand(409, { error: 'squash of feat/git-panel onto main conflicts', conflicts: ['a.txt', 'dir/b.txt'] });
    renderActions();

    const dialog = openMergeDialog();
    fireEvent.click(confirmButton(/^Squash & merge$/));

    const alert = await screen.findByRole('alert');
    expect(dialog.contains(alert)).toBe(true);
    expect(alert.textContent).toContain('conflicts');
    expect(Array.from(alert.querySelectorAll('li')).map(item => item.textContent)).toEqual(['a.txt', 'dir/b.txt']);
    expect(confirmButton(/^Squash & merge$/).disabled).toBe(false);
  });

  it('shows a refused PR inline as a plain error', async () => {
    stubLand(409, { error: 'worktree is dirty' });
    renderActions();

    fireEvent.click(screen.getByRole('button', { name: /Create draft PR/ }));

    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toContain('worktree is dirty');
    expect(alert.querySelector('li')).toBeNull();
  });

  it.each([
    ['ahead', { ahead: 0 }, 'The branch has no commits ahead of main.'],
    ['dirty', { dirtyCount: 3 }, '3 uncommitted changes in the worktree: commit or discard them before landing.'],
  ])('says in the dialog, not inline, why a %s-blocked branch cannot merge, and disables PR', (_label, overrides, reason) => {
    const fetchMock = stubLand(200, mergeResult);
    renderActions(overrides);

    expect(screen.queryByText(reason)).toBeNull();
    expect((screen.getByRole('button', { name: /^Create PR$/ }) as HTMLButtonElement).disabled).toBe(true);
    const dialog = openMergeDialog();
    expect(dialog.textContent).toContain(reason);
    expect(confirmButton(/^Squash & merge$/).disabled).toBe(true);
    fireEvent.click(confirmButton(/^Squash & merge$/));
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('refuses only the merge when the main checkout is off the base branch', () => {
    renderActions({ baseCheckedOut: false });

    expect((screen.getByRole('button', { name: /^Create PR$/ }) as HTMLButtonElement).disabled).toBe(false);
    openMergeDialog();
    expect(screen.getByText('The main checkout is not on main: check out main there before merging.')).toBeTruthy();
    expect(confirmButton(/^Squash & merge$/).disabled).toBe(true);
    fireEvent.click(screen.getByRole('radio', { name: /Each commit/ }));
    expect(confirmButton(/^Merge each commit$/).disabled).toBe(true);
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
  // Opens the dialog on the AI method and returns its confirm button.
  const chooseAIMerge = () => {
    openMergeDialog();
    fireEvent.click(screen.getByRole('radio', { name: /With AI/ }));
    return () => confirmButton(/^Merge with AI$/);
  };

  function startedRun(fetchMock: ReturnType<typeof stubLand>, call = 0) {
    const [url, init] = fetchMock.mock.calls[call] as unknown as [string, RequestInit];
    expect(url).toBe('/api/v1/prompt/run?args=branch-merge');
    expect(init.method).toBe('POST');
    return JSON.parse(String(init.body)).variables;
  }

  it('runs the branch-merge prompt from the dialog, tails its session and reports the landed merge', async () => {
    const fetchMock = stubLand(200, { runId: RUN, status: 'running' });
    const { onMerged } = renderActions({ worktreePath: WORKTREE });

    const aiMerge = chooseAIMerge();
    await waitFor(() => expect(aiMerge().disabled).toBe(false));
    fireEvent.change(screen.getByRole('textbox', { name: 'Squash commit message' }), { target: { value: ' feat: git panel ' } });
    fireEvent.click(aiMerge());

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(startedRun(fetchMock)).toEqual({
      repo: REPO, worktree: WORKTREE, branch: BRANCH, base: 'main', mode: 'squash', message: 'feat: git panel', conflicts: [],
    });
    await waitFor(() => expect(streams.at(-1)?.url).toBe(`/api/captain/prompt/runs/${RUN}/stream`));
    expect(screen.getByText(/Starting the AI merge/)).toBeTruthy();

    emit('run', { runId: RUN, sessionId: SESSION });
    expect(screen.getByTestId('session-tail').textContent).toBe(`/api/captain/sessions/${SESSION}`);
    expect(aiMerge().disabled).toBe(true);
    expect(screen.getByRole('button', { name: /Merging with AI/ })).toBeTruthy();

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

    openMergeDialog();
    fireEvent.click(screen.getByRole('radio', { name: /Each commit/ }));
    fireEvent.click(confirmButton(/^Merge each commit$/));
    const resolve = await screen.findByRole('button', { name: /Resolve with AI & merge each commit/ });
    await waitFor(() => expect((resolve as HTMLButtonElement).disabled).toBe(false));
    fetchMock.mockImplementationOnce(async () => json(200, { runId: RUN, status: 'running' }));
    fireEvent.click(resolve);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(startedRun(fetchMock, 1)).toMatchObject({ mode: 'incremental', conflicts: ['a.txt'] });
    await waitFor(() => expect(screen.queryByRole('button', { name: /Resolve with AI/ })).toBeNull());
  });

  it('shows a failed run and reports no merge, also once the dialog is closed', async () => {
    stubLand(200, { runId: RUN, status: 'running' });
    const { onMerged } = renderActions({ worktreePath: WORKTREE });

    const aiMerge = chooseAIMerge();
    await waitFor(() => expect(aiMerge().disabled).toBe(false));
    fireEvent.click(aiMerge());
    await waitFor(() => expect(streams).toHaveLength(1));
    emit('error', { runId: RUN, success: false, error: 'workflow.after[0] "git merge --squash" failed' });

    expect((await screen.findByRole('alert')).textContent).toContain('AI merge failed: workflow.after[0]');
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(screen.getByRole('alert').textContent).toContain('AI merge failed: workflow.after[0]');
    expect(onMerged).not.toHaveBeenCalled();
    expect(streams[0]?.closed).toBe(true);
  });

  it('fails a run Captain has forgotten once the stream drops, instead of waiting forever', async () => {
    const fetchMock = stubLand(200, { runId: RUN, status: 'running' });
    renderActions({ worktreePath: WORKTREE });

    const aiMerge = chooseAIMerge();
    await waitFor(() => expect(aiMerge().disabled).toBe(false));
    fireEvent.click(aiMerge());
    await waitFor(() => expect(streams).toHaveLength(1));
    fetchMock.mockImplementationOnce(async () => new Response('unknown run', { status: 404 }));
    act(() => streams[0]?.listeners.get('error')?.(new Event('error')));

    expect((await screen.findByRole('alert')).textContent).toContain('no longer knows this merge run');
    expect(fetchMock.mock.calls.at(-1)?.[0]).toBe(`/api/captain/prompt/runs/${RUN}`);
    expect(streams[0]?.closed).toBe(true);
  });

  it('disables the AI merge for a branch without a worktree, saying why in the dialog', async () => {
    const fetchMock = stubLand(200, { runId: RUN });
    renderActions();

    const aiMerge = chooseAIMerge();
    await waitFor(() => expect(screen.getByRole('dialog').textContent).toMatch(/has none/));
    expect(aiMerge().disabled).toBe(true);
    fireEvent.click(aiMerge());
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
