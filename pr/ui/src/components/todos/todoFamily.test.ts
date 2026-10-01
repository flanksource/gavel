import { describe, expect, it } from 'vitest';
import type { TodoItem, TodoPhase, TodoPhaseRun } from '../../types';
import {
  childCountsById,
  childHasLivePhase,
  childPhaseTotals,
  childRollup,
  childrenOf,
  openChildrenOf,
  parentCandidates,
  parentOf,
  parentTitlesById,
} from './todoFamily';

const NOW = Date.parse('2026-09-30T12:00:00Z');
const PARENT_ID = '11111111-1111-4111-8111-111111111111';
const OTHER_ID = '22222222-2222-4222-8222-222222222222';

function todo(ref: string, overrides: Partial<TodoItem> = {}): TodoItem {
  return { ref, id: `id-${ref}`, title: `Todo ${ref}`, status: 'pending', priority: 'medium', ...overrides };
}

function run(phase: TodoPhase, overrides: Partial<TodoPhaseRun> = {}): TodoPhaseRun {
  return { phase, state: 'succeeded', ...overrides };
}

describe('childrenOf', () => {
  const parent = todo('parent', { id: PARENT_ID });
  const first = todo('first', { parentId: PARENT_ID });
  const second = todo('second', { parentId: PARENT_ID });
  const foreign = todo('foreign', { parentId: OTHER_ID });
  const orphanless = todo('plain');
  const items = [foreign, first, parent, orphanless, second];

  it('returns only the items whose parentId is the given id, in list order', () => {
    expect(childrenOf(items, PARENT_ID).map(item => item.ref)).toEqual(['first', 'second']);
  });

  it('returns nothing for a todo with no children or no id', () => {
    expect(childrenOf(items, '33333333-3333-4333-8333-333333333333')).toEqual([]);
    expect(childrenOf(items, undefined)).toEqual([]);
  });
});

describe('childPhaseTotals', () => {
  it('sums duration and cost across the recorded phases only', () => {
    const child = todo('c', {
      phases: {
        plan: run('plan', { duration_ms: 60_000, cost_usd: 0.5 }),
        verify: run('verify', { duration_ms: 30_000, cost_usd: 0.25 }),
      },
    });
    expect(childPhaseTotals(child, NOW)).toEqual({ durationMs: 90_000, costUsd: 0.75 });
  });

  it('measures a running phase against the clock rather than its stale snapshot', () => {
    const child = todo('c', {
      phases: {
        run: run('run', { state: 'running', started_at: '2026-09-30T11:59:15Z', duration_ms: 10_000, cost_usd: 0.1 }),
      },
    });
    expect(childPhaseTotals(child, NOW)).toEqual({ durationMs: 45_000, costUsd: 0.1 });
  });

  it('is zero for a child that has never run a phase', () => {
    expect(childPhaseTotals(todo('c'), NOW)).toEqual({ durationMs: 0, costUsd: 0 });
  });

  // A run that verified in-run is reported under both phases, each carrying the
  // whole run's duration and cost.
  describe('a run reported under both run and verify', () => {
    const RUN_ID = 'a1b2c3d4-0000-4000-8000-00000000aaaa';
    const VERIFY_ID = 'a1b2c3d4-0000-4000-8000-00000000bbbb';

    it('counts a settled run once when run and verify share its prompt run id', () => {
      const child = todo('c', {
        phases: {
          run: run('run', { prompt_run_id: RUN_ID, duration_ms: 120_000, cost_usd: 1.5 }),
          verify: run('verify', { prompt_run_id: RUN_ID, duration_ms: 120_000, cost_usd: 1.5 }),
        },
      });
      expect(childPhaseTotals(child, NOW)).toEqual({ durationMs: 120_000, costUsd: 1.5 });
    });

    it('counts a live run once and ticks it once', () => {
      const live = { state: 'running' as const, started_at: '2026-09-30T11:59:15Z', duration_ms: 10_000, cost_usd: 0.1 };
      const child = todo('c', {
        phases: {
          run: run('run', { ...live, prompt_run_id: RUN_ID }),
          verify: run('verify', { ...live, prompt_run_id: RUN_ID }),
        },
      });
      expect(childPhaseTotals(child, NOW)).toEqual({ durationMs: 45_000, costUsd: 0.1 });
    });

    it('counts a run and a standalone verify with different ids separately', () => {
      const child = todo('c', {
        phases: {
          run: run('run', { prompt_run_id: RUN_ID, duration_ms: 120_000, cost_usd: 1.5 }),
          verify: run('verify', { prompt_run_id: VERIFY_ID, duration_ms: 30_000, cost_usd: 0.25 }),
        },
      });
      expect(childPhaseTotals(child, NOW)).toEqual({ durationMs: 150_000, costUsd: 1.75 });
    });

    it('counts every phase that carries no id on its own', () => {
      const child = todo('c', {
        phases: {
          run: run('run', { duration_ms: 120_000, cost_usd: 1.5 }),
          verify: run('verify', { duration_ms: 30_000, cost_usd: 0.25 }),
        },
      });
      expect(childPhaseTotals(child, NOW)).toEqual({ durationMs: 150_000, costUsd: 1.75 });
    });
  });
});

