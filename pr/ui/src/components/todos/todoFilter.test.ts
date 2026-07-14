import { describe, expect, it } from 'vitest';
import type { TodoItem } from '../../types';
import { todoMatchesQuery } from './todoFilter';

const todo: TodoItem = {
  ref: '3785b0a4-0bf6-4f65-b1c2-41eab73e9f6b',
  id: '3785b0a4-0bf6-4f65-b1c2-41eab73e9f6b',
  title: 'Parse user shell commands',
  status: 'completed',
  priority: 'medium',
  sessionId: '019f5b2e-75b7-7de2-911b-de8b70266479',
};

describe('todoMatchesQuery', () => {
  it('matches both Todo and current session UUID fragments', () => {
    expect(todoMatchesQuery(todo, '3785b0a4')).toBe(true);
    expect(todoMatchesQuery(todo, '019F5B2E-75B7')).toBe(true);
  });

  it('does not treat an unrelated UUID fragment as a match', () => {
    expect(todoMatchesQuery(todo, 'aaaaaaaa-bbbb')).toBe(false);
  });
});
