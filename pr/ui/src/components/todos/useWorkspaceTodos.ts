import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import type { Project, TodoCounts, TodoDensity, TodoGroupBy, TodoItem, TodoListResponse, TodoStatus } from '../../types';
import { addCounts, emptyCounts } from './format';
import { loadDensity, saveDensity } from './todoDensity';
import { loadGroupBy, saveGroupBy } from './todoGroup';
import { loadTodoFilters, saveTodoFilters, toggleStatusFilter, type TodoFilters } from './todoFilter';
import type { TodoSort } from './todoSort';
import { loadTodoSort, saveTodoSort } from './todoSort';
import { loadTimeRange, saveTimeRange, type TodoTimeRange } from './todoTimeRange';
import { setTodoQueryData, todoGlobalItemQueryOptions, todoItemQueryOptions, todoQueryKeys } from './todoQueries';
import { useTodoSelection, type SelectedTodo } from './todoSelection';
import { workspaceTodoBatchKeys } from './workspaceTodoQueries';

export type { SelectedTodo };

const activeTodoDetailPollInterval = 1000;

export interface WorkspaceTodoError {
  code: string;
  message: string;
}

interface WorkspaceTodoBatchResult {
  dir: string;
  counts?: TodoCounts;
  items: TodoItem[] | null;
  error?: WorkspaceTodoError;
}

interface WorkspaceTodoBatchResponse {
  results: WorkspaceTodoBatchResult[];
}

interface WorkspaceTodoBatchState {
  byDir: Record<string, TodoListResponse>;
  errorsByDir: Record<string, WorkspaceTodoError>;
  error: string;
}

function normalizeWorkspaceDir(dir: string): string {
  const trimmed = dir.trim();
  if (!trimmed) return '';
  const parts: string[] = [];
  for (const part of trimmed.split('/')) {
    if (!part || part === '.') continue;
    if (part === '..') parts.pop();
    else parts.push(part);
  }
  return `${trimmed.startsWith('/') ? '/' : ''}${parts.join('/')}` || (trimmed.startsWith('/') ? '/' : '.');
}

function workspaceTodoBatchState(dirs: string[], payload: WorkspaceTodoBatchResponse): WorkspaceTodoBatchState {
  if (!Array.isArray(payload.results) || payload.results.length !== dirs.length) {
    throw new Error(`Todo batch response returned ${payload.results?.length ?? 'no'} results for ${dirs.length} workspaces`);
  }
  const byDir: Record<string, TodoListResponse> = {};
  const errorsByDir: Record<string, WorkspaceTodoError> = {};
  for (let index = 0; index < dirs.length; index++) {
    const dir = dirs[index];
    const result = payload.results[index];
    if (result?.dir !== dir) {
      throw new Error(`Todo batch result ${index} returned workspace ${JSON.stringify(result?.dir)} instead of ${JSON.stringify(dir)}`);
    }
    if (result.error) {
      if (!result.error.code?.trim() || !result.error.message?.trim()) {
        throw new Error(`Todo batch result for ${dir} returned an invalid error`);
      }
      errorsByDir[dir] = result.error;
      byDir[dir] = { dir, counts: emptyCounts, items: [] };
      continue;
    }
    if (!result.counts || !Array.isArray(result.items)) {
      throw new Error(`Todo batch result for ${dir} omitted counts or items`);
    }
    byDir[dir] = { dir, counts: result.counts, items: result.items };
  }
  return {
    byDir,
    errorsByDir,
    error: [...new Set(Object.values(errorsByDir).map(result => result.message))].join('; '),
  };
}

function responseError(payload: unknown, status: number): string {
  if (typeof payload === 'object' && payload !== null && 'error' in payload && typeof payload.error === 'string') {
    return payload.error || `Load failed (HTTP ${status})`;
  }
  return `Load failed (HTTP ${status})`;
}

async function fetchWorkspaceTodoBatch(dirs: string[], signal: AbortSignal): Promise<WorkspaceTodoBatchState> {
  const response = await fetch('/api/todos/batch', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ dirs }),
    signal,
  });
  const payload = await response.json().catch(() => null) as WorkspaceTodoBatchResponse | null;
  if (!response.ok) throw new Error(responseError(payload, response.status));
  if (!payload) throw new Error('Todo batch response was not valid JSON');
  return workspaceTodoBatchState(dirs, payload);
}


