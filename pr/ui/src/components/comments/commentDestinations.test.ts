import { describe, expect, it } from 'vitest';
import type { GavelAgentSession } from '@flanksource/clicky-ui/comments/gavel';
import type { TodoItem } from '../../types';
import { commentDestinations } from './commentDestinations';

const session = (id: string, status: string): GavelAgentSession => ({ id, source: 'claude', title: `Session ${id}`, status, inCmux: false });
const todo = (ref: string, extra: Partial<TodoItem>): TodoItem => ({ ref, title: `Todo ${ref}`, status: 'in_progress', priority: 'medium', ...extra } as TodoItem);

describe('commentDestinations', () => {
  it('lists running sessions, then todos that ran newest first, then the project rule', () => {
    const destinations = commentDestinations({
      sessions: [session('live', 'active'), session('done', 'ended'), session('idle', 'sleeping')],
      todos: [
        todo('older', { lastRun: '2026-11-01T10:00:00Z' }),
        todo('never-ran', { created: '2026-11-20T10:00:00Z' }),
        todo('newer', { lastRun: '2026-11-20T10:00:00Z' }),
        todo('session-only', { sessionId: 's-1', created: '2026-10-01T10:00:00Z' }),
      ],
    });
    expect(destinations.map(destination => destination.id)).toEqual([
      'session:live',
      'session:idle',
      'todo:newer',
      'todo:older',
      'todo:session-only',
      'rule',
    ]);
  });

  it('falls back to the project rule when nothing is running and no todo has run', () => {
    const destinations = commentDestinations({ sessions: [session('done', 'ended')], todos: [todo('never-ran', {})] });
    expect(destinations.map(destination => destination.kind)).toEqual(['rule']);
  });
});
