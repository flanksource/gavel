import { act, renderHook, waitFor } from '@testing-library/react';
import { focusManager, QueryClient } from '@tanstack/react-query';
import { todoQueryKeys } from './todoQueries';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Project, TodoItem, TodoListResponse } from '../../types';
import { queryTestWrapper } from './queryTestWrapper';
import { useWorkspaceTodos } from './useWorkspaceTodos';

vi.mock('./useWorkspaceChanges', () => ({ useWorkspaceChanges: () => ({ ready: true, error: '' }) }));

vi.mock('./format', async importOriginal => {
  const emptyCounts = {
    total: 0,
    open: 0,
    draft: 0,
    pending: 0,
    inProgress: 0,
    review: 0,
    ask: 0,
    failed: 0,
    unverified: 0,
    verified: 0,
    completed: 0,
    skipped: 0,
  };
  return {
    ...await importOriginal<typeof import('./format')>(),
    emptyCounts,
    addCounts: (_left: unknown, right: unknown) => right,
    todoQuery: (dir: string) => {
      const params = new URLSearchParams();
      if (dir.trim()) params.set('dir', dir.trim());
      return params.toString();
    },
  };
});

const projects: Project[] = [{
  name: 'gavel',
  dir: '/work/gavel',
  repos: ['flanksource/gavel'],
}];

const counts = {
  total: 0,
  open: 0,
  draft: 0,
  pending: 0,
  inProgress: 0,
  review: 0,
  ask: 0,
  failed: 0,
  unverified: 0,
  verified: 0,
  completed: 0,
  skipped: 0,
};

const LIST_URL = '/api/todos?exclude=hidden-projects';

// The hook also loads the tag taxonomy (one cached request per workspace), so
// these assertions scope themselves to the list endpoint rather than counting
// every fetch — the behaviour under test is the list query, not the total.
function isTagRequest(input: RequestInfo | URL): boolean {
  return String(input).startsWith('/api/todos/labels');
}

function listCalls(mock: { mock: { calls: [RequestInfo | URL, RequestInit?][] } }) {
  return mock.mock.calls.filter(([input]) => String(input) === LIST_URL);
}

function jsonResponse(body: unknown, ok = true, status = ok ? 200 : 500): Response {
  return new Response(JSON.stringify(body), {
    status,
    statusText: ok ? 'OK' : 'Error',
    headers: { 'Content-Type': 'application/json' },
  });
}

beforeEach(() => {
  // Own the storage stub in this suite. Other component suites temporarily
  // replace the global, and Vitest may schedule files in the same worker.
  const store: Record<string, string> = {};
  vi.stubGlobal('localStorage', {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = String(value); }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { for (const key of Object.keys(store)) delete store[key]; }),
  });
});

