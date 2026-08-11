import { useMutation, useQueryClient, type QueryClient } from '@tanstack/react-query';
import type { TodoItem, TodoListResponse } from '../../types';
import { todoQuery } from './format';
import { setTodoQueryData, todoQueryKeys } from './todoQueries';
import { workspaceTodoBatchKeys } from './workspaceTodoQueries';

export interface TodoMutationBody {
  body: BodyInit;
  headers?: HeadersInit;
}

export interface TodoUpdateMutationBody extends TodoMutationBody {
  ref: string;
}

export interface CreateTodoResponse {
  todo: TodoItem;
}

interface WorkspaceTodoBatchCache {
  byDir?: Record<string, TodoListResponse>;
}

function workspaceLabel(dir: string) {
  return dir.trim() || 'the default workspace';
}

export async function todoMutationJSON<T>(url: string, init: RequestInit, context: string): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, init);
  } catch (cause) {
    if (init.signal?.aborted) throw cause;
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
  if (!response.ok) throw await todoMutationResponseError(response, context);
  if (response.status === 204) return undefined as T;
  try {
    return await response.json() as T;
  } catch (cause) {
    throw new Error(`${context}: invalid JSON response`, { cause });
  }
}

/** Carries the HTTP status so a caller can react to a specific rejection —
 *  a 409 already-linked push offers to rewrite the issue instead. */
export class TodoMutationError extends Error {
  constructor(message: string, readonly status: number) {
    super(message);
    this.name = 'TodoMutationError';
  }
}

async function todoMutationResponseError(response: Response, context: string) {
  const jsonResponse = typeof response.clone === 'function' ? response.clone() : response;
  const payload = await jsonResponse.json().catch(() => null) as { error?: unknown } | null;
  if (typeof payload?.error === 'string' && payload.error.trim()) {
    return new TodoMutationError(`${context}: ${payload.error.trim()}`, response.status);
  }
  const detail = typeof response.text === 'function' ? (await response.text().catch(() => '')).trim() : '';
  return new TodoMutationError(detail ? `${context}: ${detail}` : `${context} (${response.status})`, response.status);
}

export async function invalidateTodoCollections(client: QueryClient, dir: string) {
  await Promise.all([
    client.invalidateQueries({ queryKey: todoQueryKeys.list(dir), exact: true }),
    client.invalidateQueries({ queryKey: workspaceTodoBatchKeys.all }),
  ]);
}

export async function setTodoCaches(client: QueryClient, dir: string, todo: TodoItem) {
  setTodoQueryData(client, dir, todo);
  await invalidateTodoCollections(client, dir);
}

export async function invalidateTodoCaches(client: QueryClient, dir: string, ref: string) {
  await Promise.all([
    client.invalidateQueries({ queryKey: todoQueryKeys.item(dir, ref), exact: true }),
    invalidateTodoCollections(client, dir),
  ]);
}

export async function invalidateTodoWorkflowCaches(client: QueryClient, dir: string, todo: TodoItem) {
  const tasks: Promise<unknown>[] = [
    setTodoCaches(client, dir, todo),
    client.invalidateQueries({ queryKey: todoQueryKeys.plan(dir, todo.ref), exact: true }),
    client.invalidateQueries({
      predicate: query => {
        if (query.queryKey[0] !== 'todos' || query.queryKey[1] !== 'session' || query.queryKey[2] !== 'detail') return false;
        const scope = query.queryKey[3];
        return typeof scope === 'object' && scope !== null && 'dir' in scope && 'ref' in scope
          && scope.dir === dir.trim() && scope.ref === todo.ref;
      },
    }),
  ];
  if (todo.sessionId) tasks.push(client.invalidateQueries({ queryKey: todoQueryKeys.sessionStats(dir, todo.sessionId), exact: true }));
  await Promise.all(tasks);
}

export function optimisticallySetTodoCaches(client: QueryClient, dir: string, todo: TodoItem) {
  const itemKey = todoQueryKeys.item(dir, todo.ref);
  const listKey = todoQueryKeys.list(dir);
  const previousItem = client.getQueryData<TodoItem>(itemKey);
  const previousList = client.getQueryData<TodoListResponse>(listKey);
  const previousBatches = client.getQueriesData({ queryKey: workspaceTodoBatchKeys.all });
  setTodoQueryData(client, dir, todo);
  client.setQueriesData<WorkspaceTodoBatchCache>({ queryKey: workspaceTodoBatchKeys.all }, current => {
    const workspace = current?.byDir?.[dir.trim()];
    if (!current || !workspace?.items.some(item => item.ref === todo.ref)) return current;
    return {
      ...current,
      byDir: {
        ...current.byDir,
        [dir.trim()]: { ...workspace, items: workspace.items.map(item => item.ref === todo.ref ? todo : item) },
      },
    };
  });
  return () => {
    if (previousItem === undefined) client.removeQueries({ queryKey: itemKey, exact: true });
    else client.setQueryData(itemKey, previousItem);
    if (previousList === undefined) client.removeQueries({ queryKey: listKey, exact: true });
    else client.setQueryData(listKey, previousList);
    for (const [queryKey, value] of previousBatches) client.setQueryData(queryKey, value);
  };
}

export async function removeTodoCaches(client: QueryClient, dir: string, ref: string) {
  client.removeQueries({ queryKey: todoQueryKeys.item(dir, ref), exact: true });
  await invalidateTodoCollections(client, dir);
}

