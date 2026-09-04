import { useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { todoMutationJSON } from './todoMutations';
import { todoQueryKeys } from './todoQueries';
import { workspaceTodoBatchKeys } from './workspaceTodoQueries';

/**
 * The todo action catalog, read from the server rather than declared here.
 *
 * Before this the dashboard hardcoded three bulk actions while the CLI had
 * roughly twenty-five, and the two drifted independently. Now every bulk action
 * is declared once on the Go side and published at `/api/entities`; adding one
 * there makes it appear in the selection toolbar with no change to this file.
 */
export interface TodoActionParamField {
  type: 'string' | 'boolean' | 'integer' | 'number' | 'array';
  description?: string;
  default?: string;
  enum?: string[];
  items?: { type: string };
}

export interface TodoActionParamSchema {
  type: 'object';
  properties: Record<string, TodoActionParamField>;
  required?: string[];
}

export interface TodoActionToolHints {
  icon?: string;
  group?: string;
  title?: string;
  destructiveHint?: boolean;
  readOnlyHint?: boolean;
  defaultPermission?: string;
}

export interface TodoBulkAction {
  name: string;
  short?: string;
  /**
   * The route this action is really served at. Both are published because
   * neither is derivable: the method is inferred from the action's name, so
   * `delete` is a DELETE onto `/{id}` while its siblings are POSTs onto
   * `/{id}/{action}`. Guessing would 404 on exactly the destructive one.
   */
  method: string;
  path: string;
  supports_filter_mode?: boolean;
  tool_hints?: TodoActionToolHints;
  param_schema?: TodoActionParamSchema;
}

interface EntityCatalogEntry {
  name: string;
  bulk_actions?: TodoBulkAction[];
}

export const TODO_ENTITY = 'todo';

async function fetchTodoBulkActions(signal?: AbortSignal): Promise<TodoBulkAction[]> {
  const response = await fetch('/api/entities', { signal });
  if (!response.ok) throw new Error(`Failed to load the todo action catalog (${response.status})`);
  const entities = await response.json() as EntityCatalogEntry[];
  const todo = entities.find(entity => entity.name === TODO_ENTITY);
  // An empty catalog is a real failure, not an empty toolbar: it means the
  // entity did not register, and silently showing no actions would hide that.
  if (!todo) throw new Error('The todo entity is not registered on this server');
  return todo.bulk_actions ?? [];
}

/** The catalog is fixed for the life of the server process, so it is fetched
 *  once and kept rather than refetched per render. */
export function useTodoBulkActions() {
  return useQuery({
    queryKey: ['entities', TODO_ENTITY, 'bulk-actions'],
    queryFn: ({ signal }) => fetchTodoBulkActions(signal),
    staleTime: Infinity,
    gcTime: Infinity,
  });
}

export interface TodoBulkItemResult {
  ref: string;
  dir?: string;
  title?: string;
  sessionId?: string;
  status?: string;
  url?: string;
  error?: string;
}

export interface TodoBulkResult {
  action: string;
  applied: number;
  failed: number;
  results: TodoBulkItemResult[];
  matchedBy?: string;
}

export interface TodoBulkActionRequest {
  action: TodoBulkAction;
  /** Bare todo references. The entity id is the ref alone — never dir+ref,
   *  which could not survive a URL path segment. */
  refs: string[];
  /** The action's own parameters, matching its `param_schema`. */
  params?: Record<string, string | boolean | number | string[] | undefined>;
}

export function todoBulkActionURL({ action, refs, params }: TodoBulkActionRequest): string {
  const path = action.path.replace('{id}', refs.map(ref => encodeURIComponent(ref)).join(','));
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params ?? {})) {
    if (value === undefined || value === '') continue;
    query.set(key, Array.isArray(value) ? value.join(',') : String(value));
  }
  const encoded = query.toString();
  return encoded ? `${path}?${encoded}` : path;
}

export function todoBulkActionLabel(action: TodoBulkAction): string {
  return action.tool_hints?.title
    ?? action.short
    ?? action.name.charAt(0).toUpperCase() + action.name.slice(1);
}

