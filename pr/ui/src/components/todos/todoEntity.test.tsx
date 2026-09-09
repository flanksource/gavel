import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { useTodoBulkActionMutation, type TodoBulkAction, type TodoBulkResult } from './todoEntity';
import { todoQueryKeys } from './todoQueries';
import { workspaceTodoBatchKeys } from './workspaceTodoQueries';

const priorityAction: TodoBulkAction = {
  name: 'priority',
  short: 'Set the severity of many TODOs',
  method: 'POST',
  path: '/api/v1/todo/{id}/priority',
  tool_hints: { icon: 'flag', group: 'Status' },
};

const deleteAction: TodoBulkAction = {
  name: 'delete',
  short: 'Delete many TODOs',
  method: 'DELETE',
  path: '/api/v1/todo/{id}',
  tool_hints: { icon: 'trash', group: 'Danger', destructiveHint: true },
};

const ALPHA = '/repos/alpha';
const BETA = '/repos/beta';

/** Everything a bulk action can leave behind: the workspace listing the table
 *  reads, the per-workspace list, and one detail cache per todo. */
function seedCaches(client: QueryClient) {
  client.setQueryData(workspaceTodoBatchKeys.list([ALPHA, BETA]), { byDir: {} });
  client.setQueryData(todoQueryKeys.list(ALPHA), { todos: [] });
  client.setQueryData(todoQueryKeys.item(ALPHA, 'todo-1'), { ref: 'todo-1', priority: 'low' });
  client.setQueryData(todoQueryKeys.globalItem('todo-1'), { ref: 'todo-1', priority: 'low' });
}

function invalidated(client: QueryClient, key: readonly unknown[]) {
  return client.getQueryState(key)?.isInvalidated === true;
}

describe('useTodoBulkActionMutation cache invalidation', () => {
  let client: QueryClient;
  let response: TodoBulkResult;

  function run(action: TodoBulkAction, refs: string[]) {
    const { result } = renderHook(() => useTodoBulkActionMutation(), {
      wrapper: ({ children }: { children: ReactNode }) =>
        <QueryClientProvider client={client}>{children}</QueryClientProvider>,
    });
    result.current.mutate({ action, refs });
    return waitFor(() => expect(result.current.isSuccess).toBe(true));
  }

  beforeEach(() => {
    client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    seedCaches(client);
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => response,
    } as Response)));
  });

  afterEach(() => vi.unstubAllGlobals());

  it('refreshes the workspace listing after an applied batch', async () => {
    response = {
      action: 'priority',
      applied: 1,
      failed: 0,
      results: [{ ref: 'todo-1', dir: ALPHA, status: 'high' }],
    };

    await run(priorityAction, ['todo-1']);

    expect(invalidated(client, workspaceTodoBatchKeys.list([ALPHA, BETA]))).toBe(true);
    expect(invalidated(client, todoQueryKeys.list(ALPHA))).toBe(true);
  });

  // The listing is what the table renders, and a batch that reported no
  // workspace can still have written — so the refresh must not be derived from
  // the per-item outcomes.
  it('refreshes the workspace listing even when no item reported its workspace', async () => {
    response = {
      action: 'priority',
      applied: 1,
      failed: 0,
      results: [{ ref: 'todo-1' }],
    };

    await run(priorityAction, ['todo-1']);

    expect(invalidated(client, workspaceTodoBatchKeys.list([ALPHA, BETA]))).toBe(true);
  });

  // refetchOnWindowFocus is off app-wide, so a detail pane open on a
  // bulk-edited todo renders its pre-action severity until this invalidates.
  it("refreshes each edited todo's own detail cache", async () => {
    response = {
      action: 'priority',
      applied: 1,
      failed: 0,
      results: [{ ref: 'todo-1', dir: ALPHA, status: 'high' }],
    };

    await run(priorityAction, ['todo-1']);

    expect(invalidated(client, todoQueryKeys.item(ALPHA, 'todo-1'))).toBe(true);
    expect(invalidated(client, todoQueryKeys.globalItem('todo-1'))).toBe(true);
  });

  // A ref that resolved to nothing was never written, and invalidating its
  // detail would refetch a todo the batch just told us does not exist.
  it('leaves the caches of a failed item alone', async () => {
    response = {
      action: 'priority',
      applied: 0,
      failed: 1,
      results: [{ ref: 'todo-1', dir: ALPHA, error: 'not found' }],
    };

    await run(priorityAction, ['todo-1']);

    expect(invalidated(client, todoQueryKeys.item(ALPHA, 'todo-1'))).toBe(false);
  });

  // Invalidating a deleted todo's detail refetches a 404. The single-todo
  // delete drops the cache instead, and a batch delete is the same operation.
  it('drops the caches of todos a destructive action removed', async () => {
    response = {
      action: 'delete',
      applied: 1,
      failed: 0,
      results: [{ ref: 'todo-1', dir: ALPHA, status: 'deleted' }],
    };

    await run(deleteAction, ['todo-1']);

    expect(client.getQueryData(todoQueryKeys.item(ALPHA, 'todo-1'))).toBeUndefined();
    expect(client.getQueryData(todoQueryKeys.globalItem('todo-1'))).toBeUndefined();
    expect(invalidated(client, workspaceTodoBatchKeys.list([ALPHA, BETA]))).toBe(true);
  });
});
