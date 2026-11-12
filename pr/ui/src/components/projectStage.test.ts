import { describe, expect, it } from 'vitest';
import type { ProjectFileStatus } from './ProjectStatusView';
import { CONFLICT_REASON, stageBlockReason, stageRequestPaths, stageStateOf, stagedCommitPaths } from './projectStage';

const file = (path: string, state: ProjectFileStatus['state'], extra: Partial<ProjectFileStatus> = {}): ProjectFileStatus => ({
  path,
  state,
  adds: 1,
  dels: 0,
  testStatus: { passed: 0, failed: 0, skipped: 0 },
  lintStatus: { errors: 0, warnings: 0, infos: 0 },
  resultsStale: false,
  ...extra,
});

describe('stageStateOf', () => {
  it.each([
    { name: 'every file staged', files: [file('a', 'staged'), file('b', 'staged')], expected: 'all' },
    { name: 'a partly staged file', files: [file('a', 'both')], expected: 'partial' },
    { name: 'a mix of staged and unstaged', files: [file('a', 'staged'), file('b', 'unstaged')], expected: 'partial' },
    { name: 'nothing staged', files: [file('a', 'unstaged'), file('b', 'untracked'), file('c', 'conflict')], expected: 'none' },
    { name: 'no files', files: [], expected: 'none' },
  ])('is $expected for $name', ({ files, expected }) => {
    expect(stageStateOf(files)).toBe(expected);
  });
});

describe('stageBlockReason', () => {
  it('blocks conflicts, queued files and unstaged linked dependencies, but not unstaging', () => {
    const linked = file('go.mod', 'unstaged', { linkedDependencies: [{ name: 'x', target: '../x' }] });
    expect(stageBlockReason({ file: file('c.go', 'conflict'), lockedGroup: undefined })).toBe(CONFLICT_REASON);
    expect(stageBlockReason({ file: file('q.go', 'unstaged'), lockedGroup: 2 })).toBe('Queued in commit group 2');
    expect(stageBlockReason({ file: linked, lockedGroup: undefined })).toContain('linked outside this repository');
    expect(stageBlockReason({ file: { ...linked, state: 'staged' }, lockedGroup: undefined })).toBeNull();
    expect(stageBlockReason({ file: file('ok.go', 'unstaged'), lockedGroup: undefined })).toBeNull();
  });
});

describe('stageRequestPaths', () => {
  const files = [file('a', 'unstaged'), file('b', 'both'), file('c', 'staged'), file('d', 'conflict'), file('e', 'untracked')];
  const locked = new Map([['e', 1]]);

  it('stages every unblocked file that is not already fully staged', () => {
    expect(stageRequestPaths({ files, locked, unstage: false })).toEqual(['a', 'b']);
  });

  it('unstages every staged or partly staged unlocked file', () => {
    expect(stageRequestPaths({ files, locked, unstage: true })).toEqual(['b', 'c']);
  });
});

describe('stagedCommitPaths', () => {
  it('lists staged and partly staged files that are not queued', () => {
    const files = [file('z', 'staged'), file('a', 'both'), file('m', 'unstaged'), file('q', 'staged')];
    expect(stagedCommitPaths(files, new Map([['q', 1]]))).toEqual(['a', 'z']);
  });
});
