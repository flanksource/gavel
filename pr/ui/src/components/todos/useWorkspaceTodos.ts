import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import type { Project, TodoDensity, TodoGroupBy, TodoItem, TodoLayout, TodoListResponse, TodoStatus } from '../../types';
import { addCounts, countsFromItems, emptyCounts } from './format';
import { useWorkspaceChanges } from './useWorkspaceChanges';
import { loadDensity, saveDensity } from './todoDensity';
import { loadGroupBy, saveGroupBy } from './todoGroup';
import { loadLayout, saveLayout } from './todoLayout';
import { loadTodoFilters, saveTodoFilters, toggleStatusFilter, type TodoFilters } from './todoFilter';
import type { TodoSort } from './todoSort';
import { loadTodoSort, saveTodoSort } from './todoSort';
import { loadTimeRange, saveTimeRange, type TodoTimeRange } from './todoTimeRange';
import {
  HIDDEN_PROJECTS_FILTER, setTodoQueryData, todoGlobalItemQueryOptions, todoItemQueryOptions, todoQueryKeys,
  workspaceTodoListQueryOptions,
} from './todoQueries';
import { useTodoSelection, type SelectedTodo } from './todoSelection';
import { useTodoTagIndexes } from './tagQueries';
import type { TodoDetailView } from '../../routes';

export type { SelectedTodo };

// useWorkspaceTodos drives the shared todos data layer: it lists every visible
// workspace's todos with one GET /api/todos?exclude=hidden-projects, loads the selected todo's detail,
// aggregates counts, and exposes the create/update/delete callbacks. Both the
// dashboard TodoView and the compact MenubarTodos render off this one hook so
// they hit the same /api/todos endpoints and stay in sync.
//
// selectedId/onNavigate wire the selection to the URL (/todos/{guid}, where the
// guid is the todo ref), and view/onViewChange do the same for what the detail
// pane shows (?tab=&sessionTab=&sessions=): the dashboard passes them so a todo
// view is deep-linkable and back/forward works; the menubar omits them and
// keeps purely-local state.
//
// `enabled` gates the list fetch: the dashboard mounts this hook permanently (so
// the Todos chrome can live in the AppShell's body slots) but passes false while
// another tab is active, so the workspaces aren't listed until the Todos tab is
// opened. Cached results survive a tab switch, so reopening is instant.
export interface WorkspaceTodosOptions {
  selectedId?: string;
  onNavigate?: (id: string) => void;
  view?: TodoDetailView;
  onViewChange?: (view: TodoDetailView, mode?: 'push' | 'replace') => void;
  enabled?: boolean;
}

