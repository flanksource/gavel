import { describe, expect, it } from 'vitest';
import type { AgentSession } from '../../types';
import {
  DEFAULT_SESSION_FILTERS,
  attentionReasons,
  contextUsedPercent,
  sessionFamilies,
  sessionState,
  visibleFamilies,
  workStateBadge,
  type SessionFamily,
} from './sessionView';

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

type Outline = [string, Outline[]];
const outline = (families: SessionFamily[]): Outline[] => families.map(f => [f.session.title, outline(f.children)]);
const ago = (hours: number) => new Date(NOW - hours * HOUR).toISOString();

describe('sessionFamilies', () => {
  it('nests subagents, and theirs, under the session that delegated them', () => {
    const families = sessionFamilies([
      session('root'),
      session('child', { parentId: 'root' }),
      session('grandchild', { parentId: 'child' }),
      session('solo'),
    ]);

    expect(outline(families)).toEqual([['root', [['child', [['grandchild', []]]]]], ['solo', []]]);
  });

  it('fails on a subagent whose parent is not listed', () => {
    expect(() => sessionFamilies([session('orphan', { parentId: 'gone' })])).toThrow('names parent gone');
  });
});

describe('visibleFamilies', () => {
  const families = sessionFamilies([
    session('Quiet parent', { lastActivityAt: ago(5), startedAt: ago(9), costUsd: 4 }),
    session('Busy subagent', { parentId: 'Quiet parent', lastActivityAt: ago(1), activityState: 'ask', project: 'gavel' }),
    session('Recent root', { lastActivityAt: ago(2), startedAt: ago(3), costUsd: 1, project: 'gavel', worktree: { path: '/w', branch: 'fix/pgproxy', base: 'main', primary: false } }),
  ]);
  const unfiltered = DEFAULT_SESSION_FILTERS;

  it('orders families by the latest activity anywhere in them', () => {
    expect(outline(visibleFamilies(families, unfiltered))).toEqual([['Quiet parent', [['Busy subagent', []]]], ['Recent root', []]]);
  });

  it('keeps a matching subagent with its parent and drops non-matching siblings', () => {
    expect(outline(visibleFamilies(families, { ...unfiltered, state: { ask: 'include' } }))).toEqual([['Quiet parent', [['Busy subagent', []]]]]);
  });

  it('drops excluded states and projects, keeping sessions with no project when a project is only excluded', () => {
    expect({
      excludeAsk: outline(visibleFamilies(families, { ...unfiltered, state: { ask: 'exclude' } })),
      excludeGavel: outline(visibleFamilies(families, { ...unfiltered, project: { gavel: 'exclude' } })),
      includeGavelExcludeAsk: outline(visibleFamilies(families, { ...unfiltered, project: { gavel: 'include' }, state: { ask: 'exclude' } })),
    }).toEqual({
      excludeAsk: [['Recent root', []], ['Quiet parent', []]],
      excludeGavel: [['Quiet parent', []]],
      includeGavelExcludeAsk: [['Recent root', []]],
    });
  });

  it('matches search text against the branch as well as the title', () => {
    expect(outline(visibleFamilies(families, { ...unfiltered, search: 'PGPROXY' }))).toEqual([['Recent root', []]]);
  });

  it('sorts by start time and by cost', () => {
    expect(outline(visibleFamilies(families, { ...unfiltered, sort: 'started-desc' })).map(([title]) => title)).toEqual(['Recent root', 'Quiet parent']);
    expect(outline(visibleFamilies(families, { ...unfiltered, sort: 'cost-desc' })).map(([title]) => title)).toEqual(['Quiet parent', 'Recent root']);
  });
});

describe('sessionState', () => {
  it('prefers the input a session waits on over its live process, then how it ended', () => {
    expect([
      session('a', { activityState: 'approval', processActive: true }),
      session('b', { activityState: 'ask' }),
      session('c', { processActive: true }),
      session('d', { lifecycleStatus: 'interrupted' }),
      session('e'),
    ].map(sessionState)).toEqual(['approval', 'ask', 'working', 'failed', 'completed']);
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
    expect(attentionReasons(flagged, { work: false })).toEqual(['Needs approval', 'Session failed']);
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

describe('contextUsedPercent', () => {
  it('is the complement of the free share, undefined without context data', () => {
    expect(contextUsedPercent(session('ctx', { context: { usedTokens: 162_000, windowTokens: 200_000, freePercent: 19 } }))).toBe(81);
    expect(contextUsedPercent(session('none'))).toBeUndefined();
  });
});

describe('workStateBadge', () => {
  const worktree = { path: '/w', branch: 'fix/pgproxy', base: 'main', primary: false };
  const pr = { number: 12, url: '', state: 'open', checkStatus: 'none' as const };
  it.each([
    {
      name: 'conflicts before any other state',
      overrides: { worktree, git: { changes: { ...CLEAN, conflict: 2, untracked: 5 }, ahead: 3, behind: 0, baseCheckedOut: true, worktreeLive: true } },
      expected: { label: '2 conflicted', tone: 'warning' },
    },
    {
      name: 'uncommitted work before commits ahead',
      overrides: { worktree, git: { changes: { ...CLEAN, unstaged: 1, untracked: 157 }, ahead: 2874, behind: 72, baseCheckedOut: true, worktreeLive: true }, pr },
      expected: { label: '158 uncommitted', tone: 'warning' },
    },
    {
      name: 'commits ahead on a clean branch',
      overrides: { worktree, git: { changes: CLEAN, ahead: 4, behind: 0, baseCheckedOut: true, worktreeLive: true }, pr },
      expected: { label: '4 ahead' },
    },
    {
      name: 'the PR of a clean branch level with its base',
      overrides: { worktree, git: { changes: CLEAN, ahead: 0, behind: 3, baseCheckedOut: true, worktreeLive: true }, pr },
      expected: { label: 'PR #12' },
    },
    {
      name: 'a clean branch with no PR',
      overrides: { worktree, git: { changes: CLEAN, ahead: 0, behind: 3, baseCheckedOut: true, worktreeLive: true } },
      expected: { label: 'clean' },
    },
    {
      name: 'a session with no git state',
      overrides: { cwd: '/repo' },
      expected: { label: 'no git' },
    },
  ] satisfies { name: string; overrides: Partial<AgentSession>; expected: { label: string; tone?: 'warning' } }[])('shows $name', ({ overrides, expected }) => {
    expect(workStateBadge(session('s', overrides))).toEqual(expected);
  });
});
