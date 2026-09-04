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

// useTodoSelection owns the checked set.
//
// There is no bulk-edit mode. A mode was a toggle the user had to find before
// any bulk action existed for them, and it hid the checkboxes that would have
// advertised the capability. Checking a row is now the entry point, and the
// action toolbar appears with the first check.
export function useTodoSelection() {
  const [selection, setSelection] = useState<SelectionState>(() => new Set<string>());

  const toggleSelected = useCallback((todo: SelectedTodo) => {
    setSelection(state => toggleSelectionKeys(state, selectionKey(todo)));
  }, []);

  const setGroupSelected = useCallback((todos: SelectedTodo[], selected: boolean) => {
    setSelection(state => setSelectionKeys(state, todos.map(selectionKey), selected));
  }, []);

  const clearSelection = useCallback(() => setSelection(new Set<string>()), []);

  // replaceSelection takes the whole checked set at once, for a host that owns
  // its own checkboxes and reports the result rather than the change — the
  // full-width table's DataTable rowSelection, whose select-all would otherwise
  // mean one state update per row.
  const replaceSelection = useCallback((keys: readonly string[]) => {
    setSelection(new Set(keys));
  }, []);

  const isSelected = useCallback((todo: SelectedTodo) => selection.has(selectionKey(todo)), [selection]);

  const targets = useMemo(() => selectionTargets(selection), [selection]);

  return { selection, isSelected, toggleSelected, setGroupSelected, clearSelection, replaceSelection, targets };
}

export type TodoSelection = ReturnType<typeof useTodoSelection>;