afterEach(() => {
  focusManager.setFocused(undefined);
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('useWorkspaceTodos', () => {
  it('resolves the route ref before workspace lists finish and adopts canonical database identity without navigating', async () => {
    const routeRef = 'legacy/ref';
    const detail: TodoItem = {
      ref: 'e2a3b8c2-d0f7-c9a9-8b40-0dc78e8a94a5',
      id: 'e2a3b8c2-d0f7-c9a9-8b40-0dc78e8a94a5',
      cwd: '/canonical/gavel',
      title: 'Standardize on PostgreSQL',
      status: 'pending',
      priority: 'high',
    };
    let resolveList!: (response: Response) => void;
    const pendingList = new Promise<Response>(resolve => { resolveList = resolve; });
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith('/api/todos/item?')) return Promise.resolve(jsonResponse(detail));
      if (url === LIST_URL) return pendingList;
      throw new Error(`unexpected request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    const onNavigate = vi.fn();

    const { result, unmount } = renderHook(() => useWorkspaceTodos(projects, { selectedId: routeRef, onNavigate }), {
      wrapper: queryTestWrapper(),
    });

    await waitFor(() => expect(result.current.detail?.ref).toBe(detail.ref));
    expect(result.current.selected).toEqual({
      dir: detail.cwd,
      ref: detail.ref,
    });
    expect(result.current.loadingList).toBe(true);
    expect(onNavigate).not.toHaveBeenCalled();
    expect(fetchMock.mock.calls.map(([input]) => String(input)).filter(url => url.startsWith('/api/todos/item?'))).toEqual([
      '/api/todos/item?ref=legacy%2Fref',
    ]);

    const list: TodoListResponse = { dir: '/work/gavel', counts, items: [] };
    await act(async () => resolveList(jsonResponse(list)));
    await waitFor(() => expect(result.current.loadingList).toBe(false));
    unmount();
  });

  it('does not let the previous canonical detail overwrite a session-UUID route lookup', async () => {
    const todoID = '3785b0a4-0bf6-4f65-b1c2-41eab73e9f6b';
    const sessionID = '019f5b29-7890-7c11-8e7a-838e5d373e39';
    const base: TodoItem = {
      ref: todoID,
      id: todoID,
      cwd: '/work/gavel',
      title: 'Parse user shell commands',
      status: 'completed',
      priority: 'medium',
      sessionId: '019f5b2e-75b7-7de2-911b-de8b70266479',
    };
    const historical = { ...base, lookupSessionId: sessionID };
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url === LIST_URL) {
        return Promise.resolve(jsonResponse({ counts, items: [{ ...base, dir: '/work/gavel' }] }));
      }
      if (url === `/api/todos/item?ref=${todoID}`) return Promise.resolve(jsonResponse(base));
      if (url === `/api/todos/item?ref=${sessionID}`) return Promise.resolve(jsonResponse(historical));
      throw new Error(`stale or unexpected detail request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    const onNavigate = vi.fn();

    const { result, rerender } = renderHook(
      ({ selectedId }: { selectedId: string }) => useWorkspaceTodos(projects, { selectedId, onNavigate }),
      { initialProps: { selectedId: todoID }, wrapper: queryTestWrapper() },
    );
    await waitFor(() => expect(result.current.detail?.ref).toBe(todoID));

    rerender({ selectedId: sessionID });
    await waitFor(() => expect(result.current.detail?.lookupSessionId).toBe(sessionID));
    expect(result.current.selected).toEqual({ dir: '/work/gavel', ref: todoID });
    expect(fetchMock.mock.calls.map(([input]) => String(input)).filter(url => url.startsWith('/api/todos/item?'))).toEqual([
      `/api/todos/item?ref=${todoID}`,
      `/api/todos/item?ref=${sessionID}`,
    ]);
  });

  it.each([
    { status: 404, message: 'todo reference missing123 was not found' },
    { status: 409, message: 'todo reference abc123 is ambiguous' },
  ])('surfaces a $status database reference error instead of waiting for workspace discovery', async ({ status, message }) => {
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith('/api/todos/item?')) {
        return Promise.resolve(jsonResponse({ error: message }, false, status));
      }
      return Promise.resolve(jsonResponse({ counts, items: [] }));
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useWorkspaceTodos(projects, { selectedId: status === 404 ? 'missing123' : 'abc123' }), {
      wrapper: queryTestWrapper(),
    });

    const selectedId = status === 404 ? 'missing123' : 'abc123';
    await waitFor(() => expect(result.current.detailError).toBe(`Failed to resolve todo ${selectedId}: ${message}`));
    expect(result.current.detail).toBeNull();
    expect(result.current.selected).toBeNull();
  });

  it('aborts a stale workspace detail query when local selection changes', async () => {
    const detailSignals: AbortSignal[] = [];
    vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url === LIST_URL) {
        return Promise.resolve(jsonResponse({ counts, items: [] }));
      }
      if (url.startsWith('/api/todos/item?')) {
        detailSignals.push(init?.signal as AbortSignal);
        return new Promise<Response>(() => {});
      }
      throw new Error(`unexpected request: ${url}`);
    }));
    const { result, unmount } = renderHook(() => useWorkspaceTodos(projects), { wrapper: queryTestWrapper() });

    act(() => result.current.select({ dir: '/work/gavel', ref: 'todo-one' }));
    await waitFor(() => expect(detailSignals).toHaveLength(1));
    act(() => result.current.select({ dir: '/work/gavel', ref: 'todo-two' }));

    await waitFor(() => expect(detailSignals).toHaveLength(2));
    expect(detailSignals[0]?.aborted).toBe(true);
    unmount();
    expect(detailSignals[1]?.aborted).toBe(true);
  });

  it('keeps an unrouted detail view locally and resets it when another todo is selected', () => {
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(() => {})));
    const { result } = renderHook(() => useWorkspaceTodos(projects), { wrapper: queryTestWrapper() });

    act(() => result.current.select({ dir: '/work/gavel', ref: 'todo-one' }));
    act(() => result.current.setDetailView({ tab: 'session', sessionTab: 'costs' }));
    expect(result.current.detailView).toEqual({ tab: 'session', sessionTab: 'costs' });

    act(() => result.current.select({ dir: '/work/gavel', ref: 'todo-two' }));
    expect(result.current.detailView).toEqual({});
  });

  it('hands a routed detail view change to the router with its history mode', () => {
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(() => {})));
    const onViewChange = vi.fn();
    const routed = { tab: 'plan' as const };
    const { result } = renderHook(
      () => useWorkspaceTodos(projects, { view: routed, onViewChange }),
      { wrapper: queryTestWrapper() },
    );

    act(() => result.current.setDetailView({ ...routed, sessionIds: ['run-a'] }, 'replace'));

    expect(onViewChange).toHaveBeenCalledWith({ tab: 'plan', sessionIds: ['run-a'] }, 'replace');
    expect(result.current.detailView).toBe(routed);
  });

  it('keeps a database list failure visible while retaining an empty workspace group', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({ error: 'database unavailable: connection refused' }, false)));

    const { result } = renderHook(() => useWorkspaceTodos(projects), { wrapper: queryTestWrapper() });

    await waitFor(() => expect(result.current.error).toContain('database unavailable: connection refused'));
    expect(result.current.byDir['/work/gavel']?.items).toEqual([]);
    expect(result.current.listReady).toBe(false);
  });

  it('loads every visible workspace with one GET and groups the flat list by item dir', async () => {
    const configuredProjects: Project[] = [
      { name: 'second', dir: '/work/second', repos: [] },
      { name: 'first', dir: '/work/first', repos: [] },
      { name: 'duplicate', dir: '/work/first', repos: [] },
      { name: 'empty', dir: '/work/empty', repos: [] },
    ];
    const firstTodo: TodoItem = {
      ref: 'first-1', id: 'first-1', dir: '/work/first', title: 'First workspace todo',
      status: 'pending', priority: 'high',
    };
    const secondTodo: TodoItem = {
      ref: 'second-1', id: 'second-1', dir: '/work/second', title: 'Second workspace todo',
      status: 'completed', priority: 'low',
    };
    const secondOpen: TodoItem = { ...secondTodo, ref: 'second-2', id: 'second-2', status: 'in_progress' };
    const fetchMock = vi.fn(async (input: RequestInfo | URL, _init?: RequestInit) => {
      if (isTagRequest(input)) return jsonResponse({ definitions: [] });
      if (String(input) !== LIST_URL) throw new Error(`unexpected request: ${String(input)}`);
      return jsonResponse({ counts: { ...counts, total: 3 }, items: [firstTodo, secondTodo, secondOpen] });
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useWorkspaceTodos(configuredProjects), { wrapper: queryTestWrapper() });

    await waitFor(() => expect(result.current.listReady).toBe(true));
    expect(Object.fromEntries(Object.entries(result.current.byDir).map(([dir, list]) => [dir, list.items.map(item => item.ref)]))).toEqual({
      '/work/first': ['first-1'],
      '/work/second': ['second-1', 'second-2'],
      '/work/empty': [],
    });
    expect(result.current.byDir['/work/second']?.counts).toMatchObject({ total: 2, completed: 1, inProgress: 1, open: 1 });
    expect(result.current.workspaces.map(workspace => workspace.dir)).toEqual(['/work/second', '/work/first', '/work/empty']);
    expect(listCalls(fetchMock)).toHaveLength(1);
    expect(listCalls(fetchMock)[0][1]).toEqual({ signal: expect.any(AbortSignal) });
  });

  it('rejects a list whose items omit their workspace dir', async () => {
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => (
      isTagRequest(input)
        ? jsonResponse({ definitions: [] })
        : jsonResponse({ counts, items: [{ ref: 'orphan', title: 'No dir', status: 'pending', priority: 'low' }] })
    )));

    const { result } = renderHook(() => useWorkspaceTodos(projects), { wrapper: queryTestWrapper() });

    await waitFor(() => expect(result.current.error).toContain('todo orphan omitted its workspace directory'));
    expect(result.current.listReady).toBe(false);
  });

  it('refetches the active list query on explicit refresh', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, _init?: RequestInit) => (
      isTagRequest(input)
        ? jsonResponse({ definitions: [] })
        : jsonResponse({ counts, items: [] })
    ));
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useWorkspaceTodos(projects), { wrapper: queryTestWrapper() });
    await waitFor(() => expect(listCalls(fetchMock)).toHaveLength(1));

    act(() => result.current.refresh());

    await waitFor(() => expect(listCalls(fetchMock)).toHaveLength(2));
    expect(listCalls(fetchMock).map(([input]) => String(input))).toEqual([
      LIST_URL,
      LIST_URL,
    ]);
  });

  it('aborts the list request when its last observer unmounts', async () => {
    let requestSignal: AbortSignal | undefined;
    vi.stubGlobal('fetch', vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      requestSignal = init?.signal as AbortSignal;
      return new Promise<Response>(() => {});
    }));
    const { unmount } = renderHook(() => useWorkspaceTodos(projects), { wrapper: queryTestWrapper() });
    await waitFor(() => expect(requestSignal).toBeDefined());

    unmount();

    await waitFor(() => expect(requestSignal?.aborted).toBe(true));
  });

  it('refreshes an active detail only when explicitly invalidated', async () => {
    focusManager.setFocused(true);
    const active: TodoItem = {
      ref: 'run-1', id: 'run-1', cwd: '/work/gavel', title: 'Live run',
      status: 'in_progress', priority: 'medium',
    };
    const attached: TodoItem = { ...active, sessionId: 'codex-session-1' };
    const completed: TodoItem = { ...attached, status: 'review', hasPlan: true };
    const details = [active, attached, completed];
    const detailURLs: string[] = [];
    let detailRead = 0;
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith('/api/todos/item?')) {
        detailURLs.push(url);
        const body = details[Math.min(detailRead, details.length - 1)];
        detailRead++;
        return Promise.resolve(jsonResponse(body));
      }
      if (url === LIST_URL) {
        return Promise.resolve(jsonResponse({ counts, items: [] }));
      }
      throw new Error(`unexpected request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);

    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useWorkspaceTodos(projects, { selectedId: 'run-1' }), { wrapper: queryTestWrapper(client) });
    await waitFor(() => expect(result.current.detail?.status).toBe('in_progress'));
    expect(result.current.detail?.sessionId).toBeUndefined();

    await act(async () => { await client.invalidateQueries({ queryKey: todoQueryKeys.item('/work/gavel', 'run-1') }); });
    await waitFor(() => expect(result.current.detail?.sessionId).toBe('codex-session-1'), { timeout: 1_500 });
    expect(detailURLs).toEqual([
      '/api/todos/item?ref=run-1',
      '/api/todos/item?dir=%2Fwork%2Fgavel&ref=run-1',
    ]);
    expect(result.current.detail?.status).toBe('in_progress');

    await act(async () => { await client.invalidateQueries({ queryKey: todoQueryKeys.item('/work/gavel', 'run-1') }); });
    await waitFor(() => expect(result.current.detail?.status).toBe('review'));
    expect(result.current.detail?.hasPlan).toBe(true);
  });

  // Layout is a view preference like density and grouping: owned by the hook so
  // the navbar toggle and the body slots read one value, and written through to
  // storage so the choice survives a reload.
  it('defaults the layout to split and persists a change', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({ counts, items: [] })));

    const { result } = renderHook(() => useWorkspaceTodos([]), { wrapper: queryTestWrapper() });
    expect(result.current.layout).toBe('split');

    act(() => result.current.setLayout('full'));
    await waitFor(() => expect(result.current.layout).toBe('full'));
    expect(localStorage.getItem('gavel.pr-ui.todoLayout.v1')).toBe('full');
  });
});
