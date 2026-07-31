import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { Project, TodoDensity, TodoGroupBy, TodoItem, TodoListResponse, TodoStatus } from '../../types';
import { addCounts, emptyCounts, todoQuery } from './format';
import { loadDensity, saveDensity } from './todoDensity';
import { loadGroupBy, saveGroupBy } from './todoGroup';
import { loadHiddenStatuses, saveHiddenStatuses, toggleHiddenStatus } from './todoFilter';
import type { TodoSort } from './todoSort';
import { loadTodoSort, saveTodoSort } from './todoSort';
import { loadTimeRange, saveTimeRange, type TodoTimeRange } from './todoTimeRange';

const activeTodoDetailPollInterval = 1000;

export interface SelectedTodo {
  dir: string;
  ref: string;
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
  const dirsKey = useMemo(() => JSON.stringify(workspaces.map(w => w.dir)), [workspaces]);

  const [byDir, setByDir] = useState<Record<string, TodoListResponse>>({});
  const [loadingList, setLoadingList] = useState(false);
  const [error, setError] = useState('');
  const [detailError, setDetailError] = useState('');
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
  // Sort column and direction for rows within each group, persisted too.
  const [sortBy, setSortByState] = useState<TodoSort>(loadTodoSort);
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

  const setSortBy = useCallback((next: TodoSort) => {
    setSortByState(next);
    saveTodoSort(next);
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
    const dirs = JSON.parse(dirsKey) as string[];
    if (dirs.length === 0) {
      setByDir({});
      setError('');
      setLoadingList(false);
      return;
    }
    let cancelled = false;
    setLoadingList(true);
    setError('');
    (async () => {
      const results = await Promise.all(dirs.map(async (dir) => {
        try {
          const res = await fetch(`/api/todos?${todoQuery(dir)}`);
          const data = await res.json().catch(() => ({}));
          if (!res.ok) throw new Error(data.error || 'Load failed');
          return { dir, data: data as TodoListResponse, error: '' };
        } catch (err: any) {
          return {
            dir,
            data: { dir, counts: emptyCounts, items: [] } as TodoListResponse,
            error: err?.message || 'Load failed',
          };
        }
      }));
      if (!cancelled) {
        setByDir(Object.fromEntries(results.map(result => [result.dir, result.data])));
        setError([...new Set(results.map(result => result.error).filter(Boolean))].join('; '));
        setLoadingList(false);
      }
    })();
    return () => { cancelled = true; };
  }, [dirsKey, refreshTick, enabled]);

  // A route lookup below already returns the full detail. Skip the ordinary
  // dir-scoped refetch for that exact canonical selection.
  const skipDetailKey = useRef('');
  // Tracks the route identity that produced the current selection. It is
  // declared before the detail effect so a route change can suppress a stale
  // refetch of the previously selected canonical Todo while the new global
  // UUID/alias lookup is in flight.
  const appliedId = useRef('');

