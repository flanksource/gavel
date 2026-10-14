import { queryOptions, useMutation, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { fetchJSON, mutationJSON, queryKeys } from '../../query';
import type { PRLinkedTodo, PRSearchItem, TodoLink, TodoLinkRelation, TodoLinksResponse, TodoPRLinkItem, TodoPRLinkRelation } from '../../types';

function requireList<T>(value: unknown, context: string): T[] {
  if (!Array.isArray(value)) throw new Error(`${context}: expected a list, got ${JSON.stringify(value)?.slice(0, 120)}`);
  return value as T[];
}

function params(entries: Record<string, string>) {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(entries)) if (value.trim()) search.set(key, value.trim());
  return search.toString();
}

export function todoLinksQueryOptions(dir: string, ref: string) {
  return queryOptions({
    queryKey: queryKeys.todoLinks(dir, ref),
    queryFn: async ({ signal }): Promise<TodoLink[]> => {
      const context = `Failed to load links for todo ${ref}`;
      const body = await fetchJSON<TodoLinksResponse>({ url: `/api/todos/links?${params({ dir, ref })}`, signal, context });
      return requireList<TodoLink>(body?.links, context);
    },
    staleTime: 5_000,
  });
}

export function todoPRLinksQueryOptions(dir: string, ref: string) {
  return queryOptions({
    queryKey: queryKeys.todoPRLinks(dir, ref),
    queryFn: async ({ signal }): Promise<TodoPRLinkItem[]> => {
      const context = `Failed to load pull request links for todo ${ref}`;
      return requireList<TodoPRLinkItem>(await fetchJSON<unknown>({ url: `/api/todos/pr-links?${params({ dir, id: ref })}`, signal, context }), context);
    },
    staleTime: 5_000,
  });
}

export function prLinkedTodosQueryOptions(url: string) {
  return queryOptions({
    queryKey: queryKeys.prLinkedTodos(url),
    queryFn: async ({ signal }): Promise<PRLinkedTodo[]> => {
      const context = `Failed to load todos linked to ${url}`;
      return requireList<PRLinkedTodo>(await fetchJSON<unknown>({ url: `/api/prs/todos?${params({ url })}`, signal, context }), context);
    },
    staleTime: 5_000,
  });
}

// An empty query is a valid "most recently updated" listing, so it is not skipped.
export function prSearchQueryOptions({ query, repo = '', limit = 20 }: { query: string; repo?: string; limit?: number }) {
  return queryOptions({
    queryKey: queryKeys.prSearch(query, repo),
    queryFn: async ({ signal }): Promise<PRSearchItem[]> => {
      const context = 'Failed to search pull requests';
      const search = params({ q: query, repo, limit: String(limit) });
      return requireList<PRSearchItem>(await fetchJSON<unknown>({ url: `/api/prs/search?${search}`, signal, context }), context);
    },
    staleTime: 15_000,
  });
}

function invalidatePRLinks(client: QueryClient, dir: string, ref: string, url: string) {
  return Promise.all([
    client.invalidateQueries({ queryKey: queryKeys.todoPRLinks(dir, ref), exact: true }),
    client.invalidateQueries({ queryKey: queryKeys.prLinkedTodos(url), exact: true }),
  ]);
}

export function useLinkTodoPR({ dir, todoRef }: { dir: string; todoRef: string }) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ url, relation }: { url: string; relation: TodoPRLinkRelation }) => mutationJSON<TodoPRLinkItem>({
      url: '/api/todos/pr-links',
      method: 'POST',
      body: { dir, id: todoRef, url, relation },
      context: `Failed to link ${url} to todo ${todoRef}`,
    }),
    onSuccess: (_link, { url }) => invalidatePRLinks(client, dir, todoRef, url),
  });
}

export function useUnlinkTodoPR({ dir, todoRef }: { dir: string; todoRef: string }) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ url }: { url: string }) => mutationJSON<void>({
      url: `/api/todos/pr-links?${params({ dir, id: todoRef, url })}`,
      method: 'DELETE',
      context: `Failed to unlink ${url} from todo ${todoRef}`,
    }),
    onSuccess: (_result, { url }) => invalidatePRLinks(client, dir, todoRef, url),
  });
}

// Unlinking answers with the remaining links, which replace the cached list.
export function useUnlinkTodo({ dir, todoRef }: { dir: string; todoRef: string }) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ target, relation }: { target: string; relation: TodoLinkRelation }) => mutationJSON<TodoLinksResponse>({
      url: '/api/todos/links',
      method: 'DELETE',
      body: { dir, ref: todoRef, target, relation },
      context: `Failed to unlink todo ${target} from ${todoRef}`,
    }),
    onSuccess: () => client.invalidateQueries({ queryKey: queryKeys.todoLinks(dir, todoRef), exact: true }),
  });
}