/**
 * The name of the field an action edits — "Status", "Labels" — for a toolbar
 * button that has room for a word and not a sentence.
 *
 * `short` is the catalog's one-line help ("Set the status of many TODOs"),
 * which is right for a tooltip and wrong for a control the reader scans a row
 * of. The action's own name is already the field name, so that is the fallback
 * rather than a second table of labels kept in the UI.
 */
export function todoBulkActionShortLabel(action: TodoBulkAction): string {
  return (
    action.tool_hints?.title
    ?? action.name.charAt(0).toUpperCase() + action.name.slice(1)
  );
}

/**
 * Run one bulk action against a selection.
 *
 * A partial batch answers 200 with per-item outcomes, so `failed > 0` is a
 * success here rather than a thrown error — only a rejected request (an
 * unassignable status, a duplicated ref) throws. That mirrors the server's
 * contract exactly: it never returns an error alongside partial results,
 * because the framework would discard them.
 *
 * Every touched workspace is invalidated, not just the caller's, because a
 * selection grouped by severity or age routinely spans workspaces.
 */
export function useTodoBulkActionMutation() {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'bulk-action'],
    mutationFn: (request: TodoBulkActionRequest) => todoMutationJSON<TodoBulkResult>(
      todoBulkActionURL(request),
      { method: request.action.method },
      `Failed to ${todoBulkActionLabel(request.action).toLowerCase()} ${request.refs.length} todo${request.refs.length === 1 ? '' : 's'}`,
    ),
    onSuccess: (result, request) => invalidateBulkTodoCaches(client, result, request.action),
  });
}

/**
 * Refresh everything one bulk action can have changed.
 *
 * The workspace listing is refreshed unconditionally rather than derived from
 * the per-item outcomes: a batch that wrote can still report no workspace —
 * `dir` is the resolved TODO's, and an item is free to answer without one — and
 * deriving the refresh from it leaves the table rendering rows the batch has
 * already changed.
 *
 * Each applied item's own detail cache goes with it. `refetchOnWindowFocus` is
 * off app-wide, so a detail pane open on a bulk-edited todo would otherwise
 * render its pre-action status until the user reselected it. A failed item is
 * skipped — nothing was written, and refetching a ref the batch just failed to
 * resolve only trades a stale row for an error.
 */
async function invalidateBulkTodoCaches(
  client: QueryClient,
  result: TodoBulkResult,
  action: TodoBulkAction,
) {
  // A removed todo cannot be refetched, so its caches are dropped rather than
  // invalidated — the same thing the single-todo delete does. The catalog's own
  // destructive hint is the test, so this never has to know an action's name.
  const removes = action.tool_hints?.destructiveHint === true;
  const applied = result.results.filter(item => !item.error);
  const tasks: Promise<unknown>[] = [client.invalidateQueries({ queryKey: workspaceTodoBatchKeys.all })];
  for (const dir of new Set(applied.map(item => item.dir ?? ''))) {
    tasks.push(client.invalidateQueries({ queryKey: todoQueryKeys.list(dir), exact: true }));
  }
  for (const item of applied) {
    for (const queryKey of [todoQueryKeys.item(item.dir ?? '', item.ref), todoQueryKeys.globalItem(item.ref)]) {
      if (removes) client.removeQueries({ queryKey, exact: true });
      else tasks.push(client.invalidateQueries({ queryKey, exact: true }));
    }
  }
  await Promise.all(tasks);
}

/** The one-line outcome a toast reports. Kept next to the mutation so the
 *  wording stays with the contract it describes. */
export function todoBulkResultMessage(result: TodoBulkResult, verb = 'Updated'): string {
  const summary = `${verb} ${result.applied} todo${result.applied === 1 ? '' : 's'}`;
  if (result.failed === 0) return summary;
  const failures = result.results
    .filter(item => item.error)
    .map(item => `${item.title || item.ref} (${item.error})`)
    .join(', ');
  return `${summary}; ${result.failed} failed: ${failures}`;
}
