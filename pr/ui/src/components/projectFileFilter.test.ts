import { describe, expect, it } from 'vitest';
import { UNSCOPED_SCOPE, filterProjectFiles, projectFileFiltersActive, scopeCounts, type ProjectFileFilters } from './projectFileFilter';

const files = [
  { path: 'cmd/main.go', scopes: ['app'] },
  { path: 'cmd/main_test.go', scopes: ['test'] },
  { path: 'ui/src/App.tsx', scopes: ['app', 'ui'] },
  { path: 'README.md' },
  { path: 'docs/guide.md', previousPath: 'docs/old-guide.md', scopes: [] },
];

const paths = (filters: ProjectFileFilters) => filterProjectFiles(files, filters).map(file => file.path);

describe('filterProjectFiles', () => {
  it('keeps every file when no filter is set', () => {
    expect(paths({ search: '', scopes: {} })).toEqual(files.map(file => file.path));
  });

  it('matches the search case-insensitively against the path and the previous path', () => {
    expect(paths({ search: 'CMD/', scopes: {} })).toEqual(['cmd/main.go', 'cmd/main_test.go']);
    expect(paths({ search: 'old-guide', scopes: {} })).toEqual(['docs/guide.md']);
  });

  it('keeps only files carrying an included scope', () => {
    expect(paths({ search: '', scopes: { app: 'include' } })).toEqual(['cmd/main.go', 'ui/src/App.tsx']);
  });

  it('drops files carrying an excluded scope', () => {
    expect(paths({ search: '', scopes: { app: 'exclude' } })).toEqual(['cmd/main_test.go', 'README.md', 'docs/guide.md']);
  });

  it('treats files without scopes as the unscoped bucket', () => {
    expect(paths({ search: '', scopes: { [UNSCOPED_SCOPE]: 'include' } })).toEqual(['README.md', 'docs/guide.md']);
  });

  it('requires both the search and the scope facet to match', () => {
    expect(paths({ search: 'cmd', scopes: { app: 'include' } })).toEqual(['cmd/main.go']);
  });
});

describe('scopeCounts', () => {
  it('counts files per scope, with unscoped files under the unscoped bucket', () => {
    expect(scopeCounts(files)).toEqual({ app: 2, test: 1, ui: 1, [UNSCOPED_SCOPE]: 2 });
  });
});

describe('projectFileFiltersActive', () => {
  it('is false for a blank search and no scopes, true otherwise', () => {
    expect(projectFileFiltersActive({ search: '  ', scopes: {} })).toBe(false);
    expect(projectFileFiltersActive({ search: 'x', scopes: {} })).toBe(true);
    expect(projectFileFiltersActive({ search: '', scopes: { app: 'exclude' } })).toBe(true);
  });
});