describe('openChildrenOf', () => {
  it('keeps the children that are not completed — archived children read as completed', () => {
    const children = [
      todo('a', { status: 'pending' }),
      todo('b', { status: 'completed' }),
      todo('c', { status: 'in_progress' }),
      todo('d', { status: 'verified' }),
      todo('e', { status: 'skipped' }),
    ];
    expect(openChildrenOf(children).map(item => item.ref)).toEqual(['a', 'c', 'd', 'e']);
  });

  it('is empty when every child is closed', () => {
    expect(openChildrenOf([todo('a', { status: 'completed' })])).toEqual([]);
  });
});

describe('childHasLivePhase', () => {
  it.each([
    ['a running phase with a start time', { run: run('run', { state: 'running', started_at: '2026-09-30T11:59:15Z' }) }, true],
    ['a queued phase that has not started', { run: run('run', { state: 'pending' }) }, false],
    ['only settled phases', { run: run('run'), verify: run('verify', { state: 'failed' }) }, false],
    ['no phases', undefined, false],
  ])('is %s', (_name, phases, expected) => {
    expect(childHasLivePhase(todo('c', { phases }))).toBe(expected);
  });
});

describe('childRollup', () => {
  const completed = todo('completed', {
    status: 'completed',
    phases: {
      plan: run('plan', { duration_ms: 60_000, cost_usd: 0.5 }),
      run: run('run', { duration_ms: 120_000, cost_usd: 1.25 }),
    },
  });
  const verified = todo('verified', {
    status: 'verified',
    phases: { verify: run('verify', { duration_ms: 30_000, cost_usd: 0.25 }) },
  });
  const noPhases = todo('idle', { status: 'pending' });
  const live = todo('live', {
    status: 'in_progress',
    phases: {
      run: run('run', { state: 'running', started_at: '2026-09-30T11:59:15Z', duration_ms: 10_000, cost_usd: 0.1 }),
    },
  });

  it('counts completed and verified as done and sums the latest run of every phase, ticking a running one', () => {
    const rollup = childRollup([completed, verified, noPhases, live], NOW);
    expect(rollup.done).toBe(2);
    expect(rollup.total).toBe(4);
    expect(rollup.durationMs).toBe(60_000 + 120_000 + 30_000 + 45_000);
    expect(rollup.costUsd).toBeCloseTo(0.5 + 1.25 + 0.25 + 0.1, 10);
  });

  it('does not count a failed or unverified child as done', () => {
    const rollup = childRollup([todo('a', { status: 'failed' }), todo('b', { status: 'unverified' })], NOW);
    expect(rollup).toEqual({ done: 0, total: 2, durationMs: 0, costUsd: 0 });
  });

  it('is all zeros for no children', () => {
    expect(childRollup([], NOW)).toEqual({ done: 0, total: 0, durationMs: 0, costUsd: 0 });
  });
});

describe('parentOf and parentCandidates', () => {
  const parent = todo('parent', { id: PARENT_ID });
  const child = todo('child', { parentId: PARENT_ID });
  const sibling = todo('sibling', { parentId: PARENT_ID });
  const loner = todo('loner');
  const noId = todo('noid', { id: undefined });
  const items = [parent, child, sibling, loner, noId];

  it('resolves the parent by full id', () => {
    expect(parentOf(items, child)).toBe(parent);
  });

  it('has no parent for a top-level todo or when the parent is not in the list', () => {
    expect(parentOf(items, loner)).toBeNull();
    expect(parentOf([child], child)).toBeNull();
  });

  it('offers only top-level todos other than the todo itself', () => {
    expect(parentCandidates(items, loner).map(item => item.ref)).toEqual(['parent']);
    expect(parentCandidates(items, child).map(item => item.ref)).toEqual(['parent', 'loner']);
  });

  it('does not offer a closed todo as a parent', () => {
    const closed = todo('closed', { status: 'completed' });
    expect(parentCandidates([...items, closed], loner).map(item => item.ref)).toEqual(['parent']);
  });
});

describe('childCountsById', () => {
  it('counts each parent its children and the done ones, leaving childless todos out', () => {
    const counts = childCountsById([
      todo('parent', { id: PARENT_ID }),
      todo('idle'),
      todo('a', { parentId: PARENT_ID, status: 'completed' }),
      todo('b', { parentId: PARENT_ID, status: 'verified' }),
      todo('c', { parentId: PARENT_ID, status: 'failed' }),
      todo('d', { parentId: OTHER_ID }),
    ]);
    expect([...counts]).toEqual([
      [PARENT_ID, { done: 2, total: 3 }],
      [OTHER_ID, { done: 0, total: 1 }],
    ]);
  });
});

describe('parentTitlesById', () => {
  it('indexes the titles of the todos that have children, across workspaces', () => {
    const titles = parentTitlesById({
      '/a': {
        dir: '/a',
        counts: {} as never,
        items: [todo('one', { id: PARENT_ID, title: 'One' }), todo('idle', { title: 'No children' }), todo('kid', { parentId: PARENT_ID })],
      },
      '/b': { dir: '/b', counts: {} as never, items: [todo('two', { id: OTHER_ID, title: 'Two' }), todo('kid2', { parentId: OTHER_ID })] },
    });
    expect([...titles]).toEqual([[PARENT_ID, 'One'], [OTHER_ID, 'Two']]);
  });
});
