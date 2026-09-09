import { beforeEach, describe, expect, it } from 'vitest';
import type { TodoItem } from '../../types';
import type { TodoSort } from './todoSort';
import { loadTodoSort, saveTodoSort, todoComparator } from './todoSort';

const items: TodoItem[] = [
  {
    ref: 'created-later',
    title: 'Created later',
    status: 'pending',
    priority: 'medium',
    created: '2026-07-20T10:00:00Z',
    lastRun: '2026-07-21T10:00:00Z',
  },
  {
    ref: 'updated-later',
    title: 'Updated later',
    status: 'pending',
    priority: 'medium',
    created: '2026-07-19T10:00:00Z',
    lastRun: '2026-07-22T10:00:00Z',
  },
];

beforeEach(() => {
  localStorage.clear();
});

describe('todoComparator', () => {
  it.each([
    { sortBy: { column: 'created', dir: 'desc' }, refs: ['created-later', 'updated-later'] },
    { sortBy: { column: 'created', dir: 'asc' }, refs: ['updated-later', 'created-later'] },
    { sortBy: { column: 'updated', dir: 'desc' }, refs: ['updated-later', 'created-later'] },
    { sortBy: { column: 'updated', dir: 'asc' }, refs: ['created-later', 'updated-later'] },
  ] satisfies { sortBy: TodoSort; refs: string[] }[])(
    'sorts $sortBy.column $sortBy.dir',
    ({ sortBy, refs }) => {
      expect([...items].sort(todoComparator(sortBy)).map(item => item.ref)).toEqual(refs);
    },
  );

  it.each([
    { column: 'created', dir: 'asc' },
    { column: 'created', dir: 'desc' },
    { column: 'updated', dir: 'asc' },
    { column: 'updated', dir: 'desc' },
  ] satisfies TodoSort[])('keeps missing $column timestamps last when sorting $dir', (sortBy) => {
    const missing: TodoItem = {
      ref: 'missing',
      title: 'Missing timestamp',
      status: 'pending',
      priority: 'medium',
    };

    expect([...items, missing].sort(todoComparator(sortBy)).at(-1)?.ref).toBe('missing');
  });

  it('sorts equal priorities by their update time in the selected direction', () => {
    expect([...items]
      .sort(todoComparator({ column: 'priority', dir: 'desc' }))
      .map(item => item.ref))
      .toEqual(['updated-later', 'created-later']);
  });
});

it('persists the sort column and direction together', () => {
  const sortBy: TodoSort = { column: 'updated', dir: 'asc' };

  saveTodoSort(sortBy);

  expect(loadTodoSort()).toEqual(sortBy);
});
