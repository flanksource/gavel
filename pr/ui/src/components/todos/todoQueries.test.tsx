import type React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { SessionStats } from '../../types';
import { useSessionStats } from './TodoSessionTimer';
import { sessionStatsQueryOptions, todoGlobalItemQueryOptions, todoItemQueryOptions, todoListQueryOptions, todoQueryKeys, workspaceTodoListQueryOptions } from './todoQueries';

function stats(overrides: Partial<SessionStats> = {}): SessionStats {
  return {
    durationMs: 1_000,
    inputTokens: 0,
    outputTokens: 0,
    cacheReadTokens: 0,
    cacheCreationTokens: 0,
    totalTokens: 0,
    contextTokens: 0,
    contextWindow: 0,
    turns: 0,
    compactions: 0,
    costUsd: 0,
    inProgress: false,
    found: true,
    ...overrides,
  };
}

function queryWrapper(client: QueryClient) {
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  };
}

function queryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, refetchOnWindowFocus: false },
    },
  });
}

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe('todo query keys', () => {
  it.each([
    ['stats workspace', todoQueryKeys.sessionStats('/repo-a', 'session-1'), todoQueryKeys.sessionStats('/repo-b', 'session-1')],
    ['stats session', todoQueryKeys.sessionStats('/repo', 'session-1'), todoQueryKeys.sessionStats('/repo', 'session-2')],
    ['detail workspace', todoQueryKeys.sessionDetail('/repo-a', 'todo-1'), todoQueryKeys.sessionDetail('/repo-b', 'todo-1')],
    ['detail todo', todoQueryKeys.sessionDetail('/repo', 'todo-1'), todoQueryKeys.sessionDetail('/repo', 'todo-2')],
    ['captain session', todoQueryKeys.captainSession('/api/captain/sessions/a'), todoQueryKeys.captainSession('/api/captain/sessions/b')],
    ['todo workspace', todoQueryKeys.list('/repo-a'), todoQueryKeys.list('/repo-b')],
    ['todo item', todoQueryKeys.item('/repo', 'todo-1'), todoQueryKeys.item('/repo', 'todo-2')],
    ['global todo item', todoQueryKeys.globalItem('todo-1'), todoQueryKeys.globalItem('todo-2')],
    ['plan todo', todoQueryKeys.plan('/repo', 'todo-1'), todoQueryKeys.plan('/repo', 'todo-2')],
    ['commit hash', todoQueryKeys.commitFiles('/repo', 'abc123'), todoQueryKeys.commitFiles('/repo', 'def456')],
    ['commit range base', todoQueryKeys.commitFiles('/repo', 'abc123', 'base1'), todoQueryKeys.commitFiles('/repo', 'abc123')],
    ['diff file', todoQueryKeys.commitDiff('/repo', 'abc123', '', 'one.go'), todoQueryKeys.commitDiff('/repo', 'abc123', '', 'two.go')],
    ['diff range base', todoQueryKeys.commitDiff('/repo', 'abc123', 'base1', 'one.go'), todoQueryKeys.commitDiff('/repo', 'abc123', '', 'one.go')],
    ['cmux agent', todoQueryKeys.cmuxSurface('/repo', 'claude'), todoQueryKeys.cmuxSurface('/repo', 'codex')],
  ])('isolates the %s request parameter', (_label, left, right) => {
    expect(left).not.toEqual(right);
  });
});

