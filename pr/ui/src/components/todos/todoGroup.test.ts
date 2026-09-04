import { describe, expect, it } from 'vitest';
import type { Project, TodoItem, TodoPriority } from '../../types';
import { bucketTodos, type TodoEntry } from './todoGroup';

const workspace: Project = { name: 'demo', dir: '/repos/demo' } as Project;

function entry(ref: string, priority: TodoPriority, lastRun?: string): TodoEntry {
  const todo: TodoItem = {
    ref,
    title: ref,
    status: 'pending',
    priority,
    created: '2026-07-01T00:00:00Z',
    lastRun,
  };
  return { todo, workspace };
}

describe('bucketTodos with no grouping', () => {
  const entries = [
    entry('low-one', 'low'),
    entry('high-one', 'high'),
    entry('medium-one', 'medium'),
  ];
  const now = Date.parse('2026-07-10T00:00:00Z');

  it('collapses every entry into a single bucket', () => {
    const buckets = bucketTodos(entries, 'none', now);
    expect(buckets).toHaveLength(1);
    expect(buckets[0].key).toBe('all');
    expect(buckets[0].entries).toHaveLength(entries.length);
  });

  // The flat bucket is still a list, so it has to honour the sort preference —
  // otherwise "no grouping" would silently mean "no ordering" as well.
  it('orders the flat bucket by the requested sort', () => {
    const [bucket] = bucketTodos(entries, 'none', now, { column: 'title', dir: 'asc' });
    expect(bucket.entries.map(e => e.todo.ref)).toEqual(['high-one', 'low-one', 'medium-one']);

    const [byPriority] = bucketTodos(entries, 'none', now, { column: 'priority', dir: 'desc' });
    expect(byPriority.entries.map(e => e.todo.priority)).toEqual(['high', 'medium', 'low']);
  });

  it('yields no buckets when there is nothing to show', () => {
    expect(bucketTodos([], 'none', now)).toEqual([]);
  });
});
