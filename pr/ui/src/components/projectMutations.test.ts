import { describe, expect, it } from 'vitest';
import { projectDiffQueryKey } from './projectMutations';

describe('project mutations', () => {
  it('uses a stable prefix for all cached diffs in one project', () => {
    expect(projectDiffQueryKey('gavel')).toEqual(['projects', 'gavel', 'diff']);
  });
});
