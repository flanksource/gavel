import { useCallback, useEffect, useMemo, useState } from 'react';
import type { Project, TodoItem, TodoListResponse, TodoStatus } from '../../types';
import { addCounts, emptyCounts, todoQuery } from './format';
import { loadHiddenStatuses, saveHiddenStatuses, toggleHiddenStatus } from './todoFilter';

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
export function useWorkspaceTodos(projects: Project[]) {
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

  const toggleStatus = useCallback((status: TodoStatus) => {
    setHiddenStatuses(prev => {
      const next = toggleHiddenStatus(prev, status);
      saveHiddenStatuses(next);
      return next;
    });
  }, []);

  // Fetch every workspace's todos in parallel; refetch only when the set of
  // workspace directories changes or on an explicit refresh, not on every
  // projects poll. A per-workspace failure degrades to an empty group.
  useEffect(() => {
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
  }, [dirsKey, refreshTick]);

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

  const aggregate = useMemo(
    () => workspaces.reduce((acc, ws) => addCounts(acc, byDir[ws.dir]?.counts ?? ws.todoCounts ?? emptyCounts), emptyCounts),
    [workspaces, byDir],
  );

  const refresh = useCallback(() => setRefreshTick(t => t + 1), []);

  const created = useCallback((dir: string, todo: TodoItem) => {
    setShowCreate(false);
    setSelected({ dir, ref: todo.ref, provider: providerFor(dir) });
    setDetail(todo);
    refresh();
  }, [providerFor, refresh]);

  const updateItem = useCallback((todo: TodoItem) => {
    setDetail(todo);
    refresh();
  }, [refresh]);

  const deleted = useCallback(() => {
    setDetail(null);
    setSelected(null);
    refresh();
  }, [refresh]);

  return {
    workspaces,
    byDir,
    loadingList,
    error,
    aggregate,
    selected,
    setSelected,
    detail,
    loadingDetail,
    providerFor,
    refresh,
    showCreate,
    setShowCreate,
    created,
    updateItem,
    deleted,
    hiddenStatuses,
    toggleStatus,
  };
}
