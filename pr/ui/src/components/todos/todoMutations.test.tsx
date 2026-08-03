import type { ReactNode } from 'react';
import { act, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem, TodoListResponse } from '../../types';
import {
  optimisticallySetTodoCaches,
  todoMutationJSON,
  useCreateTodoMutation,
  useUpdateTodoMutation,
} from './todoMutations';
import { todoQueryKeys } from './todoQueries';

function todo(overrides: Partial<TodoItem> = {}): TodoItem {
  return {
    ref: 'todo-1',
    title: 'Original title',
    status: 'pending',
    priority: 'medium',
    ...overrides,
  };
}

function list(items: TodoItem[]): TodoListResponse {
  return {
    counts: {
      total: items.length,
      open: items.length,
      draft: 0,
      pending: items.length,
      inProgress: 0,
      review: 0,
      ask: 0,
      completed: 0,
      failed: 0,
      unverified: 0,
      verified: 0,
      skipped: 0,
    },
    items,
  };
}

function client() {
  return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
}

function wrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

afterEach(() => vi.unstubAllGlobals());

describe('todo mutations', () => {
  it('stores a created todo and invalidates its direct list plus the shared workspace batches', async () => {
    const queryClient = client();
    const created = todo({ ref: 'created-1', title: 'Created' });
    queryClient.setQueryData(todoQueryKeys.list('/repo'), list([]));
    queryClient.setQueryData(todoQueryKeys.list('/other'), list([]));
    queryClient.setQueryData(['todos', 'batch', { dirs: ['/other', '/repo'] }], { byDir: {} });
    queryClient.setQueryData(['todos', 'batch', { dirs: ['/other'] }], { byDir: {} });
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ todo: created }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })));

    const hook = renderHook(() => useCreateTodoMutation('/repo'), { wrapper: wrapper(queryClient) });
    await act(() => hook.result.current.mutateAsync({ body: JSON.stringify({ title: 'Created' }) }));

    expect(queryClient.getQueryData(todoQueryKeys.item('/repo', created.ref))).toEqual(created);
    expect(queryClient.getQueryState(todoQueryKeys.list('/repo'))?.isInvalidated).toBe(true);
    expect(queryClient.getQueryState(todoQueryKeys.list('/other'))?.isInvalidated).toBe(false);
    expect(queryClient.getQueryState(['todos', 'batch', { dirs: ['/other', '/repo'] }])?.isInvalidated).toBe(true);
    expect(queryClient.getQueryState(['todos', 'batch', { dirs: ['/other'] }])?.isInvalidated).toBe(true);
  });

  it('updates exact item/list caches and leaves another workspace unchanged', async () => {
    const queryClient = client();
    const original = todo();
    const updated = todo({ title: 'Server title' });
    queryClient.setQueryData(todoQueryKeys.item('/repo', original.ref), original);
    queryClient.setQueryData(todoQueryKeys.list('/repo'), list([original]));
    queryClient.setQueryData(['todos', 'batch', { dirs: ['/repo'] }], { byDir: { '/repo': list([original]) } });
    queryClient.setQueryData(todoQueryKeys.item('/other', original.ref), original);
    queryClient.setQueryData(todoQueryKeys.list('/other'), list([original]));
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(updated), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })));

    const hook = renderHook(() => useUpdateTodoMutation('/repo'), { wrapper: wrapper(queryClient) });
    await act(() => hook.result.current.mutateAsync({
      ref: original.ref,
      body: JSON.stringify({ ref: original.ref, title: updated.title }),
    }));

    expect(queryClient.getQueryData(todoQueryKeys.item('/repo', original.ref))).toEqual(updated);
    expect(queryClient.getQueryData<TodoListResponse>(todoQueryKeys.list('/repo'))?.items).toEqual([updated]);
    expect(queryClient.getQueryData(todoQueryKeys.item('/other', original.ref))).toEqual(original);
    expect(queryClient.getQueryData<TodoListResponse>(todoQueryKeys.list('/other'))?.items).toEqual([original]);
  });

  it('restores exact item/list caches when an optimistic mutation is rolled back', () => {
    const queryClient = client();
    const original = todo();
    queryClient.setQueryData(todoQueryKeys.item('/repo', original.ref), original);
    queryClient.setQueryData(todoQueryKeys.list('/repo'), list([original]));
    queryClient.setQueryData(['todos', 'batch', { dirs: ['/repo'] }], { byDir: { '/repo': list([original]) } });

    const rollback = optimisticallySetTodoCaches(queryClient, '/repo', { ...original, title: 'Optimistic' });
    expect(queryClient.getQueryData<TodoItem>(todoQueryKeys.item('/repo', original.ref))?.title).toBe('Optimistic');
    expect(queryClient.getQueryData<TodoListResponse>(todoQueryKeys.list('/repo'))?.items[0]?.title).toBe('Optimistic');
    expect(queryClient.getQueryData<{ byDir: Record<string, TodoListResponse> }>(['todos', 'batch', { dirs: ['/repo'] }])?.byDir['/repo'].items[0]?.title).toBe('Optimistic');

    rollback();
    expect(queryClient.getQueryData(todoQueryKeys.item('/repo', original.ref))).toEqual(original);
    expect(queryClient.getQueryData<TodoListResponse>(todoQueryKeys.list('/repo'))?.items).toEqual([original]);
    expect(queryClient.getQueryData<{ byDir: Record<string, TodoListResponse> }>(['todos', 'batch', { dirs: ['/repo'] }])?.byDir['/repo'].items).toEqual([original]);
  });

  it('adds context to JSON server, non-JSON, and network failures', async () => {
    vi.stubGlobal('fetch', vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: 'provider unavailable' }), { status: 503 }))
      .mockResolvedValueOnce(new Response('proxy exploded', { status: 502 }))
      .mockRejectedValueOnce(new Error('connection refused')));

    await expect(todoMutationJSON('/api/todos/new', { method: 'POST' }, 'Failed to create todo'))
      .rejects.toThrow('Failed to create todo: provider unavailable');
    await expect(todoMutationJSON('/api/todos/new', { method: 'POST' }, 'Failed to create todo'))
      .rejects.toThrow('Failed to create todo: proxy exploded');
    await expect(todoMutationJSON('/api/todos/new', { method: 'POST' }, 'Failed to create todo'))
      .rejects.toThrow('Failed to create todo: connection refused');
  });
});