export function useCreateTodoMutation(dir: string, context = `Failed to create todo in ${workspaceLabel(dir)}`) {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'create', { dir: dir.trim() }],
    mutationFn: ({ body, headers }: TodoMutationBody) => todoMutationJSON<CreateTodoResponse>(
      `/api/todos/new?${todoQuery(dir)}`,
      { method: 'POST', body, headers },
      context,
    ),
    onSuccess: async ({ todo }) => {
      if (!todo?.ref) throw new Error(`${context}: response did not include the created todo`);
      client.setQueryData(todoQueryKeys.item(dir, todo.ref), todo);
      await invalidateTodoCollections(client, dir);
    },
  });
}

export function useUpdateTodoMutation(dir: string, context = 'Failed to update todo') {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'update', { dir: dir.trim() }],
    mutationFn: ({ body, headers }: TodoUpdateMutationBody) => todoMutationJSON<TodoItem>(
      `/api/todos/item?${todoQuery(dir)}`,
      { method: 'PATCH', body, headers },
      context,
    ),
    onSuccess: todo => setTodoCaches(client, dir, todo),
  });
}

export function useDeleteTodoMutation(dir: string) {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'archive', { dir: dir.trim() }],
    mutationFn: (ref: string) => {
      const params = new URLSearchParams(todoQuery(dir));
      params.set('ref', ref);
      return todoMutationJSON<void>(`/api/todos/item?${params.toString()}`, { method: 'DELETE' }, `Failed to archive todo ${ref}`);
    },
    onSuccess: (_result, ref) => removeTodoCaches(client, dir, ref),
  });
}

// The response of POST /api/todos/verification/run. Only the parts a caller
// acts on are modelled: recorded results are read back from the attempt list,
// never from this payload.
export interface VerificationRunResponse {
  todo?: TodoItem;
  verification?: { allPassed?: boolean; error?: string };
  error?: string;
}

/**
 * Run the definition-of-done fixture.
 *
 * Shared by the Verification tab and the header's Verify phase so both post the
 * same body and invalidate the same caches — a check started from the header
 * has to land in the attempt list the tab reads, or the tab would show stale
 * evidence for a run that just happened.
 */
export function useTodoVerificationRun(dir: string, ref: string) {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'verification', 'run', { dir: dir.trim(), ref }],
    mutationFn: ({ ref: target, spec }: { ref: string; spec: unknown }) => todoMutationJSON<VerificationRunResponse>(
      `/api/todos/verification/run?${todoQuery(dir)}`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: target, spec }),
      },
      'Verification run failed',
    ),
    onSuccess: async (data) => {
      if (data.todo) await setTodoCaches(client, dir, data.todo);
      await client.invalidateQueries({ queryKey: todoQueryKeys.sessionDetail(dir, ref, undefined, true) });
    },
  });
}

/**
 * Interrupt one in-flight attempt.
 *
 * Keyed on the attempt's promptRunId rather than the todo, because a todo can
 * have several attempts and only the running one is stoppable.
 */
export function useTodoSessionStop(dir: string, ref: string, sessionId?: string) {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'session', 'stop', { dir: dir.trim(), ref }],
    mutationFn: (promptRunId: string) => todoMutationJSON<{ status: string; promptRunId: string }>(
      `/api/todos/session/stop?${todoQuery(dir)}`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref, promptRunId }),
      },
      'Could not stop the attempt',
    ),
    onSuccess: async () => {
      await Promise.all([
        invalidateTodoCollections(client, dir),
        client.invalidateQueries({ queryKey: todoQueryKeys.sessionStats(dir, sessionId ?? '') }),
        client.invalidateQueries({ queryKey: todoQueryKeys.sessionDetail(dir, ref, sessionId, false) }),
        client.invalidateQueries({ queryKey: todoQueryKeys.sessionDetail(dir, ref, undefined, true) }),
      ]);
    },
  });
}

export interface GithubPushResponse {
  todo: TodoItem;
  repo: string;
  number: number;
  url: string;
  alias: string;
  /** True when the push rewrote an existing issue instead of opening one. */
  updated: boolean;
}

export interface GithubPushRequest {
  ref: string;
  /** Opens a second issue for an already-linked todo. */
  force?: boolean;
  /** Rewrites the issue the todo is already linked to. */
  update?: boolean;
  /** Rewrites and links one specific issue: 123, owner/repo#123, or its URL. */
  issue?: string;
}

export function useGithubPushTodoMutation(dir: string) {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'github', 'push', { dir: dir.trim() }],
    mutationFn: ({ ref, force, update, issue }: GithubPushRequest) => todoMutationJSON<GithubPushResponse>(
      '/api/todos/github',
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref, dir, force, update, issue }),
      },
      update || issue ? `Failed to update todo ${ref} on GitHub` : `Failed to push todo ${ref} to GitHub`,
    ),
    onSuccess: async ({ todo }) => {
      await setTodoCaches(client, dir, todo);
    },
  });
}

export function useTransferTodoMutation() {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'transfer'],
    mutationFn: ({ ref, fromDir, toDir }: { ref: string; fromDir: string; toDir: string }) => todoMutationJSON<CreateTodoResponse>(
      '/api/todos/transfer',
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref, fromDir, toDir }),
      },
      `Failed to move todo ${ref} to ${workspaceLabel(toDir)}`,
    ),
    onSuccess: async ({ todo }, { ref, fromDir, toDir }) => {
      client.removeQueries({ queryKey: todoQueryKeys.item(fromDir, ref), exact: true });
      client.setQueryData(todoQueryKeys.item(toDir, todo.ref), todo);
      await Promise.all([invalidateTodoCollections(client, fromDir), invalidateTodoCollections(client, toDir)]);
    },
  });
}