export function useWorkspaceTodos(
  projects: Project[],
  { selectedId = '', onNavigate, view, onViewChange, enabled = true }: WorkspaceTodosOptions = {},
) {
  // De-duplicate configured directories. Rendering retains projects.json order;
  // the visible-workspace set itself is decided server-side by the
  // hidden-projects filter.
  const workspaces = useMemo(() => {
    const seen = new Set<string>();
    return projects.filter(project => {
      if (!project.dir || seen.has(project.dir)) return false;
      seen.add(project.dir);
      return true;
    });
  }, [projects]);
  const workspaceDirsKey = JSON.stringify(workspaces.map(workspace => workspace.dir).sort());
  const workspaceDirs = useMemo(() => JSON.parse(workspaceDirsKey) as string[], [workspaceDirsKey]);
  // One tag index per workspace, resolved here rather than in the rows: a row
  // component renders once per todo, and a subscription per row would be
  // hundreds of them.
  const queryClient = useQueryClient();
  const changes = useWorkspaceChanges(enabled);
  const tagsByDir = useTodoTagIndexes(workspaceDirs, enabled && changes.ready);
  const listQuery = useQuery({ ...workspaceTodoListQueryOptions(HIDDEN_PROJECTS_FILTER), enabled: enabled && changes.ready });
  // Group the flat list by workspace. Every configured workspace keeps an entry
  // (empty when it has no todos) so the grouped rendering and the aggregate do
  // not depend on which workspaces happen to hold items.
  const byDir = useMemo(() => {
    const items = listQuery.data?.items ?? [];
    const grouped: Record<string, TodoItem[]> = Object.fromEntries(workspaceDirs.map(dir => [dir, []]));
    for (const item of items) (grouped[item.dir!] ??= []).push(item);
    return Object.fromEntries(Object.entries(grouped).map(([dir, list]): [string, TodoListResponse] => (
      [dir, { dir, counts: list.length ? countsFromItems(list) : emptyCounts, items: list }]
    )));
  }, [listQuery.data, workspaceDirs]);
  const error = changes.error || listQuery.error?.message || '';
  const loadingList = listQuery.isFetching;
  // False until the first list has landed: `byDir` is then a placeholder of
  // empty lists, which must not be read as workspaces that hold no todos.
  const listReady = listQuery.data !== undefined;
  const [selected, setSelected] = useState<SelectedTodo | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  // Status / priority / external-issue facets applied to the lists. Defaults to
  // excluding completed (closed) and persists the user's choice across reloads.
  const [filters, setFiltersState] = useState<TodoFilters>(loadTodoFilters);
  // Row density (comfortable/compact) for the lists, persisted across reloads.
  const [density, setDensityState] = useState<TodoDensity>(loadDensity);
  // Grouping dimension (workspace/severity/age/none) for the lists, persisted too.
  const [groupBy, setGroupByState] = useState<TodoGroupBy>(loadGroupBy);
  // Master-detail (split) vs full-width table, persisted too. Dashboard-only:
  // the menubar layout is single-column regardless.
  const [layout, setLayoutState] = useState<TodoLayout>(loadLayout);
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
    const next = { ...filters, statuses: toggleStatusFilter(filters.statuses, status) };
    setFiltersState(next);
    saveTodoFilters(next);
  }, [filters]);

  const setDensity = useCallback((next: TodoDensity) => {
    setDensityState(next);
    saveDensity(next);
  }, []);

  const setGroupBy = useCallback((next: TodoGroupBy) => {
    setGroupByState(next);
    saveGroupBy(next);
  }, []);

  const setLayout = useCallback((next: TodoLayout) => {
    setLayoutState(next);
    saveLayout(next);
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
  // Without a routed view (the menubar) the detail view is local and a newly
  // selected todo opens on its defaults, as a routed selection does.
  const [localView, setLocalView] = useState<TodoDetailView>({});
  const detailView = view ?? localView;
  const setDetailView = useCallback((next: TodoDetailView, mode: 'push' | 'replace' = 'push') => {
    if (onViewChange) onViewChange(next, mode);
    else setLocalView(next);
  }, [onViewChange]);

  const select = useCallback((next: SelectedTodo | null) => {
    appliedId.current = next?.ref ?? '';
    setSelection(next);
    setLocalView({});
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
    ...todoItemQueryOptions(selected?.dir ?? '', selected?.ref ?? ''),
    enabled: !!selected && !resolvingRoute,
  });
  const detail = resolvingRoute ? null : selectedDetailQuery.data ?? null;
  const detailError = resolvingRoute
    ? globalDetailQuery.error?.message ?? ''
    : selectedDetailQuery.error?.message ?? '';
  const loadingDetail = resolvingRoute
    ? globalDetailQuery.isFetching
    : !!selected && selectedDetailQuery.isFetching && !selectedDetailQuery.data;

  // Resolve a deep link independently of the list. The global query accepts a
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
    void queryClient.invalidateQueries({ queryKey: todoQueryKeys.workspaceList(HIDDEN_PROJECTS_FILTER), exact: true });
  }, [queryClient]);

  const created = useCallback((dir: string, todo: TodoItem) => {
    setShowCreate(false);
    setTodoQueryData(queryClient, dir, todo);
    select({ dir, ref: todo.ref });
  }, [queryClient, select]);

  const updateItem = useCallback((todo: TodoItem) => {
    if (selected) setTodoQueryData(queryClient, selected.dir, todo);
  }, [queryClient, selected]);

  const deleted = useCallback(() => {
    select(null);
  }, [select]);

  // A transferred todo now lives in the target workspace: follow it there so the
  // detail pane keeps showing it after the move (the source list loses it).
  const transferred = useCallback((toDir: string, todo: TodoItem) => {
    setTodoQueryData(queryClient, toDir, todo);
    select({ dir: toDir, ref: todo.ref });
  }, [queryClient, select]);

  // This object is the todos data layer: App threads it as a prop into the
  // toolbar, the sidebar list, the detail pane and every navbar control. A
  // fresh literal per render meant any App render — one every SSE tick — cost a
  // full reconcile of the whole todos subtree, hundreds of rows included.
  // Every field below is already stable on its own (useState setters, useCallback
  // handlers, useMemo values), so memoising here is what makes that stability
  // reach the consumers.
  return useMemo(() => ({
    workspaces,
    byDir,
    loadingList,
    listReady,
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
    layout,
    setLayout,
    sortBy,
    setSortBy,
    timeRange,
    setTimeRange,
    selection,
    tagsByDir,
    detailView,
    setDetailView,
  }), [
    workspaces, byDir, loadingList, listReady, error, detailError, aggregate,
    selected, setSelection, select, detail, loadingDetail, refresh, showCreate,
    created, updateItem, deleted, transferred, filters, setFilters, toggleStatus,
    density, setDensity, groupBy, setGroupBy, layout, setLayout, sortBy, setSortBy,
    timeRange, setTimeRange, selection, tagsByDir, detailView, setDetailView,
  ]);
}

// WorkspaceTodos is the shared todos data layer the dashboard's AppShell body
// slots render off of (header, actions, filter toolbar, sidebar list, detail).
export type WorkspaceTodos = ReturnType<typeof useWorkspaceTodos>;
