import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { Project, TodoDensity, TodoGroupBy, TodoItem, TodoListResponse, TodoStatus } from '../../types';
import { addCounts, emptyCounts, todoQuery } from './format';
import { loadDensity, saveDensity } from './todoDensity';
import { loadGroupBy, saveGroupBy } from './todoGroup';
import { loadHiddenStatuses, saveHiddenStatuses, toggleHiddenStatus } from './todoFilter';
import { loadTimeRange, saveTimeRange, type TodoTimeRange } from './todoTimeRange';

export interface SelectedTodo {
  dir: string;
  ref: string;
  provider: string;
}

// useWorkspaceTodos drives the shared todos data layer: it lists every
// configured workspace's todos in parallel, loads the selected todo's detail,
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
  // Every configured workspace with a directory is listed straight from
  // projects.json; ones with no todos render an empty "No todos" group.
  const workspaces = useMemo(() => projects.filter(p => !!p.dir), [projects]);
  // Each entry encodes the workspace dir and its provider (tab-separated) so the
  // list refetches when either the set of workspaces or a pinned provider changes.
  const dirsKey = useMemo(() => workspaces.map(w => `${w.dir}\t${w.todoProvider || 'auto'}`).join('|'), [workspaces]);
  const providerFor = useCallback(
    (dir: string) => workspaces.find(w => w.dir === dir)?.todoProvider || 'auto',
    [workspaces],
  );

  const [byDir, setByDir] = useState<Record<string, TodoListResponse>>({});
  const [loadingList, setLoadingList] = useState(false);
  const [error, setError] = useState('');
  const [refreshTick, setRefreshTick] = useState(0);
  const [selected, setSelected] = useState<SelectedTodo | null>(null);
  const [detail, setDetail] = useState<TodoItem | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  // Closed/Status filter: the set of statuses hidden from the lists. Defaults to
  // hiding completed (closed) and persists the user's choice across reloads.
  const [hiddenStatuses, setHiddenStatuses] = useState<Set<TodoStatus>>(loadHiddenStatuses);
  // Row density (comfortable/compact) for the lists, persisted across reloads.
  const [density, setDensityState] = useState<TodoDensity>(loadDensity);
  // Grouping dimension (workspace/severity/age) for the lists, persisted too.
  const [groupBy, setGroupByState] = useState<TodoGroupBy>(loadGroupBy);
  // Activity time-range filter (clicky-ui TimeRange tokens); null shows all.
  const [timeRange, setTimeRangeState] = useState<TodoTimeRange | null>(loadTimeRange);

  const toggleStatus = useCallback((status: TodoStatus) => {
    setHiddenStatuses(prev => {
      const next = toggleHiddenStatus(prev, status);
      saveHiddenStatuses(next);
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

  const setTimeRange = useCallback((next: TodoTimeRange | null) => {
    setTimeRangeState(next);
    saveTimeRange(next);
  }, []);

  // Fetch every workspace's todos in parallel; refetch only when the set of
  // workspace directories changes or on an explicit refresh, not on every
  // projects poll. A per-workspace failure degrades to an empty group.
  useEffect(() => {
    if (!enabled) return;
    const specs = dirsKey ? dirsKey.split('|').map(s => s.split('\t')) : [];
    if (specs.length === 0) {
      setByDir({});
      return;
    }
    let cancelled = false;
    setLoadingList(true);
    setError('');
    (async () => {
      const entries = await Promise.all(specs.map(async ([dir, provider]) => {
        try {
          const res = await fetch(`/api/todos?${todoQuery(dir, provider)}`);
          const data = await res.json();
          if (!res.ok) throw new Error(data.error || 'Load failed');
          return [dir, data as TodoListResponse] as const;
        } catch {
          return [dir, { provider, dir, counts: emptyCounts, items: [] } as TodoListResponse] as const;
        }
      }));
      if (!cancelled) {
        setByDir(Object.fromEntries(entries));
        setLoadingList(false);
      }
    })();
    return () => { cancelled = true; };
  }, [dirsKey, refreshTick, enabled]);

  // Load the selected todo's detail (body + history).
  useEffect(() => {
    if (!selected) {
      setDetail(null);
      return;
    }
    let cancelled = false;
    setLoadingDetail(true);
    (async () => {
      try {
        const params = new URLSearchParams(todoQuery(selected.dir, selected.provider));
        params.set('ref', selected.ref);
        const res = await fetch(`/api/todos/item?${params.toString()}`);
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Load failed');
        if (!cancelled) setDetail(data as TodoItem);
      } catch {
        if (!cancelled) setDetail(null);
      } finally {
        if (!cancelled) setLoadingDetail(false);
      }
    })();
    return () => { cancelled = true; };
  }, [selected]);

  // select changes the active todo and pushes its ref into the URL (when the
  // caller wired onNavigate). The resolution effect below mirrors the reverse —
  // a URL change (deep link, back/forward) into `selected`.
  const select = useCallback((next: SelectedTodo | null) => {
    setSelected(next);
    onNavigate?.(next?.ref ?? '');
  }, [onNavigate]);

  // Resolve the URL's selectedId into a concrete {dir, ref, provider} by finding
  // which workspace's list holds that ref. appliedId tracks the last id we
  // resolved so a user click (which sets `selected` before the URL catches up)
  // is never clobbered, and a not-yet-loaded deep link retries when byDir fills.
  const appliedId = useRef('');
  useEffect(() => {
    if (selectedId === appliedId.current) return;
    if (!selectedId) {
      appliedId.current = '';
      setSelected(null);
      return;
    }
    for (const ws of workspaces) {
      if (byDir[ws.dir]?.items.some(item => item.ref === selectedId)) {
        appliedId.current = selectedId;
        // Keep the existing selection object when it already points here (a user
        // click set it before the URL caught up) so the detail effect doesn't
        // refetch the same todo.
        setSelected(prev =>
          prev && prev.dir === ws.dir && prev.ref === selectedId
            ? prev
            : { dir: ws.dir, ref: selectedId, provider: ws.todoProvider || 'auto' });
        return;
      }
    }
  }, [selectedId, byDir, workspaces]);

  const aggregate = useMemo(
    () => workspaces.reduce((acc, ws) => addCounts(acc, byDir[ws.dir]?.counts ?? ws.todoCounts ?? emptyCounts), emptyCounts),
    [workspaces, byDir],
  );

  const refresh = useCallback(() => setRefreshTick(t => t + 1), []);

  const created = useCallback((dir: string, todo: TodoItem) => {
    setShowCreate(false);
    select({ dir, ref: todo.ref, provider: providerFor(dir) });
    setDetail(todo);
    refresh();
  }, [providerFor, refresh, select]);

  const updateItem = useCallback((todo: TodoItem) => {
    setDetail(todo);
    refresh();
  }, [refresh]);

  const deleted = useCallback(() => {
    setDetail(null);
    select(null);
    refresh();
  }, [refresh, select]);

  // A transferred todo now lives in the target workspace: follow it there so the
  // detail pane keeps showing it after the move (the source list loses it).
  const transferred = useCallback((toDir: string, todo: TodoItem) => {
    select({ dir: toDir, ref: todo.ref, provider: providerFor(toDir) });
    setDetail(todo);
    refresh();
  }, [providerFor, refresh, select]);

  return {
    workspaces,
    byDir,
    loadingList,
    error,
    aggregate,
    selected,
    setSelected,
    select,
    detail,
    loadingDetail,
    providerFor,
    refresh,
    showCreate,
    setShowCreate,
    created,
    updateItem,
    deleted,
    transferred,
    hiddenStatuses,
    toggleStatus,
    density,
    setDensity,
    groupBy,
    setGroupBy,
    timeRange,
    setTimeRange,
  };
}

// WorkspaceTodos is the shared todos data layer the dashboard's AppShell body
// slots render off of (header, actions, filter toolbar, sidebar list, detail).
export type WorkspaceTodos = ReturnType<typeof useWorkspaceTodos>;
