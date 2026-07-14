import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Project, TodoItem, TodoListResponse } from '../../types';
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

function jsonResponse(body: unknown, ok = true, status = ok ? 200 : 500): Response {
  return { ok, status, json: async () => body } as Response;
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
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('useWorkspaceTodos deep links', () => {
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
      if (url.startsWith('/api/todos?')) return pendingList;
      throw new Error(`unexpected request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    const onNavigate = vi.fn();

    const { result, unmount } = renderHook(() => useWorkspaceTodos(projects, routeRef, onNavigate));

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
      if (url.startsWith('/api/todos?')) {
        return Promise.resolve(jsonResponse({ dir: '/work/gavel', counts, items: [base] }));
      }
      if (url === `/api/todos/item?ref=${todoID}`) return Promise.resolve(jsonResponse(base));
      if (url === `/api/todos/item?ref=${sessionID}`) return Promise.resolve(jsonResponse(historical));
      throw new Error(`stale or unexpected detail request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    const onNavigate = vi.fn();

    const { result, rerender } = renderHook(
      ({ selectedId }: { selectedId: string }) => useWorkspaceTodos(projects, selectedId, onNavigate),
      { initialProps: { selectedId: todoID } },
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
      return Promise.resolve(jsonResponse({ dir: '/work/gavel', counts, items: [] }));
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useWorkspaceTodos(projects, status === 404 ? 'missing123' : 'abc123'));

    await waitFor(() => expect(result.current.detailError).toBe(message));
    expect(result.current.detail).toBeNull();
    expect(result.current.selected).toBeNull();
  });

  it('keeps a database list failure visible while retaining an empty workspace group', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({ error: 'database unavailable: connection refused' }, false, 503)));

    const { result } = renderHook(() => useWorkspaceTodos(projects));

    await waitFor(() => expect(result.current.error).toBe('database unavailable: connection refused'));
    expect(result.current.byDir['/work/gavel']?.items).toEqual([]);
  });

  it('polls an active detail until the provider session and terminal projection arrive', async () => {
    vi.useFakeTimers();
    const active: TodoItem = {
      ref: 'run-1', id: 'run-1', cwd: '/work/gavel', title: 'Live run',
      status: 'in_progress', priority: 'medium',
    };
    const attached: TodoItem = { ...active, sessionId: 'codex-session-1' };
    const completed: TodoItem = { ...attached, status: 'review', hasPlan: true };
    const details = [active, attached, completed];
    let detailRead = 0;
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith('/api/todos/item?')) {
        const body = details[Math.min(detailRead, details.length - 1)];
        detailRead++;
        return Promise.resolve(jsonResponse(body));
      }
      if (url.startsWith('/api/todos?')) {
        return Promise.resolve(jsonResponse({ dir: '/work/gavel', counts, items: [] }));
      }
      throw new Error(`unexpected request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useWorkspaceTodos(projects, 'run-1'));
    await act(async () => { await Promise.resolve(); await Promise.resolve(); });
    expect(result.current.detail?.status).toBe('in_progress');
    expect(result.current.detail?.sessionId).toBeUndefined();

    await act(async () => { await vi.advanceTimersByTimeAsync(1000); });
    expect(result.current.detail?.sessionId).toBe('codex-session-1');
    expect(result.current.detail?.status).toBe('in_progress');

    await act(async () => { await vi.advanceTimersByTimeAsync(1000); });
    expect(result.current.detail?.status).toBe('review');
    expect(result.current.detail?.hasPlan).toBe(true);
  });
});
