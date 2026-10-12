import type { DataTableColumn } from '@flanksource/clicky-ui/data';
import { describe, expect, it } from 'vitest';
import type { Project, TodoItem, TodoListResponse } from '../../types';
import { defaultTodoFilters } from './todoFilter';
import type { TodoTableRow } from './TodoTable';
import {
  filterTodoNavigationEntries,
  nextTodoNavigationPin,
  orderedTodoNavigationEntries,
  siblingNavigationEntries,
  todoNavigationState,
  withPinnedTodo,
  type TodoNavigationView,
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
      includeChildren: false,
    });

    expect(entries.map(entry => [entry.workspace.dir, entry.todo.ref])).toEqual([
      [alpha.dir, high.ref],
      [alpha.dir, low.ref],
      [beta.dir, medium.ref],
    ]);
  });

  describe('children', () => {
    const child = { ...todo('a-child', 'Alpha child', 'high', '2026-09-03T06:00:00Z'), parentId: 'parent-id' };
    const withChild = {
      [alpha.dir]: response(alpha.dir, [low, child, high]),
      [beta.dir]: response(beta.dir, [medium]),
    };
    const queue = (groupBy: 'workspace' | 'severity', includeChildren: boolean) => orderedTodoNavigationEntries({
      workspaces: [alpha, beta],
      byDir: withChild,
      filters: defaultTodoFilters(),
      groupBy,
      sortBy: { column: 'title', dir: 'asc' },
      timeRange: null,
      now: Date.parse('2026-09-03T10:00:00Z'),
      includeChildren,
    }).map(entry => entry.todo.ref);

    it.each(['workspace', 'severity'] as const)('leaves a child out of the %s queue unless children are included', groupBy => {
      expect(queue(groupBy, false)).not.toContain(child.ref);
      expect(queue(groupBy, true)).toContain(child.ref);
    });

    it('keeps every top-level todo in the queue when children are included', () => {
      expect(queue('workspace', true)).toEqual(['a-child', 'a-high', 'a-low', 'b-medium']);
      expect(queue('workspace', false)).toEqual(['a-high', 'a-low', 'b-medium']);
    });
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
      includeChildren: false,
    });

    expect(entries.map(entry => entry.todo.ref)).toEqual([high.ref, medium.ref]);
  });
});

