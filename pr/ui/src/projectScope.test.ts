import { describe, expect, it } from 'vitest';
import type { PRItem, Project } from './types';
import { applyProjectScope } from './projectScope';

const projects: Project[] = [
  { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] },
  { name: 'local-only', dir: '/work/local-only', repos: [] },
];

const prs = [
  { repo: 'acme/gavel', number: 1 },
  { repo: 'acme/clicky', number: 2 },
] as PRItem[];

describe('applyProjectScope', () => {
  it('returns only the selected project and its pull requests', () => {
    expect(applyProjectScope(projects, prs, 'gavel')).toEqual({
      projects: [projects[0]],
      prs: [prs[0]],
    });
  });

  it('keeps a local-only project while returning no pull requests', () => {
    expect(applyProjectScope(projects, prs, 'local-only')).toEqual({
      projects: [projects[1]],
      prs: [],
    });
  });

  it('returns the original collections when no project is selected', () => {
    expect(applyProjectScope(projects, prs, '')).toEqual({ projects, prs });
  });
});
