import type { ComponentType } from 'react';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiAdd, UiHistory, UiListFlat, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import type { TodoItem } from '../../types';

export type TodoSortColumn = 'priority' | 'updated' | 'created' | 'title';
export type TodoSortDirection = 'asc' | 'desc';

export interface TodoSort {
  column: TodoSortColumn;
  dir: TodoSortDirection;
}

export const TODO_SORT_COLUMN_OPTIONS: {
  value: TodoSortColumn;
  label: string;
  icon: ComponentType<IconProps>;
}[] = [
  { value: 'priority', label: 'Priority', icon: UiWarningTriangle },
  { value: 'updated', label: 'Updated', icon: UiHistory },
  { value: 'created', label: 'Created', icon: UiAdd },
  { value: 'title', label: 'Title', icon: UiListFlat },
];

const SORT_STORAGE_KEY = 'gavel.pr-ui.todoSort.v2';
const PRIORITY_VALUE: Record<string, number> = { low: 0, medium: 1, high: 2 };

export function defaultTodoSort(): TodoSort {
  return { column: 'priority', dir: 'desc' };
}

function isTodoSort(value: unknown): value is TodoSort {
  if (!value || typeof value !== 'object') return false;
  const sort = value as Partial<TodoSort>;
  return TODO_SORT_COLUMN_OPTIONS.some(option => option.value === sort.column)
    && (sort.dir === 'asc' || sort.dir === 'desc');
}

export function loadTodoSort(): TodoSort {
  try {
    const raw = localStorage.getItem(SORT_STORAGE_KEY);
    if (!raw) return defaultTodoSort();
    const value: unknown = JSON.parse(raw);
    return isTodoSort(value) ? value : defaultTodoSort();
  } catch {
    return defaultTodoSort();
  }
}

export function saveTodoSort(sort: TodoSort): void {
  try {
    localStorage.setItem(SORT_STORAGE_KEY, JSON.stringify(sort));
  } catch {
    // Storage is an optional view preference.
  }
}

function timestamp(value?: string): number | null {
  if (!value) return null;
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? null : parsed;
}

function compareOptionalNumbers(a: number | null, b: number | null, dir: TodoSortDirection): number {
  if (a === b) return 0;
  if (a === null) return 1;
  if (b === null) return -1;
  return dir === 'asc' ? a - b : b - a;
}

export function todoComparator(sort: TodoSort): (a: TodoItem, b: TodoItem) => number {
  return (a, b) => {
    let compared: number;
    switch (sort.column) {
      case 'created':
        compared = compareOptionalNumbers(timestamp(a.created), timestamp(b.created), sort.dir);
        break;
      case 'updated':
        compared = compareOptionalNumbers(
          timestamp(a.lastRun ?? a.created),
          timestamp(b.lastRun ?? b.created),
          sort.dir,
        );
        break;
      case 'title':
        compared = sort.dir === 'asc'
          ? a.title.localeCompare(b.title)
          : b.title.localeCompare(a.title);
        break;
      case 'priority':
        compared = compareOptionalNumbers(
          PRIORITY_VALUE[a.priority] ?? PRIORITY_VALUE.medium,
          PRIORITY_VALUE[b.priority] ?? PRIORITY_VALUE.medium,
          sort.dir,
        );
        if (compared === 0) {
          compared = compareOptionalNumbers(
            timestamp(a.lastRun ?? a.created),
            timestamp(b.lastRun ?? b.created),
            sort.dir,
          );
        }
        break;
    }
    return compared || a.title.localeCompare(b.title);
  };
}
