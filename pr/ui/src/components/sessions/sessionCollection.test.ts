import { describe, expect, it } from 'vitest';
import type { AgentSession } from '../../types';
import { sessionCollection } from './sessionCollection';

function session(id: string, parentId?: string): AgentSession {
  return {
    id, title: `Session ${id}`, source: 'claude', provider: 'anthropic', model: 'opus', effort: 'high',
    lifecycleStatus: 'succeeded', activityState: 'idle', processActive: false, pid: 7, durationMs: 60_000,
    lastActivityAt: '2026-11-18T12:00:00Z', tokens: { input: 0, output: 0, cacheRead: 0, total: 0 }, costUsd: 1.5,
    ...(parentId ? { parentId } : {}),
  };
}

const root = session('root');
const child = session('child', 'root');
const grandchild = session('grandchild', 'child');
const unrelated = session('unrelated');
const all = [unrelated, grandchild, root, child];

describe('sessionCollection', () => {
  it('makes the session current and nests every descendant, each loaded from Captain', () => {
    const collection = sessionCollection(root, all);

    expect(collection.currentSessionId).toBe('root');
    expect(collection.sessions.map(({ id, parentId, src, label }) => ({ id, parentId, src, label }))).toEqual([
      { id: 'root', parentId: undefined, src: '/api/captain/sessions/root', label: 'Session root' },
      { id: 'child', parentId: 'root', src: '/api/captain/sessions/child', label: 'Session child' },
      { id: 'grandchild', parentId: 'child', src: '/api/captain/sessions/grandchild', label: 'Session grandchild' },
    ]);
  });

  it('roots a sub-session opened on its own at itself, dropping its parent link', () => {
    const collection = sessionCollection(child, all);

    expect(collection.sessions.map(({ id, parentId }) => ({ id, parentId }))).toEqual([
      { id: 'child', parentId: undefined },
      { id: 'grandchild', parentId: 'child' },
    ]);
  });

  it('summarises each item for the session content menu', () => {
    expect(sessionCollection(root, all).sessions[0]!.summary).toEqual({
      provider: 'anthropic', model: 'opus', effort: 'high', status: 'completed', pid: 7, durationMs: 60_000,
      updatedAt: '2026-11-18T12:00:00Z', cost: 1.5,
    });
  });
});
