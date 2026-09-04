import type { DataTableColumn } from '@flanksource/clicky-ui/data';
import type { Project, TodoGroupBy, TodoListResponse } from '../../types';
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