describe('todo list query', () => {
  it('loads the workspace list with a contextual error contract', async () => {
    const expected = { counts: { total: 0 }, items: [] };
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(expected), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    const result = await todoListQueryOptions('/repo').queryFn?.({ signal: new AbortController().signal } as never);

    expect(result).toEqual(expected);
    expect(fetchMock).toHaveBeenCalledWith('/api/todos?dir=%2Frepo', { signal: expect.any(AbortSignal) });
  });

  it('loads the flat all-workspace list excluding hidden projects and keeps it apart from per-dir lists', async () => {
    const expected = { counts: { total: 1 }, items: [{ ref: 'todo-1', dir: '/repo' }] };
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(expected), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    const options = workspaceTodoListQueryOptions('hidden-projects');
    const result = await options.queryFn?.({ signal: new AbortController().signal } as never);

    expect(result).toEqual(expected);
    expect(fetchMock).toHaveBeenCalledWith('/api/todos?exclude=hidden-projects', { signal: expect.any(AbortSignal) });
    expect(options.queryKey).toEqual(['todos', 'list', { exclude: 'hidden-projects' }]);
    expect(options.queryKey).not.toEqual(todoQueryKeys.list('/repo'));
  });

  it('refuses a flat list item that does not say which workspace it belongs to', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ counts: { total: 1 }, items: [{ ref: 'todo-1' }] }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })));

    await expect(workspaceTodoListQueryOptions('hidden-projects').queryFn?.({ signal: new AbortController().signal } as never))
      .rejects.toThrow('todo todo-1 omitted its workspace directory');
  });
});

describe('todo detail queries', () => {
  it('passes cancellation to workspace and global detail requests', async () => {
    const signals: AbortSignal[] = [];
    vi.stubGlobal('fetch', vi.fn((_url: RequestInfo | URL, init?: RequestInit) => {
      signals.push(init?.signal as AbortSignal);
      return Promise.resolve(new Response(JSON.stringify({ ref: 'todo-1', cwd: '/repo' }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }));
    }));
    const workspaceSignal = new AbortController().signal;
    const globalSignal = new AbortController().signal;

    await todoItemQueryOptions('/repo', 'todo-1').queryFn?.({ signal: workspaceSignal } as never);
    await todoGlobalItemQueryOptions('todo-1').queryFn?.({ signal: globalSignal } as never);

    expect(signals).toEqual([workspaceSignal, globalSignal]);
  });

  it('keeps details fresh until an explicit invalidation', () => {
    const options = todoItemQueryOptions('/repo', 'todo-1');
    expect(options.staleTime).toBe(Infinity);
    expect(options.refetchInterval).toBe(false);
    expect(options.refetchIntervalInBackground).toBe(false);
  });
});

describe('session stats query', () => {
  it('refreshes stats through invalidation without recurring reads', () => {
    const options = sessionStatsQueryOptions({ dir: '/repo', sessionId: 'session-1' });
    expect(options.staleTime).toBe(Infinity);
    expect(options.refetchInterval).toBe(false);
  });

  it('deduplicates concurrent consumers and reuses fresh cached data on remount', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(stats()), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);
    const client = queryClient();
    const wrapper = queryWrapper(client);

    const first = renderHook(
      () => [useSessionStats({ dir: '/repo', sessionId: 'session-1', active: true }), useSessionStats({ dir: '/repo', sessionId: 'session-1', active: true })],
      { wrapper },
    );
    await waitFor(() => expect(first.result.current[0].stats?.found).toBe(true));
    expect(fetchMock).toHaveBeenCalledTimes(1);
    first.unmount();

    const second = renderHook(() => useSessionStats({ dir: '/repo', sessionId: 'session-1', active: true }), { wrapper });
    expect(second.result.current.stats?.found).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    second.unmount();
    client.clear();
  });

  it('passes React Query cancellation to the stats request', async () => {
    let signal: AbortSignal | undefined;
    vi.stubGlobal('fetch', vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      signal = init?.signal ?? undefined;
      return new Promise<Response>(() => {});
    }));
    const client = queryClient();
    const hook = renderHook(() => useSessionStats({ dir: '/repo', sessionId: 'session-1', active: true }), {
      wrapper: queryWrapper(client),
    });

    await waitFor(() => expect(signal).toBeDefined());
    hook.unmount();
    expect(signal?.aborted).toBe(true);
    client.clear();
  });

  it('stops polling after the server reports final totals', async () => {
    vi.useFakeTimers();
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(stats()), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);
    const client = queryClient();
    const hook = renderHook(() => useSessionStats({ dir: '/repo', sessionId: 'session-1', active: true }), {
      wrapper: queryWrapper(client),
    });

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    await act(async () => {
      await vi.advanceTimersByTimeAsync(20_000);
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    hook.unmount();
    client.clear();
  });
});
