import type { DataTableColumn } from '@flanksource/clicky-ui/data';
import { describe, expect, it } from 'vitest';
import type { Project, TodoItem, TodoListResponse } from '../../types';
import { defaultTodoFilters } from './todoFilter';
import type { TodoTableRow } from './TodoTable';
import {
  filterTodoNavigationEntries,
  orderedTodoNavigationEntries,
  todoNavigationState,
} from './todoNavigation';

const alpha: Project = { name: 'Alpha', dir: '/repos/alpha', repos: [] } as Project;
const beta: Project = { name: 'Beta', dir: '/repos/beta', repos: [] } as Project;

function todo(ref: string, title: string, priority: TodoItem['priority'], lastRun?: string): TodoItem {
  return { ref, title, priority, status: 'pending', lastRun };
}

function response(dir: string, items: TodoItem[]): TodoListResponse {
  return { dir, items, counts: {} as TodoListResponse['counts'] };
}

describe('orderedTodoNavigationEntries', () => {
  const high = todo('a-high', 'Alpha high', 'high', '2026-09-03T08:00:00Z');
  const low = todo('a-low', 'Alpha low', 'low', '2026-09-03T07:00:00Z');
  const medium = todo('b-medium', 'Beta medium', 'medium', '2026-09-03T09:00:00Z');
  const byDir = {
    [alpha.dir]: response(alpha.dir, [low, high]),
    [beta.dir]: response(beta.dir, [medium]),
  };

  it('follows workspace order and the active row sort within each workspace', () => {
    const entries = orderedTodoNavigationEntries({
      workspaces: [alpha, beta],
      byDir,
      filters: defaultTodoFilters(),
      groupBy: 'workspace',
      sortBy: { column: 'priority', dir: 'desc' },
      timeRange: null,
      now: Date.parse('2026-09-03T10:00:00Z'),
    });

    expect(entries.map(entry => [entry.workspace.dir, entry.todo.ref])).toEqual([
      [alpha.dir, high.ref],
      [alpha.dir, low.ref],
      [beta.dir, medium.ref],
    ]);
  });

  it('follows bucket order after applying the current facets', () => {
    const entries = orderedTodoNavigationEntries({
      workspaces: [alpha, beta],
      byDir,
      filters: { ...defaultTodoFilters(), priorities: { low: 'exclude' } },
      groupBy: 'severity',
      sortBy: { column: 'title', dir: 'asc' },
      timeRange: null,
      now: Date.parse('2026-09-03T10:00:00Z'),
    });

    expect(entries.map(entry => entry.todo.ref)).toEqual([high.ref, medium.ref]);
  });
});

describe('filterTodoNavigationEntries', () => {
  const entries = [
    { workspace: alpha, todo: todo('todo-alpha', 'Deploy API', 'high', '2026-09-03T08:00:00Z') },
    { workspace: beta, todo: todo('todo-beta', 'Write docs', 'medium', '2026-09-02T08:00:00Z') },
  ];
  const columns: DataTableColumn<TodoTableRow>[] = [
    { key: 'title', label: 'Title', accessor: row => row.todo.title },
    { key: 'workspace', label: 'Workspace', accessor: row => row.workspace.name },
    { key: 'updated', label: 'Updated', kind: 'timestamp', accessor: row => row.todo.lastRun },
  ];

  it('uses the table columns and skips timestamp columns exactly like DataTable search', () => {
    expect(filterTodoNavigationEntries(entries, columns, 'beta').map(entry => entry.todo.ref)).toEqual(['todo-beta']);
    expect(filterTodoNavigationEntries(entries, columns, '2026-09-03')).toEqual([]);
  });
});

describe('todoNavigationState', () => {
  const sameRef = 'shared-ref';
  const entries = [
    { workspace: alpha, todo: todo(sameRef, 'Alpha todo', 'medium') },
    { workspace: beta, todo: todo(sameRef, 'Beta todo', 'medium') },
    { workspace: beta, todo: todo('last', 'Last todo', 'medium') },
  ];

  it('identifies the current todo by workspace and ref without wrapping', () => {
    expect(todoNavigationState(entries, { dir: beta.dir, ref: sameRef })).toEqual({
      position: 2,
      total: 3,
      previous: { dir: alpha.dir, ref: sameRef },
      next: { dir: beta.dir, ref: 'last' },
    });
    expect(todoNavigationState(entries, { dir: beta.dir, ref: 'last' })?.next).toBeNull();
  });

  it('hides navigation when a deep link is outside the current queue', () => {
    expect(todoNavigationState(entries, { dir: '/repos/elsewhere', ref: 'missing' })).toBeNull();
  });
});
