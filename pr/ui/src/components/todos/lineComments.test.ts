import { describe, expect, it } from 'vitest';
import type { TodoEvent } from '../../types';
import { lineComments, matchesRange, unresolvedLineCommentCount, type LineCommentAnchor } from './lineComments';

const anchor: LineCommentAnchor = {
  path: 'cmd/main.go', side: 'new', line: 12, lineText: 'return nil', base: 'setup1', commit: 'head1', branch: 'shell/1', attemptId: 'run-1',
};

function lineComment(id: string, overrides: Partial<LineCommentAnchor> = {}, timestamp = '2026-10-01T10:00:00Z'): TodoEvent {
  return { id, kind: 'comment', actor: 'moshe', timestamp, body: `body ${id}`, payload: { anchor: { ...anchor, ...overrides } } };
}

function resolution(commentId: string, resolved: boolean, timestamp: string): TodoEvent {
  return { id: `res-${commentId}-${timestamp}`, kind: 'comment_resolved', timestamp, payload: { commentId, resolved } };
}

describe('lineComments', () => {
  it('keeps only comments carrying an anchor, in event order, unresolved by default', () => {
    const plain: TodoEvent = { id: 'plain', kind: 'comment', body: 'just a comment' };
    const other: TodoEvent = { id: 'status', kind: 'status_changed' };
    const a = lineComment('a');
    const b = lineComment('b', { line: 30 });

    expect(lineComments([plain, a, other, b])).toEqual([
      { event: a, anchor: a.payload!.anchor, resolved: false },
      { event: b, anchor: b.payload!.anchor, resolved: false },
    ]);
  });

  it('lets the later resolution event win, so a comment can be resolved, reopened and resolved again', () => {
    const events = [
      lineComment('a'),
      resolution('a', true, '2026-10-01T11:00:00Z'),
      resolution('a', false, '2026-10-01T12:00:00Z'),
      resolution('a', true, '2026-10-01T13:00:00Z'),
      lineComment('b'),
      resolution('b', true, '2026-10-01T11:00:00Z'),
      resolution('b', false, '2026-10-01T12:00:00Z'),
    ];

    expect(lineComments(events).map(comment => [comment.event.id, comment.resolved])).toEqual([['a', true], ['b', false]]);
  });

  it('ignores a resolution that names no line comment', () => {
    expect(lineComments([lineComment('a'), resolution('missing', true, '2026-10-01T11:00:00Z')]).map(comment => comment.resolved)).toEqual([false]);
  });

  it('throws on an anchored comment without an id because it could never be resolved', () => {
    expect(() => lineComments([{ kind: 'comment', payload: { anchor } }])).toThrow(/without an id/);
  });

  it('throws naming the missing field when an anchor is malformed', () => {
    const { commit: _commit, ...broken } = anchor;
    expect(() => lineComments([{ id: 'x', kind: 'comment', payload: { anchor: broken } }])).toThrow(/commit/);
    expect(() => lineComments([{ id: 'x', kind: 'comment', payload: { anchor: { ...anchor, side: 'middle' } } }])).toThrow(/side/);
    expect(() => lineComments([{ id: 'x', kind: 'comment', payload: { anchor: { ...anchor, line: '12' } } }])).toThrow(/line/);
  });

  it('throws when a resolution payload has no boolean resolved flag', () => {
    expect(() => lineComments([{ id: 'r', kind: 'comment_resolved', payload: { commentId: 'a' } }])).toThrow(/resolved/);
  });
});

describe('unresolvedLineCommentCount', () => {
  it('counts only unresolved anchored comments', () => {
    const events = [lineComment('a'), lineComment('b'), resolution('a', true, '2026-10-01T11:00:00Z'), { id: 'p', kind: 'comment', body: 'plain' }];
    expect(unresolvedLineCommentCount(events)).toBe(1);
  });
});

describe('matchesRange', () => {
  it('compares path, base and commit, treating an absent base as the empty single-commit base', () => {
    expect(matchesRange(anchor, { path: 'cmd/main.go', base: 'setup1', commit: 'head1' })).toBe(true);
    expect(matchesRange(anchor, { path: 'cmd/other.go', base: 'setup1', commit: 'head1' })).toBe(false);
    expect(matchesRange(anchor, { path: 'cmd/main.go', base: '', commit: 'head1' })).toBe(false);
    expect(matchesRange(anchor, { path: 'cmd/main.go', base: 'setup1', commit: 'head2' })).toBe(false);
    const single = { ...anchor, base: undefined };
    expect(matchesRange(single, { path: 'cmd/main.go', base: '', commit: 'head1' })).toBe(true);
  });
});
