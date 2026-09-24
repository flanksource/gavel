import { queryOptions } from '@tanstack/react-query';
import type { SessionStats, TodoItem, TodoListResponse, TodoSessionDetailResponse } from '../../types';
import type { QueryClient } from '@tanstack/react-query';
import { responseError } from '../oneShotQueries';
import { todoQuery } from './format';
import { sessionResponseError } from './SessionErrorDetails';

function queryDir(dir: string) {
  return dir.trim();
}

function workspaceParams(dir: string) {
  const params = new URLSearchParams();
  if (queryDir(dir)) params.set('dir', queryDir(dir));
  return params;
}

async function sessionRequest(url: string, signal: AbortSignal | undefined, context: string) {
  try {
    return await fetch(url, { signal });
  } catch (cause) {
    if (signal?.aborted) throw cause;
    throw new Error(`${context}\n${cause instanceof Error ? cause.message : String(cause)}`, { cause });
  }
}

export const todoQueryKeys = {
  list: (dir: string) => ['todos', 'list', { dir: queryDir(dir) }] as const,
  item: (dir: string, ref: string) => ['todos', 'item', { dir: queryDir(dir), ref }] as const,
  globalItem: (ref: string) => ['todos', 'item', 'global', { ref }] as const,
  launch: (dir: string, ref: string) => ['todos', 'launch', { dir: queryDir(dir), ref }] as const,
  sessionStats: (dir: string, sessionId: string) =>
    ['todos', 'session', 'stats', { dir: queryDir(dir), sessionId }] as const,
  sessionDetail: (dir: string, ref: string) =>
    ['todos', 'session', 'detail', { dir: queryDir(dir), ref }] as const,
  captainSession: (src: string) => ['captain', 'session', { src }] as const,
  plan: (dir: string, ref: string) =>
    ['todos', 'session', 'plan', { dir: queryDir(dir), ref }] as const,
  commits: (dir: string, ref: string) =>
    ['todos', 'commits', { dir: queryDir(dir), ref }] as const,
  commitFiles: (dir: string, hash: string) =>
    ['todos', 'commits', 'files', { dir: queryDir(dir), hash }] as const,
  commitDiff: (dir: string, hash: string, file: string) =>
    ['todos', 'commits', 'diff', { dir: queryDir(dir), hash, file }] as const,
  verificationSchema: () => ['todos', 'verification', 'schema'] as const,
  cmuxSurface: (dir: string, agent: string | undefined) =>
    ['todos', 'session', 'cmux', { dir: queryDir(dir), agent: agent ?? '' }] as const,
};

async function fetchTodoJSON<T>(url: string, signal: AbortSignal, context: string): Promise<T> {
  const response = await sessionRequest(url, signal, context);
  if (!response.ok) throw await responseError(response, context);
  try {
    return await response.json() as T;
  } catch (cause) {
    throw new Error(`${context}: invalid JSON response`, { cause });
  }
}

export function todoListQueryOptions(dir: string) {
  return queryOptions({
    queryKey: todoQueryKeys.list(dir),
    queryFn: ({ signal }) => fetchTodoJSON<TodoListResponse>(
      `/api/todos?${todoQuery(dir)}`,
      signal,
      `Failed to load todos for ${queryDir(dir) || 'the default workspace'}`,
    ),
    staleTime: 30_000,
  });
}

export function todoItemQueryOptions(
  dir: string,
  ref: string,
  options: { pollWhileActive?: boolean; intervalMs?: number } = {},
) {
  return queryOptions({
    queryKey: todoQueryKeys.item(dir, ref),
    queryFn: ({ signal }) => {
      const params = workspaceParams(dir);
      params.set('ref', ref);
      return fetchTodoJSON<TodoItem>(
        `/api/todos/item?${params.toString()}`,
        signal,
        `Failed to load todo ${ref}`,
      );
    },
    staleTime: 5_000,
    refetchInterval: options.pollWhileActive
      ? query => query.state.data?.status === 'in_progress' ? (options.intervalMs ?? 1_000) : false
      : false,
    refetchIntervalInBackground: false,
  });
}

