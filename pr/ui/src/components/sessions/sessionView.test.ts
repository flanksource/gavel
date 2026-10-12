import { describe, expect, it } from 'vitest';
import type { AgentSession } from '../../types';
import { attentionReasons, contextUsedPercent, filterSessions, groupSessions } from './sessionView';

const NOW = Date.parse('2026-10-06T12:00:00Z');
const HOUR = 60 * 60 * 1000;
const CLEAN = { staged: 0, unstaged: 0, both: 0, untracked: 0, conflict: 0, adds: 0, dels: 0 };

function session(title: string, overrides: Partial<AgentSession> = {}): AgentSession {
  return {
    id: title,
    title,
    source: 'claude',
    lifecycleStatus: 'succeeded',
    activityState: 'idle',
    processActive: false,
    lastActivityAt: new Date(NOW - HOUR).toISOString(),
    tokens: { input: 0, output: 0, cacheRead: 0, total: 0 },
    costUsd: 0,
    ...overrides,
  };
}

describe('groupSessions', () => {
  it('puts live and input-waiting sessions in Active, then splits the rest at 24h, keeping server order', () => {
    const sessions = [
      session('asking', { activityState: 'ask', lastActivityAt: new Date(NOW - 30 * HOUR).toISOString() }),
      session('running', { processActive: true }),
      session('finished today'),
      session('finished last week', { lastActivityAt: new Date(NOW - 7 * 24 * HOUR).toISOString() }),
    ];
    const titles = (rows: AgentSession[]) => rows.map(row => row.title);

    const groups = groupSessions(sessions, NOW);

    expect({ active: titles(groups.active), recent: titles(groups.recent), older: titles(groups.older) }).toEqual({
      active: ['asking', 'running'],
      recent: ['finished today'],
      older: ['finished last week'],
    });
  });
});

describe('attentionReasons', () => {
  it('flags input waits, failures, conflicts, failing PR checks and abandoned dirty work', () => {
    const flagged = session('broken', {
      activityState: 'approval',
      lifecycleStatus: 'failed',
      git: { changes: { ...CLEAN, unstaged: 2, conflict: 1 }, ahead: 1, behind: 0, baseCheckedOut: true, worktreeLive: true },
      pr: { number: 12, url: '', state: 'open', checkStatus: 'failure' },
    });

    expect(attentionReasons(flagged)).toEqual([
      'Needs approval', 'Session failed', '1 conflicts', 'PR #12 checks failing', 'Uncommitted changes left behind',
    ]);
  });

  it('does not flag a live session for its uncommitted work, nor a merged PR for old check failures', () => {
    const fine = session('fine', {
      processActive: true,
      git: { changes: { ...CLEAN, unstaged: 4 }, ahead: 0, behind: 0, baseCheckedOut: true, worktreeLive: true },
      pr: { number: 3, url: '', state: 'merged', checkStatus: 'failure' },
    });

    expect(attentionReasons(fine)).toEqual([]);
  });
});

describe('filterSessions', () => {
  const sessions = [
    session('Fix sniffer', { project: 'commons-db', worktree: { path: '/w', branch: 'fix/pgproxy', base: 'main', primary: false } }),
    session('Sessions tab', { project: 'gavel', activityState: 'ask' }),
  ];

  it('matches search text against the branch as well as the title', () => {
    expect(filterSessions(sessions, { search: 'PGPROXY', project: '', attention: false }).map(s => s.title)).toEqual(['Fix sniffer']);
  });

  it('combines the project and needs-attention filters', () => {
    expect(filterSessions(sessions, { search: '', project: 'gavel', attention: true }).map(s => s.title)).toEqual(['Sessions tab']);
    expect(filterSessions(sessions, { search: '', project: 'commons-db', attention: true })).toEqual([]);
  });
});

describe('contextUsedPercent', () => {
  it('is the complement of the free share, undefined without context data', () => {
    expect(contextUsedPercent(session('ctx', { context: { usedTokens: 162_000, windowTokens: 200_000, freePercent: 19 } }))).toBe(81);
    expect(contextUsedPercent(session('none'))).toBeUndefined();
  });
});
