import { describe, expect, it } from 'vitest';
import type { TodoEvent, TodoRunWorkspace, TodoSessionAttempt } from '../../types';
import { continueBranchLabel, reusableRunBranch } from './runBranchChoice';

function attempt(ordinal: number, workspace?: TodoRunWorkspace, step = 'run'): TodoSessionAttempt {
  return {
    promptRunId: `run-${ordinal}`, ordinal, step, requested: {}, resolved: {}, status: 'completed', processActive: false,
    state: 'succeeded', phase: 'finished', queuedAt: '2026-09-14T10:00:00Z', admissionSessionId: `admission-${ordinal}`,
    createdAt: '2026-09-14T10:00:00Z', updatedAt: '2026-09-14T10:01:00Z', verification: null, workspace,
  };
}

function lineComment(id: string): TodoEvent {
  return { id, kind: 'comment', body: id, payload: { anchor: { path: 'a.go', side: 'new', line: 1, lineText: 'x', commit: 'abc' } } };
}

describe('reusableRunBranch', () => {
  const events = [lineComment('a'), lineComment('b'), { id: 'r', kind: 'comment_resolved', payload: { commentId: 'a', resolved: true } }];

  it('offers the latest run branch with its unresolved line comment count', () => {
    expect(reusableRunBranch([attempt(2, { worktree: { branch: 'shell/2' } }), attempt(1, { worktree: { branch: 'shell/1' } })], events))
      .toEqual({ branch: 'shell/2', unresolved: 1 });
  });

  it('offers nothing once teardown deleted the branch', () => {
    expect(reusableRunBranch([attempt(2, { worktree: { branch: 'shell/2', branchDeleted: true } })], events)).toBeNull();
  });

  it('offers nothing when no run recorded a worktree branch', () => {
    expect(reusableRunBranch([attempt(2, { cwd: '/repo' }), attempt(1)], events)).toBeNull();
    expect(reusableRunBranch([attempt(1, { worktree: { path: '/repo/.wt' } })], events)).toBeNull();
  });

  it('ignores verify steps, which never carry the run branch', () => {
    expect(reusableRunBranch([attempt(3, { worktree: { branch: 'shell/verify' } }, 'verify')], events)).toBeNull();
  });
});

describe('continueBranchLabel', () => {
  it('pluralises unresolved comments and omits the count at zero', () => {
    expect(continueBranchLabel({ branch: 'shell/1', unresolved: 0 })).toBe('Continue on shell/1');
    expect(continueBranchLabel({ branch: 'shell/1', unresolved: 1 })).toBe('Continue on shell/1 · 1 unresolved comment');
    expect(continueBranchLabel({ branch: 'shell/1', unresolved: 3 })).toBe('Continue on shell/1 · 3 unresolved comments');
  });
});