export function todoGlobalItemQueryOptions(ref: string) {
  return queryOptions({
    queryKey: todoQueryKeys.globalItem(ref),
    queryFn: async ({ signal }) => {
      const params = new URLSearchParams({ ref });
      const todo = await fetchTodoJSON<TodoItem & { dir?: string }>(
        `/api/todos/item?${params.toString()}`,
        signal,
        `Failed to resolve todo ${ref}`,
      );
      if (!todo.ref?.trim()) throw new Error(`Failed to resolve todo ${ref}: response omitted its canonical reference`);
      if (!todo.cwd?.trim() && !todo.dir?.trim()) throw new Error(`Failed to resolve todo ${ref}: response omitted its workspace directory`);
      return todo;
    },
    staleTime: 5_000,
  });
}

export function setTodoQueryData(client: QueryClient, dir: string, todo: TodoItem) {
  client.setQueryData(todoQueryKeys.item(dir, todo.ref), todo);
  client.setQueryData<TodoListResponse>(todoQueryKeys.list(dir), current => {
    if (!current) return current;
    return { ...current, items: current.items.map(item => item.ref === todo.ref ? todo : item) };
  });
}

async function fetchSessionStats(dir: string, sessionId: string, signal: AbortSignal) {
  const params = workspaceParams(dir);
  params.set('sessionId', sessionId);
  const response = await sessionRequest(`/api/todos/session/stats?${params.toString()}`, signal, 'Session stats request failed');
  if (!response.ok) throw new Error(await sessionResponseError(response, 'Session stats request failed'));
  return response.json() as Promise<SessionStats>;
}

// sessionStatsQueryOptions polls one agent session's rolled-up stats. A running
// session polls fast and a finished one stops for good, because its totals are
// final. `expectLive` covers the third case — a session that has produced no log
// or DB row yet. That is only worth retrying while a run is actually expected to
// start; for a settled todo whose session was pruned it would poll forever, and a
// list showing many such todos would poll forever in parallel.
export function sessionStatsQueryOptions({ dir, sessionId, expectLive = true }: {
  dir: string;
  sessionId: string;
  expectLive?: boolean;
}) {
  return queryOptions({
    queryKey: todoQueryKeys.sessionStats(dir, sessionId),
    queryFn: ({ signal }) => fetchSessionStats(dir, sessionId, signal),
    // Settled totals do not move, so a row that only wants the final duration and
    // cost re-reads them rarely instead of on every remount.
    staleTime: expectLive ? 1_500 : 300_000,
    refetchInterval: query => {
      const stats = query.state.data;
      if (stats?.inProgress) return 2_000;
      if (!stats?.found) return expectLive ? 4_000 : false;
      return false;
    },
  });
}

export async function fetchTodoSessionDetail(dir: string, ref: string, signal?: AbortSignal) {
  const params = new URLSearchParams();
  if (dir.trim()) params.set('dir', dir.trim());
  params.set('ref', ref);
  const response = await sessionRequest(`/api/todos/session/detail?${params.toString()}`, signal, 'Session detail request failed');
  if (!response.ok) throw new Error(await sessionResponseError(response, 'Session detail request failed'));
  const body = await response.json().catch((cause: unknown) => {
    throw new Error('Session detail request failed: invalid JSON response', { cause });
  }) as TodoSessionDetailResponse;
  if (!Array.isArray(body?.attempts)) throw new Error(`Session detail request failed: response has no attempts list: ${JSON.stringify(body).slice(0, 200)}`);
  return body;
}

// SETTLED_ATTEMPTS_POLL_MS is how often a todo whose attempts have all finished
// is re-read: a new attempt can still be started elsewhere (the CLI, another tab).
const SETTLED_ATTEMPTS_POLL_MS = 15_000;

function attemptsSettled(detail: TodoSessionDetailResponse | undefined) {
  return !!detail && detail.attempts.every(attempt => attempt.phase === 'finished');
}

export function sessionDetailQueryOptions(dir: string, ref: string, intervalMs: number) {
  return queryOptions({
    queryKey: todoQueryKeys.sessionDetail(dir, ref),
    queryFn: ({ signal }) => fetchTodoSessionDetail(dir, ref, signal),
    staleTime: Math.min(intervalMs, 1_500),
    refetchInterval: query => attemptsSettled(query.state.data) ? Math.max(intervalMs, SETTLED_ATTEMPTS_POLL_MS) : intervalMs,
  });
}
