import { useCallback, useMemo, useState } from 'react';

// SelectedTodo addresses one TODO: its owning workspace directory plus its ref.
// It lives here rather than in useWorkspaceTodos because the selection helpers
// below are the pure core the hook is built on; useWorkspaceTodos re-exports it.
export interface SelectedTodo {
  dir: string;
  ref: string;
}

// A selection spans workspaces (the severity and age groupings flatten across
// them), so a todo is identified by workspace dir *and* ref. The NUL separator
// cannot occur in either half, so the key round-trips losslessly.
const SELECTION_SEPARATOR = '\u0000';

export type SelectionState = ReadonlySet<string>;

export function selectionKey({ dir, ref }: SelectedTodo): string {
  return `${dir.trim()}${SELECTION_SEPARATOR}${ref.trim()}`;
}

export function toggleSelectionKeys(state: SelectionState, key: string): SelectionState {
  const next = new Set(state);
  if (!next.delete(key)) next.add(key);
  return next;
}

export function setSelectionKeys(state: SelectionState, keys: string[], selected: boolean): SelectionState {
  const next = new Set(state);
  for (const key of keys) {
    if (selected) next.add(key);
    else next.delete(key);
  }
  return next;
}

export function selectionTargets(state: SelectionState): SelectedTodo[] {
  return [...state].map(key => {
    const separator = key.indexOf(SELECTION_SEPARATOR);
    return { dir: key.slice(0, separator), ref: key.slice(separator + 1) };
  });
}

// useTodoSelection owns bulk-edit mode and the checked set. Leaving bulk mode
// clears the selection: a hidden selection that survives into the next session
// is how a bulk edit lands on todos the user forgot they had checked.
export function useTodoSelection() {
  const [bulkMode, setBulkModeState] = useState(false);
  const [selection, setSelection] = useState<SelectionState>(() => new Set<string>());

  const setBulkMode = useCallback((next: boolean) => {
    setBulkModeState(next);
    if (!next) setSelection(new Set<string>());
  }, []);

  const toggleSelected = useCallback((todo: SelectedTodo) => {
    setSelection(state => toggleSelectionKeys(state, selectionKey(todo)));
  }, []);

  const setGroupSelected = useCallback((todos: SelectedTodo[], selected: boolean) => {
    setSelection(state => setSelectionKeys(state, todos.map(selectionKey), selected));
  }, []);

  const clearSelection = useCallback(() => setSelection(new Set<string>()), []);

  const isSelected = useCallback((todo: SelectedTodo) => selection.has(selectionKey(todo)), [selection]);

  const targets = useMemo(() => selectionTargets(selection), [selection]);

  return { bulkMode, setBulkMode, selection, isSelected, toggleSelected, setGroupSelected, clearSelection, targets };
}

export type TodoSelection = ReturnType<typeof useTodoSelection>;
