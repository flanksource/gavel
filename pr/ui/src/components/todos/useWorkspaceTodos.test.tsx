import { act, renderHook, waitFor } from '@testing-library/react';
import { focusManager } from '@tanstack/react-query';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Project, TodoItem, TodoListResponse } from '../../types';
import { queryTestWrapper } from './queryTestWrapper';
import { useWorkspaceTodos } from './useWorkspaceTodos';

vi.mock('./format', () => {
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

// The hook also loads the tag taxonomy (one cached request per workspace), so
// these assertions scope themselves to the batch endpoint rather than counting
// every fetch — the behaviour under test is the batch query, not the total.
function isTagRequest(input: RequestInfo | URL): boolean {
  return String(input).startsWith('/api/todos/labels');
}

function batchCalls(mock: { mock: { calls: [RequestInfo | URL, RequestInit?][] } }) {
  return mock.mock.calls.filter(([input]) => String(input) === '/api/todos/batch');
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
      if (url === '/api/todos/batch') return pendingList;
      throw new Error(`unexpected request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    const onNavigate = vi.fn();

    const { result, unmount } = renderHook(() => useWorkspaceTodos(projects, routeRef, onNavigate), {
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
    await act(async () => resolveList(jsonResponse({ results: [list] })));
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
      if (url === '/api/todos/batch') {
        return Promise.resolve(jsonResponse({ results: [{ dir: '/work/gavel', counts, items: [base] }] }));
      }
      if (url === `/api/todos/item?ref=${todoID}`) return Promise.resolve(jsonResponse(base));
      if (url === `/api/todos/item?ref=${sessionID}`) return Promise.resolve(jsonResponse(historical));
      throw new Error(`stale or unexpected detail request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    const onNavigate = vi.fn();

    const { result, rerender } = renderHook(
      ({ selectedId }: { selectedId: string }) => useWorkspaceTodos(projects, selectedId, onNavigate),
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
      return Promise.resolve(jsonResponse({ results: [{ dir: '/work/gavel', counts, items: [] }] }));
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useWorkspaceTodos(projects, status === 404 ? 'missing123' : 'abc123'), {
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
      if (url === '/api/todos/batch') {
        return Promise.resolve(jsonResponse({ results: [{ dir: '/work/gavel', counts, items: [] }] }));
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

  it('keeps a database list failure visible while retaining an empty workspace group', async () => {
    const workspaceError = { code: 'load_failed', message: 'database unavailable: connection refused' };
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({
      results: [{ dir: '/work/gavel', items: null, error: workspaceError }],
    })));

    const { result } = renderHook(() => useWorkspaceTodos(projects), { wrapper: queryTestWrapper() });

    await waitFor(() => expect(result.current.error).toBe('database unavailable: connection refused'));
    expect(result.current.byDir['/work/gavel']?.items).toEqual([]);
    expect(result.current.errorsByDir).toEqual({ '/work/gavel': workspaceError });
  });

  it('loads the complete normalized workspace set with one ordered batch request', async () => {
    const configuredProjects: Project[] = [
      { name: 'second', dir: '/work/second/', repos: [] },
      { name: 'first', dir: ' /work/first/./ ', repos: [] },
      { name: 'duplicate', dir: '/work/first', repos: [] },
    ];
    const firstTodo: TodoItem = {
      ref: 'first-1', id: 'first-1', cwd: '/work/first', title: 'First workspace todo',
      status: 'pending', priority: 'high',
    };
    const fetchMock = vi.fn(async (input: RequestInfo | URL, _init?: RequestInit) => {
      if (isTagRequest(input)) return jsonResponse({ definitions: [] });
      if (String(input) !== '/api/todos/batch') throw new Error(`unexpected request: ${String(input)}`);
      return jsonResponse({ results: [
        { dir: '/work/first', counts: { ...counts, total: 1, open: 1, pending: 1 }, items: [firstTodo] },
        { dir: '/work/second', counts, items: [] },
      ] });
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useWorkspaceTodos(configuredProjects), { wrapper: queryTestWrapper() });

    await waitFor(() => expect(result.current.byDir['/work/first']?.items).toEqual([firstTodo]));
    expect(result.current.workspaces.map(workspace => workspace.dir)).toEqual(['/work/second', '/work/first']);
    expect(batchCalls(fetchMock)).toHaveLength(1);
    const [, init] = batchCalls(fetchMock)[0];
    expect(init).toMatchObject({
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ dirs: ['/work/first', '/work/second'] }),
    });
  });

  it('retains successful workspaces beside an explicit per-workspace failure', async () => {
    const twoProjects: Project[] = [
      projects[0],
      { name: 'captain', dir: '/work/captain', repos: ['flanksource/captain'] },
    ];
    const successfulTodo: TodoItem = {
      ref: 'todo-1', id: 'todo-1', cwd: '/work/gavel', title: 'Successful result',
      status: 'pending', priority: 'medium',
    };
    const workspaceError = { code: 'load_failed', message: 'captain database unavailable' };
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({ results: [
      { dir: '/work/captain', items: null, error: workspaceError },
      { dir: '/work/gavel', counts: { ...counts, total: 1, open: 1, pending: 1 }, items: [successfulTodo] },
    ] })));

    const { result } = renderHook(() => useWorkspaceTodos(twoProjects), { wrapper: queryTestWrapper() });

    await waitFor(() => expect(result.current.error).toBe(workspaceError.message));
    expect(result.current.byDir['/work/gavel']?.items).toEqual([successfulTodo]);
    expect(result.current.byDir['/work/captain']?.items).toEqual([]);
    expect(result.current.errorsByDir['/work/captain']).toEqual(workspaceError);
  });

  it('refetches the active normalized batch query on explicit refresh', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, _init?: RequestInit) => (
      isTagRequest(input)
        ? jsonResponse({ definitions: [] })
        : jsonResponse({ results: [{ dir: '/work/gavel', counts, items: [] }] })
    ));
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useWorkspaceTodos(projects), { wrapper: queryTestWrapper() });
    await waitFor(() => expect(batchCalls(fetchMock)).toHaveLength(1));

    act(() => result.current.refresh());

    await waitFor(() => expect(batchCalls(fetchMock)).toHaveLength(2));
    expect(batchCalls(fetchMock).map(([input]) => String(input))).toEqual([
      '/api/todos/batch',
      '/api/todos/batch',
    ]);
  });

  it('aborts the batch request when its last observer unmounts', async () => {
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

  it('polls an active detail until the provider session and terminal projection arrive', async () => {
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
      if (url === '/api/todos/batch') {
        return Promise.resolve(jsonResponse({ results: [{ dir: '/work/gavel', counts, items: [] }] }));
      }
      throw new Error(`unexpected request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useWorkspaceTodos(projects, 'run-1'), { wrapper: queryTestWrapper() });
    await waitFor(() => expect(result.current.detail?.status).toBe('in_progress'));
    expect(result.current.detail?.sessionId).toBeUndefined();

    await waitFor(() => expect(result.current.detail?.sessionId).toBe('codex-session-1'), { timeout: 1_500 });
    expect(detailURLs).toEqual([
      '/api/todos/item?ref=run-1',
      '/api/todos/item?dir=%2Fwork%2Fgavel&ref=run-1',
    ]);
    expect(result.current.detail?.status).toBe('in_progress');

    await waitFor(() => expect(result.current.detail?.status).toBe('review'), { timeout: 1_500 });
    expect(result.current.detail?.hasPlan).toBe(true);
  });

  // Layout is a view preference like density and grouping: owned by the hook so
  // the navbar toggle and the body slots read one value, and written through to
  // storage so the choice survives a reload.
  it('defaults the layout to split and persists a change', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({ results: [] })));

    const { result } = renderHook(() => useWorkspaceTodos([]), { wrapper: queryTestWrapper() });
    expect(result.current.layout).toBe('split');

    act(() => result.current.setLayout('full'));
    await waitFor(() => expect(result.current.layout).toBe('full'));
    expect(localStorage.getItem('gavel.pr-ui.todoLayout.v1')).toBe('full');
  });
});
