import { useCallback, useMemo, useRef } from 'react';
import type { TodoItem, TodoListResponse } from '../../types';
import { childCountsById, childrenOf, parentCandidates, parentOf, parentTitlesById, type ChildCount, type TodoFamilyStatus } from './todoFamily';
import type { WorkspaceTodos } from './useWorkspaceTodos';

const noTodos: TodoItem[] = [];
const ready: TodoFamilyStatus = { state: 'ready' };
const loading: TodoFamilyStatus = { state: 'loading' };

// The selected todo's place in the hierarchy, read from its workspace's list:
// the listing is fetched whole, so children and candidate parents need no
// request of their own. Spread the result onto TodoDetail.
//
// `familyStatus` says whether that list can be trusted: before the first batch
// arrives, or after it fails, the family is empty because nothing is known, not
// because the todo has no children.
export function useTodoFamily({ byDir, selected, detail, select, listReady, error, errorsByDir }:
  Pick<WorkspaceTodos, 'byDir' | 'selected' | 'detail' | 'select' | 'listReady' | 'error' | 'errorsByDir'>) {
  const items = (selected ? byDir[selected.dir]?.items : undefined) ?? noTodos;
  const family = useMemo(() => (detail ? {
    childTodos: childrenOf(items, detail.id),
    parentTodo: parentOf(items, detail),
    parentCandidates: parentCandidates(items, detail),
  } : { childTodos: noTodos, parentTodo: null, parentCandidates: noTodos }), [items, detail]);
  const dir = selected?.dir;
  const workspaceError = dir === undefined ? undefined : errorsByDir[dir]?.message;
  const message = workspaceError ?? (listReady ? '' : error);
  const familyStatus = useMemo<TodoFamilyStatus>(() => {
    if (message) return { state: 'error', message };
    return listReady ? ready : loading;
  }, [message, listReady]);
  const onOpenTodo = useCallback((ref: string) => {
    if (dir !== undefined) select({ dir, ref });
  }, [dir, select]);
  return { ...family, familyStatus, onOpenTodo };
}

function sameEntries<V>(a: Map<string, V>, b: Map<string, V>, same: (x: V, y: V) => boolean): boolean {
  if (a.size !== b.size) return false;
  for (const [key, value] of a) {
    const other = b.get(key);
    if (other === undefined || !same(value, other)) return false;
  }
  return true;
}

// Keeps a derived map's identity across list refreshes that change none of its
// entries, so the table's columns — which close over it — are not rebuilt on
// every poll.
function useStableMap<V>(next: Map<string, V>, same: (x: V, y: V) => boolean): Map<string, V> {
  const kept = useRef(next);
  if (!sameEntries(kept.current, next, same)) kept.current = next;
  return kept.current;
}

// Titles of every todo that has children.
export function useParentTitles(byDir: Record<string, TodoListResponse>): Map<string, string> {
  return useStableMap(useMemo(() => parentTitlesById(byDir), [byDir]), (a, b) => a === b);
}

// Every parent's child count, across all workspaces.
export function useChildCounts(byDir: Record<string, TodoListResponse>): Map<string, ChildCount> {
  const next = useMemo(() => childCountsById(Object.values(byDir).flatMap(list => list.items)), [byDir]);
  return useStableMap(next, (a, b) => a.done === b.done && a.total === b.total);
}