describe('siblingNavigationEntries', () => {
  const parentId = '8d2f6a41-3c59-4e07-b1a8-6f0e9c7d5b32';
  const parent = { ...todo('a-parent', 'Parent', 'medium'), id: parentId };
  const first = { ...todo('a-first', 'First child', 'low'), parentId };
  const second = { ...todo('a-second', 'Second child', 'high'), parentId, status: 'completed' as const };
  const third = { ...todo('a-third', 'Third child', 'medium'), parentId };
  const stranger = { ...todo('a-stranger', 'Other parent child', 'medium'), parentId: 'some-other-parent' };
  const byDir = { [alpha.dir]: response(alpha.dir, [parent, third, stranger, first, second]), [beta.dir]: response(beta.dir, []) };
  const siblings = (ref: string, dir = alpha.dir) => siblingNavigationEntries({ workspaces: [alpha, beta], byDir, selected: { dir, ref } });

  it('is the parent\'s children in list order, closed ones included, for a selected child', () => {
    expect(siblings('a-first')?.map(entry => [entry.workspace.dir, entry.todo.ref])).toEqual([
      [alpha.dir, 'a-third'],
      [alpha.dir, 'a-first'],
      [alpha.dir, 'a-second'],
    ]);
  });

  it('walks the siblings through the navigation state', () => {
    expect(todoNavigationState(siblings('a-first')!, { dir: alpha.dir, ref: 'a-first' })).toEqual({
      position: 2,
      total: 3,
      previous: { dir: alpha.dir, ref: 'a-third' },
      next: { dir: alpha.dir, ref: 'a-second' },
    });
  });

  it.each([
    ['a top-level todo', 'a-parent', alpha.dir],
    ['a todo that is not in the list', 'missing', alpha.dir],
    ['a todo in another workspace', 'a-first', beta.dir],
  ])('is null for %s, leaving the ordinary queue in charge', (_name, ref, dir) => {
    expect(siblings(ref, dir)).toBeNull();
  });

  it('is null with nothing selected', () => {
    expect(siblingNavigationEntries({ workspaces: [alpha, beta], byDir, selected: null })).toBeNull();
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

describe('navigation pinned to the open todo as it was when selected', () => {
  const first = todo('first', 'First', 'medium', '2026-09-03T09:00:00Z');
  const middle = todo('middle', 'Middle', 'medium', '2026-09-03T08:00:00Z');
  const last = todo('last', 'Last', 'medium', '2026-09-03T07:00:00Z');
  const selected = { dir: alpha.dir, ref: middle.ref };
  const view: TodoNavigationView = {
    filters: defaultTodoFilters(),
    groupBy: 'workspace',
    sortBy: { column: 'priority', dir: 'desc' },
    timeRange: null,
    query: '',
  };
  const byDirWith = (edited: TodoItem) => ({ [alpha.dir]: response(alpha.dir, [first, edited, last]) });
  const original = byDirWith(middle);
  const queue = (byDir: Record<string, TodoListResponse>) => orderedTodoNavigationEntries({
    workspaces: [alpha],
    byDir,
    filters: view.filters,
    groupBy: view.groupBy,
    sortBy: view.sortBy,
    timeRange: view.timeRange,
    now: Date.parse('2026-09-03T10:00:00Z'),
    includeChildren: false,
  });
  const pin = nextTodoNavigationPin(null, { byDir: original, selected, view });

  it('captures the selected todo from the list it was opened from', () => {
    expect(pin).toEqual({ dir: alpha.dir, ref: middle.ref, todo: middle, view });
  });

  it('keeps the same pin while the same todo stays open under the same view', () => {
    const edited = byDirWith({ ...middle, priority: 'high' });
    expect(nextTodoNavigationPin(pin, { byDir: edited, selected: { ...selected }, view: { ...view } })).toBe(pin);
  });

  it('re-captures when another todo is selected or the view changes', () => {
    expect(nextTodoNavigationPin(pin, { byDir: original, selected: { dir: alpha.dir, ref: last.ref }, view })?.todo).toBe(last);
    const resorted = { ...view, sortBy: { column: 'title', dir: 'asc' } as const };
    expect(nextTodoNavigationPin(pin, { byDir: original, selected, view: resorted })).toEqual({ ...pin, view: resorted });
    const refiltered = { ...view, filters: defaultTodoFilters() };
    expect(nextTodoNavigationPin(pin, { byDir: original, selected, view: refiltered })).not.toBe(pin);
  });

  it('re-captures when the open todo is attached to or detached from a parent, so the queue follows where it now lives', () => {
    const child = { ...middle, parentId: 'parent-id' };
    const childPin = nextTodoNavigationPin(null, { byDir: byDirWith(child), selected, view })!;
    expect(childPin.todo).toBe(child);

    const detached = byDirWith(middle);
    const next = nextTodoNavigationPin(childPin, { byDir: detached, selected, view });
    expect(next).not.toBe(childPin);
    expect(next?.todo).toBe(middle);
    expect(todoNavigationState(queue(withPinnedTodo(detached, next)), selected)?.position).toBe(2);
  });

  it('pins nothing until the selected todo is in the list', () => {
    expect(nextTodoNavigationPin(null, { byDir: original, selected: { dir: beta.dir, ref: middle.ref }, view })).toBeNull();
    expect(nextTodoNavigationPin(pin, { byDir: original, selected: null, view })).toBeNull();
  });

  it('leaves the list untouched while the open todo is unchanged or gone', () => {
    expect(withPinnedTodo(original, pin)).toBe(original);
    expect(withPinnedTodo(original, null)).toBe(original);
    const deleted = { [alpha.dir]: response(alpha.dir, [first, last]) };
    expect(withPinnedTodo(deleted, pin)).toBe(deleted);
  });

  it('keeps the list slot and neighbours of a todo whose priority was raised', () => {
    const edited = byDirWith({ ...middle, priority: 'high' });
    expect(queue(edited)[0].todo.ref).toBe(middle.ref);
    expect(withPinnedTodo(edited, pin)[alpha.dir].items).toEqual([first, middle, last]);
    expect(todoNavigationState(queue(withPinnedTodo(edited, pin)), selected)).toEqual({
      position: 2,
      total: 3,
      previous: { dir: alpha.dir, ref: first.ref },
      next: { dir: alpha.dir, ref: last.ref },
    });
  });

  it('keeps a todo in the queue after an edit the active filter hides', () => {
    const edited = byDirWith({ ...middle, status: 'completed' });
    expect(todoNavigationState(queue(edited), selected)).toBeNull();
    expect(todoNavigationState(queue(withPinnedTodo(edited, pin)), selected)).toEqual({
      position: 2,
      total: 3,
      previous: { dir: alpha.dir, ref: first.ref },
      next: { dir: alpha.dir, ref: last.ref },
    });
  });
});
