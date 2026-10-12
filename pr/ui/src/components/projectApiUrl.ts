/**
 * URL of a per-project endpoint (`/api/projects/{name}/{resource}`) with its
 * query parameters, in order, followed by `worktree` when the request is scoped
 * to a linked worktree. The main checkout passes no worktree and gets the
 * unscoped URL.
 */
export function projectApiUrl({ projectName, resource, query = [], worktree }: {
  projectName: string;
  resource: string;
  query?: readonly (readonly [string, string])[];
  worktree?: string;
}): string {
  const pairs = worktree ? [...query, ['worktree', worktree] as const] : query;
  const search = pairs.map(([key, value]) => `${key}=${encodeURIComponent(value)}`).join('&');
  return `/api/projects/${encodeURIComponent(projectName)}/${resource}${search ? `?${search}` : ''}`;
}