  // Load a clicked todo's detail (body + history).
  useEffect(() => {
    if (!selected) {
      if (!selectedId) {
        setDetail(null);
        setDetailError('');
        setLoadingDetail(false);
      }
      return;
    }
    if (onNavigate && selectedId !== appliedId.current) return;
    const selectionKey = `${selected.dir}\u0000${selected.ref}`;
    if (skipDetailKey.current === selectionKey) {
      skipDetailKey.current = '';
      setLoadingDetail(false);
      return;
    }
    let cancelled = false;
    setLoadingDetail(true);
    setDetailError('');
    (async () => {
      try {
        const params = new URLSearchParams(todoQuery(selected.dir));
        params.set('ref', selected.ref);
        const res = await fetch(`/api/todos/item?${params.toString()}`);
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data.error || 'Load failed');
        if (!cancelled) setDetail(data as TodoItem);
      } catch (err: any) {
        if (!cancelled) {
          setDetail(null);
          setDetailError(err?.message || 'Load failed');
        }
      } finally {
        if (!cancelled) setLoadingDetail(false);
      }
    })();
    return () => { cancelled = true; };
  }, [selected, selectedId, onNavigate]);

  // Inline agents discover their provider session ID after the run request has
  // already returned. Keep the selected detail current while execution is live
  // so the Session tab can attach as soon as that identity is persisted, and so
  // the terminal plan/run projection replaces the optimistic in-progress row
  // without requiring a manual page refresh. This deliberately polls only the
  // selected detail; workspace lists refresh once when the run leaves the
  // active state instead of fanning out on every tick.
  useEffect(() => {
    if (!selected || detail?.status !== 'in_progress') return;

    let cancelled = false;
    let timer: number | undefined;
    const poll = async () => {
      let stillActive = true;
      try {
        const params = new URLSearchParams(todoQuery(selected.dir));
        params.set('ref', selected.ref);
        const res = await fetch(`/api/todos/item?${params.toString()}`);
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data.error || 'Load failed');
        if (cancelled) return;
        const next = data as TodoItem;
        stillActive = next.status === 'in_progress';
        setDetail(next);
        if (!stillActive) setRefreshTick(tick => tick + 1);
      } catch {
        // A transient detail read must not disconnect a live session. The next
        // tick retries while the locally projected todo is still active.
      }
      if (!cancelled && stillActive) {
        timer = window.setTimeout(poll, activeTodoDetailPollInterval);
      }
    };

    timer = window.setTimeout(poll, activeTodoDetailPollInterval);
    return () => {
      cancelled = true;
      if (timer !== undefined) window.clearTimeout(timer);
    };
  }, [detail?.status, selected]);

  // select changes the active todo and pushes its ref into the URL (when the
  // caller wired onNavigate). The resolution effect below mirrors the reverse —
  // a URL change (deep link, back/forward) into `selected`.
  const selectedRef = useRef<SelectedTodo | null>(null);
  const setSelection = useCallback((next: SelectedTodo | null) => {
    selectedRef.current = next;
    setSelected(next);
  }, []);
  const select = useCallback((next: SelectedTodo | null) => {
    appliedId.current = next?.ref ?? '';
    setDetailError('');
    setSelection(next);
    onNavigate?.(next?.ref ?? '');
  }, [onNavigate, setSelection]);

  // Resolve a deep link directly by reference, independently of the workspace
  // list requests. The database endpoint resolves canonical IDs, short IDs, and
  // imported aliases globally and returns the canonical ref plus workspace CWD.
  // Adopt those for subsequent mutations but do not navigate: an alias URL must
  // remain stable when it resolves successfully.
  useEffect(() => {
    if (selectedId === appliedId.current) return;
    if (!selectedId) {
      appliedId.current = '';
      setSelection(null);
      setDetail(null);
      setDetailError('');
      setLoadingDetail(false);
      return;
    }
    if (selectedRef.current?.ref === selectedId) {
      appliedId.current = selectedId;
      return;
    }

    let cancelled = false;
    setSelection(null);
    setDetail(null);
    setDetailError('');
    setLoadingDetail(true);
    (async () => {
      try {
        const params = new URLSearchParams({ ref: selectedId });
        const res = await fetch(`/api/todos/item?${params.toString()}`);
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data.error || `Load failed (HTTP ${res.status})`);
        const todo = data as TodoItem & { dir?: string };
        const canonicalRef = todo.ref?.trim();
        const canonicalDir = todo.cwd?.trim() || todo.dir?.trim();
        if (!canonicalRef) throw new Error('Database returned a todo without a canonical reference');
        if (!canonicalDir) throw new Error('Database returned a todo without a workspace directory');
        if (cancelled) return;
        appliedId.current = selectedId;
        skipDetailKey.current = `${canonicalDir}\u0000${canonicalRef}`;
        setSelection({ dir: canonicalDir, ref: canonicalRef });
        setDetail(todo);
      } catch (err: any) {
        if (!cancelled) {
          appliedId.current = selectedId;
          setDetailError(err?.message || 'Load failed');
        }
      } finally {
        if (!cancelled) setLoadingDetail(false);
      }
    })();
    return () => { cancelled = true; };
  }, [selectedId, setSelection]);

  const aggregate = useMemo(
    () => workspaces.reduce((acc, ws) => addCounts(acc, byDir[ws.dir]?.counts ?? ws.todoCounts ?? emptyCounts), emptyCounts),
    [workspaces, byDir],
  );

  const refresh = useCallback(() => setRefreshTick(t => t + 1), []);

  const created = useCallback((dir: string, todo: TodoItem) => {
    setShowCreate(false);
    select({ dir, ref: todo.ref });
    setDetail(todo);
    refresh();
  }, [refresh, select]);

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
    select({ dir: toDir, ref: todo.ref });
    setDetail(todo);
    refresh();
  }, [refresh, select]);

  return {
    workspaces,
    byDir,
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
    hiddenStatuses,
    toggleStatus,
    density,
    setDensity,
    groupBy,
    setGroupBy,
    sortBy,
    setSortBy,
    timeRange,
    setTimeRange,
  };
}

// WorkspaceTodos is the shared todos data layer the dashboard's AppShell body
// slots render off of (header, actions, filter toolbar, sidebar list, detail).
export type WorkspaceTodos = ReturnType<typeof useWorkspaceTodos>;
