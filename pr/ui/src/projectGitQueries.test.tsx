import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { PropsWithChildren } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  PROJECT_GIT_SUMMARY_CACHE_KEY,
  parseProjectGit,
  parseProjectGitSummaries,
  useProjectGit,
  useProjectGitSummary,
} from './projectGitQueries';
import type { ProjectGit, ProjectGitSummary } from './types';

const summaries: ProjectGitSummary[] = [
  { name: 'gavel', base: 'main', adds: 12, dels: 3, worktrees: 2, branches: 1 },
  { name: 'broken', base: '', adds: 0, dels: 0, worktrees: 0, branches: 0, error: 'not a git repository' },
];

const git: ProjectGit = {
  base: 'main',
  currentBranch: 'main',
  baseCheckedOut: true,
  worktrees: [{
    path: '/work/gavel', branch: 'main', head: 'a1b2c3d', primary: true, detached: false, prunable: false,
    changes: { staged: 0, unstaged: 1, both: 0, untracked: 0, conflict: 0, adds: 2, dels: 1 }, ahead: 0,
    lastCommitAt: '2026-10-01T08:00:00Z', touchedAt: '2026-10-04T08:00:00Z',
  }],
  branches: [{ name: 'feat/x', head: 'e4f5a6b', ahead: 2, behind: 0, worktree: '', diff: { commits: 2, files: 3, adds: 10, dels: 4 }, lastCommitAt: '2026-10-02T08:00:00Z' }],
  computedAt: '2026-10-05T08:00:00Z',
  generation: 7,
};

function wrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } } });
  return ({ children }: PropsWithChildren) => <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

beforeEach(() => {
  const values = new Map<string, string>();
  vi.stubGlobal('localStorage', {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => values.set(key, value),
    removeItem: (key: string) => values.delete(key),
    clear: () => values.clear(),
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('parseProjectGitSummaries', () => {
  it('accepts the git-summary payload', () => {
    expect(parseProjectGitSummaries(summaries)).toEqual(summaries);
  });

  it.each([
    ['a non-array payload', { projects: [] }],
    ['an entry without a name', [{ ...summaries[0], name: undefined }]],
    ['a non-numeric count', [{ ...summaries[0], adds: '12' }]],
    ['a non-string error', [{ ...summaries[0], error: 7 }]],
  ])('rejects %s', (_label, payload) => {
    expect(() => parseProjectGitSummaries(payload)).toThrow(/git summary/i);
  });
});

describe('parseProjectGit', () => {
  it('accepts the project git payload', () => {
    expect(parseProjectGit(git)).toEqual(git);
  });

  it('accepts a failed ref scan and a failed worktree status scan', () => {
    const failed: ProjectGit = {
      ...git,
      error: 'git for-each-ref: exit status 128',
      worktrees: [{ ...git.worktrees[0], statusError: 'git status: index.lock exists' }],
    };
    expect(parseProjectGit(failed)).toEqual(failed);
  });

  it.each([
    ['a missing worktree list', { ...git, worktrees: undefined }],
    ['a worktree without change counts', { ...git, worktrees: [{ ...git.worktrees[0], changes: undefined }] }],
    ['a branch without a diff', { ...git, branches: [{ ...git.branches[0], diff: undefined }] }],
    ['a non-boolean baseCheckedOut', { ...git, baseCheckedOut: 'yes' }],
    ['a worktree without a last commit time', { ...git, worktrees: [{ ...git.worktrees[0], lastCommitAt: undefined }] }],
    ['a worktree with a non-string touched time', { ...git, worktrees: [{ ...git.worktrees[0], touchedAt: 7 }] }],
    ['a branch without a last commit time', { ...git, branches: [{ ...git.branches[0], lastCommitAt: undefined }] }],
    ['a state without the time it was computed', { ...git, computedAt: undefined }],
    ['a state without a generation', { ...git, generation: undefined }],
    ['a state with a non-numeric generation', { ...git, generation: '7' }],
    ['a state with a non-string error', { ...git, error: 500 }],
    ['a worktree with a non-string statusError', { ...git, worktrees: [{ ...git.worktrees[0], statusError: true }] }],
  ])('rejects %s', (_label, payload) => {
    expect(() => parseProjectGit(payload)).toThrow(/project git/i);
  });
});

describe('useProjectGitSummary', () => {
  it('indexes the fetched summaries by project and caches them for the next load', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(summaries), { status: 200 })));

    const { result } = renderHook(() => useProjectGitSummary({ enabled: true }), { wrapper: wrapper() });

    expect(result.current.loading).toBe(true);
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('');
    expect(result.current.byProject.get('gavel')).toEqual(summaries[0]);
    expect(JSON.parse(localStorage.getItem(PROJECT_GIT_SUMMARY_CACHE_KEY) ?? 'null')).toEqual(summaries);
  });

  it('renders the cached summaries while the request is in flight', () => {
    localStorage.setItem(PROJECT_GIT_SUMMARY_CACHE_KEY, JSON.stringify(summaries));
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(() => {})));

    const { result } = renderHook(() => useProjectGitSummary({ enabled: true }), { wrapper: wrapper() });

    expect(result.current.loading).toBe(false);
    expect(result.current.byProject.get('gavel')?.adds).toBe(12);
  });

  it('surfaces a failed request as an error with the server reason', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: 'git unavailable' }), { status: 500 })));

    const { result } = renderHook(() => useProjectGitSummary({ enabled: true }), { wrapper: wrapper() });

    await waitFor(() => expect(result.current.error).toContain('git unavailable'));
    expect(result.current.loading).toBe(false);
    expect(result.current.byProject.size).toBe(0);
  });

  it('does not poll: the summary is fetched once and left to the git stream to invalidate', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(summaries), { status: 200 }));
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useProjectGitSummary({ enabled: true }), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    await vi.advanceTimersByTimeAsync(5 * 60_000);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    vi.useRealTimers();
  });

  it('does not poll the project git state either', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(git), { status: 200 }));
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useProjectGit('gavel'), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.data).toEqual(git));

    await vi.advanceTimersByTimeAsync(5 * 60_000);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    vi.useRealTimers();
  });

  it('re-reads the project git state on the interval a page asks for', async () => {
    const intervalMs = 5_000;
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(git), { status: 200 }));
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useProjectGit('gavel', { refetchIntervalMs: intervalMs }), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.data).toEqual(git));

    await vi.advanceTimersByTimeAsync(2 * intervalMs);

    expect(fetchMock).toHaveBeenCalledTimes(3);
    vi.useRealTimers();
  });

  it('does not fetch while disabled', () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);

    renderHook(() => useProjectGitSummary({ enabled: false }), { wrapper: wrapper() });

    expect(fetchMock).not.toHaveBeenCalled();
  });
});
