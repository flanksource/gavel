import type { DataTableColumn } from '@flanksource/clicky-ui/data';
import type { Project, TodoGroupBy, TodoItem, TodoListResponse } from '../../types';
import type { TodoFilters } from './todoFilter';
import { isEntryVisible } from './todoFilter';
import { bucketTodos, flattenTodos, type TodoEntry } from './todoGroup';
import type { SelectedTodo } from './todoSelection';
import type { TodoSort } from './todoSort';
import { todoComparator } from './todoSort';
import { resolveRange, type TodoTimeRange } from './todoTimeRange';

export interface TodoNavigationState {
  position: number;
  total: number;
  previous: SelectedTodo | null;
  next: SelectedTodo | null;
}

export function orderedTodoNavigationEntries({
  workspaces,
  byDir,
  filters,
  groupBy,
  sortBy,
  timeRange,
  now,
}: {
  workspaces: Project[];
  byDir: Record<string, TodoListResponse>;
  filters: TodoFilters;
  groupBy: TodoGroupBy;
  sortBy: TodoSort;
  timeRange: TodoTimeRange | null;
  now: number;
}): TodoEntry[] {
  const range = resolveRange(timeRange, now);
  const entries = flattenTodos(workspaces, byDir).filter(entry => isEntryVisible(entry, filters, range));
  if (groupBy !== 'workspace') {
    return bucketTodos(entries, groupBy, now, sortBy).flatMap(bucket => bucket.entries);
  }

  const compare = todoComparator(sortBy);
  return workspaces.flatMap(workspace => entries
    .filter(entry => entry.workspace.dir === workspace.dir)
    .sort((a, b) => compare(a.todo, b.todo)));
}

// The user-chosen settings that order and filter the queue. Each field is a
// state value that keeps its identity until the user changes it.
export interface TodoNavigationView {
  filters: TodoFilters;
  groupBy: TodoGroupBy;
  sortBy: TodoSort;
  timeRange: TodoTimeRange | null;
  query: string;
}

// A TodoNavigationPin is the open todo as the list held it when it was opened.
// Navigation orders and filters against this copy rather than the live item, so
// editing a field the queue sorts or filters on (priority, status, labels) does
// not move the open todo away from its neighbours or drop it from the queue.
export interface TodoNavigationPin {
  dir: string;
  ref: string;
  todo: TodoItem;
  view: TodoNavigationView;
}

function sameView(a: TodoNavigationView, b: TodoNavigationView): boolean {
  return a.filters === b.filters && a.groupBy === b.groupBy && a.sortBy === b.sortBy
    && a.timeRange === b.timeRange && a.query === b.query;
}

// nextTodoNavigationPin returns `pin` itself while the same todo stays open
// under the same view, and otherwise re-captures the selected todo from the
// live list. Returning the same object when nothing changed is what lets a
// caller store the result in state during render without looping.
export function nextTodoNavigationPin(pin: TodoNavigationPin | null, { byDir, selected, view }: {
  byDir: Record<string, TodoListResponse>;
  selected: SelectedTodo | null;
  view: TodoNavigationView;
}): TodoNavigationPin | null {
  if (!selected) return null;
  if (pin && pin.dir === selected.dir && pin.ref === selected.ref && sameView(pin.view, view)) return pin;
  const todo = byDir[selected.dir]?.items.find(item => item.ref === selected.ref);
  return todo ? { dir: selected.dir, ref: selected.ref, todo, view } : null;
}

// withPinnedTodo substitutes the pinned copy for the live item. It returns
// `byDir` itself when there is nothing to substitute, including when the todo
// is gone from its list: a deleted todo must not be resurrected.
export function withPinnedTodo(
  byDir: Record<string, TodoListResponse>,
  pin: TodoNavigationPin | null,
): Record<string, TodoListResponse> {
  const list = pin ? byDir[pin.dir] : undefined;
  if (!pin || !list) return byDir;
  const index = list.items.findIndex(item => item.ref === pin.ref);
  if (index < 0 || list.items[index] === pin.todo) return byDir;
  const items = list.items.slice();
  items[index] = pin.todo;
  return { ...byDir, [pin.dir]: { ...list, items } };
}

function filterTokens(value: unknown): string[] {
  if (Array.isArray(value)) return value.flatMap(filterTokens).filter(Boolean);
  if (value == null) return [];
  if (typeof value === 'object') return [JSON.stringify(value)];
  const token = String(value).trim();
  return token ? [token] : [];
}

export function filterTodoNavigationEntries<T extends TodoEntry & Record<string, unknown>>(
  entries: TodoEntry[],
  columns: DataTableColumn<T>[],
  query: string,
): TodoEntry[] {
  const needle = query.trim().toLowerCase();
  if (!needle) return entries;
  const searchable = columns.filter(column => column.filterable !== false && column.kind !== 'timestamp');
  return entries.filter(entry => {
    const row = entry as T;
    const haystack = searchable.flatMap(column => {
      const value = column.accessor ? column.accessor(row) : row[column.key];
      return filterTokens(column.filterValue ? column.filterValue(value, row) : value);
    }).join(' ').toLowerCase();
    return haystack.includes(needle);
  });
}

export function todoNavigationState(entries: TodoEntry[], selected: SelectedTodo | null): TodoNavigationState | null {
  if (!selected) return null;
  const index = entries.findIndex(entry => entry.workspace.dir === selected.dir && entry.todo.ref === selected.ref);
  if (index < 0) return null;
  const target = (at: number): SelectedTodo | null => {
    const entry = entries[at];
    return entry ? { dir: entry.workspace.dir, ref: entry.todo.ref } : null;
  };
  return {
    position: index + 1,
    total: entries.length,
    previous: target(index - 1),
    next: target(index + 1),
  };
}
