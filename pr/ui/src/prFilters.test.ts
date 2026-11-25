import { describe, expect, it } from 'vitest';
import { emptyFilters } from './components/FilterBar';
import { buildRoute, parseRoute } from './routes';
import { filtersFromStored, loadUIState, saveUIState } from './storage';
import type { PRItem } from './types';
import { filterPRs } from './utils';

const now = Date.parse('2026-11-01T12:00:00Z');
const day = 86_400_000;
const pr: PRItem = {
  number: 1, title: 'Fix a failing check', author: 'developer', repo: 'acme/app',
  source: 'fix', target: 'main', state: 'OPEN', isDraft: false,
  url: 'https://example.com/pr/1', updatedAt: '',
};
const prs = [0, 1, 7, 30].map((days, index) => ({
  ...pr, number: index + 1, updatedAt: new Date(now - days * day).toISOString(),
}));

describe('PR age filters', () => {
  it.each([
    ['day', [1]],
    ['week', [2]],
    ['month', [3]],
    ['older', [4]],
    ['unknown', [5]],
  ])('includes only the %s age bucket at its boundary', (bucket, expected) => {
    const filters = { ...emptyFilters(), age: { [bucket]: 'include' as const } };
    expect(filterPRs([...prs, { ...pr, number: 5 }], { filters, viewer: '', now }).map(item => item.number)).toEqual(expected);
  });

  it('combines age exclusions with existing repository filters', () => {
    const filters = { ...emptyFilters(), age: { day: 'exclude' as const }, repos: { [pr.repo]: 'include' as const } };
    expect(filterPRs([...prs, { ...prs[1]!, number: 5, repo: 'acme/other' }], { filters, viewer: '', now }).map(item => item.number)).toEqual([2, 3, 4]);
  });

  it('treats missing, invalid, and Go zero timestamps as unknown', () => {
    const undated = ['', 'invalid', '0001-01-01T00:00:00Z'].map((updatedAt, index) => ({ ...pr, number: index + 1, updatedAt }));
    const filters = { ...emptyFilters(), age: { unknown: 'include' as const } };
    expect(filterPRs(undated, { filters, viewer: '', now }).map(item => item.number)).toEqual([1, 2, 3]);
  });

  it('round-trips included and excluded ages with other URL facets', () => {
    const route = parseRoute(new URL('http://localhost/prs?state=open&age=day%2C-older') as unknown as Location);
    expect(route.filters.age).toEqual({ day: 'include', older: 'exclude' });
    expect(buildRoute(route)).toBe('/prs?state=open&age=day%2C-older');
  });

  it('persists and restores age alongside the other filters', () => {
    const filters = { ...emptyFilters(), age: { week: 'include' as const } };
    saveUIState({ repos: [] }, filters);
    expect(filtersFromStored(loadUIState().filters)).toEqual(filters);
  });
});