// useWorkspaceTodos drives the shared todos data layer: it lists every
// configured workspace's todos in one batch, loads the selected todo's detail,
// aggregates counts, and exposes the create/update/delete callbacks. Both the
// dashboard TodoView and the compact MenubarTodos render off this one hook so
// they hit the same /api/todos endpoints and stay in sync.
//
// selectedId/onNavigate wire the selection to the URL (/todos/{guid}, where the
// guid is the todo ref): the dashboard passes them so a todo is deep-linkable
// and back/forward works; the menubar omits them and keeps purely-local state.
//
// `enabled` gates the list fetch: the dashboard mounts this hook permanently (so
// the Todos chrome can live in the AppShell's body slots) but passes false while
// another tab is active, so the workspaces aren't listed until the Todos tab is
// opened. Cached results survive a tab switch, so reopening is instant.
export function useWorkspaceTodos(
  projects: Project[],
  selectedId = '',
  onNavigate?: (id: string) => void,
  enabled = true,
) {
  // Normalize and de-duplicate configured directories before they reach the
  // strict batch API. Rendering retains projects.json order while the query key
  // and request body use a stable set order.
  const workspaces = useMemo(() => {
    const seen = new Set<string>();
    return projects.flatMap(project => {
      const dir = normalizeWorkspaceDir(project.dir);
      if (!dir || seen.has(dir)) return [];
      seen.add(dir);
      return [{ ...project, dir }];
    });
  }, [projects]);
  const workspaceDirsKey = JSON.stringify(workspaces.map(workspace => workspace.dir).sort());
  const workspaceDirs = useMemo(() => JSON.parse(workspaceDirsKey) as string[], [workspaceDirsKey]);
  const listQueryKey = useMemo(() => workspaceTodoBatchKeys.list(workspaceDirs), [workspaceDirs]);
  const queryClient = useQueryClient();
  const listQuery = useQuery({
    queryKey: listQueryKey,
    queryFn: ({ signal }) => fetchWorkspaceTodoBatch(workspaceDirs, signal),
    enabled: enabled && workspaceDirs.length > 0,
    staleTime: Infinity,
  });
  const emptyByDir = useMemo(
    () => Object.fromEntries(workspaceDirs.map(dir => [dir, { dir, counts: emptyCounts, items: [] }])),
    [workspaceDirs],
  );
  const byDir = listQuery.data?.byDir ?? emptyByDir;
  const errorsByDir = listQuery.data?.errorsByDir ?? {};
  const error = listQuery.error?.message ?? listQuery.data?.error ?? '';
  const loadingList = listQuery.isFetching;
  const [selected, setSelected] = useState<SelectedTodo | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  // Status / priority / external-issue facets applied to the lists. Defaults to
  // excluding completed (closed) and persists the user's choice across reloads.
  const [filters, setFiltersState] = useState<TodoFilters>(loadTodoFilters);
  // Row density (comfortable/compact) for the lists, persisted across reloads.
  const [density, setDensityState] = useState<TodoDensity>(loadDensity);
  // Grouping dimension (workspace/severity/age) for the lists, persisted too.
  const [groupBy, setGroupByState] = useState<TodoGroupBy>(loadGroupBy);
  // Sort column and direction for rows within each group, persisted too.
  const [sortBy, setSortByState] = useState<TodoSort>(loadTodoSort);
  // Activity time-range filter (clicky-ui TimeRange tokens); null shows all.
  const [timeRange, setTimeRangeState] = useState<TodoTimeRange | null>(loadTimeRange);
  // Bulk-edit mode and its checked set. Deliberately not persisted: a selection
  // restored from a previous session would apply an edit to todos the user has
  // long forgotten checking.
  const selection = useTodoSelection();

  const setFilters = useCallback((next: TodoFilters) => {
    setFiltersState(next);
    saveTodoFilters(next);
  }, []);

  // toggleStatus is the per-workspace count badge's click: hide this status, or
  // un-hide it. The FilterBar's own chips drive the full tri-state through
  // setFilters.
  const toggleStatus = useCallback((status: TodoStatus) => {
    setFiltersState(prev => {
      const next = { ...prev, statuses: toggleStatusFilter(prev.statuses, status) };
      saveTodoFilters(next);
      return next;
    });
  }, []);

  const setDensity = useCallback((next: TodoDensity) => {
    setDensityState(next);
    saveDensity(next);
  }, []);

  const setGroupBy = useCallback((next: TodoGroupBy) => {
    setGroupByState(next);
    saveGroupBy(next);
  }, []);

  const setSortBy = useCallback((next: TodoSort) => {
    setSortByState(next);
    saveTodoSort(next);
  }, []);

  const setTimeRange = useCallback((next: TodoTimeRange | null) => {
    setTimeRangeState(next);
    saveTimeRange(next);
  }, []);

  // The applied route identity distinguishes a URL-driven selection from a
  // locally clicked one while the parent catches up with onNavigate.
  const appliedId = useRef('');
  const selectedRef = useRef<SelectedTodo | null>(null);
  const setSelection = useCallback((next: SelectedTodo | null) => {
    selectedRef.current = next;
    setSelected(next);
  }, []);
  const select = useCallback((next: SelectedTodo | null) => {
    appliedId.current = next?.ref ?? '';
    setSelection(next);
    onNavigate?.(next?.ref ?? '');
  }, [onNavigate, setSelection]);

  const resolvingRoute = !!selectedId
    && selectedId !== appliedId.current
    && selectedRef.current?.ref !== selectedId;
  const globalDetailQuery = useQuery({
    ...todoGlobalItemQueryOptions(selectedId),
    enabled: resolvingRoute,
  });
  const selectedDetailQuery = useQuery({
    ...todoItemQueryOptions(selected?.dir ?? '', selected?.ref ?? '', {
      pollWhileActive: true,
      intervalMs: activeTodoDetailPollInterval,
    }),
    enabled: !!selected && !resolvingRoute,
  });
  const detail = resolvingRoute ? null : selectedDetailQuery.data ?? null;
  const detailError = resolvingRoute
    ? globalDetailQuery.error?.message ?? ''
    : selectedDetailQuery.error?.message ?? '';
  const loadingDetail = resolvingRoute
    ? globalDetailQuery.isFetching
    : !!selected && selectedDetailQuery.isFetching && !selectedDetailQuery.data;

  // Refresh the batch once when a polled run becomes terminal. React Query owns
  // the visible-only interval and retains the last good detail on read errors.
  const previousDetail = useRef<{ key: string; status?: TodoStatus }>({ key: '' });
  useEffect(() => {
    const key = selected ? `${selected.dir}\u0000${selected.ref}` : '';
    const status = detail?.status;
    const previous = previousDetail.current;
    if (previous.key === key && previous.status === 'in_progress' && status && status !== 'in_progress') {
      void queryClient.invalidateQueries({ queryKey: listQueryKey, exact: true });
    }
    previousDetail.current = { key, status };
  }, [detail?.status, listQueryKey, queryClient, selected]);

  // Resolve a deep link independently of the batch. The global query accepts a
  // canonical ID, short ID, session ID, or imported alias; seed its canonical
  // workspace query so adopting the selection does not issue a duplicate read.
  useEffect(() => {
    if (selectedId === appliedId.current) return;
    if (!selectedId) {
      appliedId.current = '';
      setSelection(null);
      return;
    }
    if (selectedRef.current?.ref === selectedId) {
      appliedId.current = selectedId;
      return;
    }
    setSelection(null);
    const todo = globalDetailQuery.data;
    if (!todo) return;
    const canonicalRef = todo.ref.trim();
    const canonicalDir = todo.cwd?.trim() || todo.dir?.trim() || '';
    appliedId.current = selectedId;
    queryClient.setQueryData(todoQueryKeys.item(canonicalDir, canonicalRef), todo);
    setSelection({ dir: canonicalDir, ref: canonicalRef });
  }, [globalDetailQuery.data, queryClient, selectedId, setSelection]);

  const aggregate = useMemo(
    () => workspaces.reduce((acc, ws) => addCounts(acc, byDir[ws.dir]?.counts ?? ws.todoCounts ?? emptyCounts), emptyCounts),
    [workspaces, byDir],
  );

  const refresh = useCallback(() => {
    void queryClient.invalidateQueries({ queryKey: listQueryKey, exact: true });
  }, [listQueryKey, queryClient]);

  const created = useCallback((dir: string, todo: TodoItem) => {
    setShowCreate(false);
    setTodoQueryData(queryClient, dir, todo);
    select({ dir, ref: todo.ref });
    refresh();
  }, [queryClient, refresh, select]);

  const updateItem = useCallback((todo: TodoItem) => {
    if (selected) setTodoQueryData(queryClient, selected.dir, todo);
    refresh();
  }, [queryClient, refresh, selected]);

  const deleted = useCallback(() => {
    select(null);
    refresh();
  }, [refresh, select]);

  // A transferred todo now lives in the target workspace: follow it there so the
  // detail pane keeps showing it after the move (the source list loses it).
  const transferred = useCallback((toDir: string, todo: TodoItem) => {
    setTodoQueryData(queryClient, toDir, todo);
    select({ dir: toDir, ref: todo.ref });
    refresh();
  }, [queryClient, refresh, select]);

  return {
    workspaces,
    byDir,
    errorsByDir,
    loadingList,
    error,
    detailError,
    aggregate,
    selected,
    setSelected: setSelection,
    select,
    detail,
    loadingDetail,
    refresh,
    showCreate,
    setShowCreate,
    created,
    updateItem,
    deleted,
    transferred,
    filters,
    setFilters,
    toggleStatus,
    density,
    setDensity,
    groupBy,
    setGroupBy,
    sortBy,
    setSortBy,
    timeRange,
    setTimeRange,
    selection,
  };
}

// WorkspaceTodos is the shared todos data layer the dashboard's AppShell body
// slots render off of (header, actions, filter toolbar, sidebar list, detail).
export type WorkspaceTodos = ReturnType<typeof useWorkspaceTodos>;
