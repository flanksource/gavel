import { describe, expect, it } from 'vitest';
import type { TaskSnapshot } from '@flanksource/clicky-ui/data';
import { projectCommitLockedFiles } from './project-commit-tasks';

const group = (details?: Record<string, unknown>): TaskSnapshot => ({
  id: 'commit-run-1',
  groupId: 'commit-run-1',
  name: 'Commit gavel',
  type: 'group',
  status: 'running',
  ...(details ? { details } : {}),
});

const task = (id: string, status: string): TaskSnapshot => ({
  id,
  groupId: 'commit-run-1',
  name: id,
  type: 'task',
  status,
});

describe('projectCommitLockedFiles', () => {
  it('locks files for active tasks using their stable generation position', () => {
    const locked = projectCommitLockedFiles([
      group({
        entries: [
          { taskId: 'done', action: 'commit', files: ['done.go'] },
          { taskId: 'running', action: 'commit', files: ['running.go', 'shared.go'] },
          { taskId: 'pending', action: 'open-pr', files: ['queued.go'] },
          { taskId: 'canceled', action: 'commit', files: ['canceled.go'] },
        ],
      }),
      task('done', 'success'),
      task('running', 'running'),
      task('pending', 'pending'),
      task('canceled', 'canceled'),
    ]);

    expect([...locked.entries()]).toEqual([
      ['running.go', 2],
      ['shared.go', 2],
      ['queued.go', 3],
    ]);
  });

  it('rejects active tasks whose file ownership metadata is missing', () => {
    expect(() => projectCommitLockedFiles([
      group(),
      task('running', 'running'),
    ])).toThrow('missing file ownership metadata');
  });

  it('does not require ownership metadata from a terminal historical run', () => {
    expect(projectCommitLockedFiles([
      { ...group(), status: 'success' },
      task('done', 'success'),
    ])).toEqual(new Map());
  });
});
