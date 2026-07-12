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
  todoProvider: 'grite',
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
  localStorage.clear();
});

afterEach(() => {
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
      provider: 'db',
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
      provider: 'db',
    });
    expect(result.current.loadingList).toBe(true);
    expect(onNavigate).not.toHaveBeenCalled();
    expect(fetchMock.mock.calls.map(([input]) => String(input)).filter(url => url.startsWith('/api/todos/item?'))).toEqual([
      '/api/todos/item?ref=legacy%2Fref',
    ]);

    const list: TodoListResponse = { provider: 'db', dir: '/work/gavel', counts, items: [] };
    await act(async () => resolveList(jsonResponse(list)));
    await waitFor(() => expect(result.current.loadingList).toBe(false));
    unmount();
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
      return Promise.resolve(jsonResponse({ provider: 'db', dir: '/work/gavel', counts, items: [] }));
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
});
