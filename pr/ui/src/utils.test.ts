import { describe, expect, it } from 'vitest';
import { aggregateGavelShards, authorProfileURL, commitWebURL, failedTestsLabel, groupByOrg, prsAwaitingFirstFetch, repoWebURL } from './utils';
import type { GavelResultsSummary, PRItem } from './types';

function makeShard(overrides: Partial<GavelResultsSummary> = {}): GavelResultsSummary {
  return {
    artifactId: 1,
    artifactUrl: '',
    testsPassed: 0,
    testsFailed: 0,
    testsSkipped: 0,
    testsTotal: 0,
    lintViolations: 0,
    lintLinters: 0,
    hasBench: false,
    ...overrides,
  };
}

describe('aggregateGavelShards testsTimedOut', () => {
  it('sums timed-out counts across shards, treating an absent count as zero', () => {
    const agg = aggregateGavelShards([
      makeShard({ testsFailed: 3, testsTimedOut: 1 }),
      makeShard({ testsFailed: 2 }),
      makeShard({ testsFailed: 4, testsTimedOut: 2 }),
    ]);

    expect(agg).toMatchObject({ testsFailed: 9, testsTimedOut: 3 });
  });

  it('keeps timed-out a subset of failed in the aggregate', () => {
    const agg = aggregateGavelShards([
      makeShard({ testsFailed: 1, testsTimedOut: 1 }),
      makeShard({ testsFailed: 1, testsTimedOut: 1 }),
    ])!;

    expect(agg.testsTimedOut).toBeLessThanOrEqual(agg.testsFailed);
  });
});

describe('failedTestsLabel', () => {
  it.each([
    ['no timeouts', { testsFailed: 3 }, '3 failed'],
    ['zero timeouts', { testsFailed: 3, testsTimedOut: 0 }, '3 failed'],
    ['some timeouts', { testsFailed: 3, testsTimedOut: 1 }, '3 failed (1 timed out)'],
    ['all timed out', { testsFailed: 2, testsTimedOut: 2 }, '2 failed (2 timed out)'],
  ])('renders %s', (_name, counts, expected) => {
    expect(failedTestsLabel(counts)).toBe(expected);
  });
});

describe('provider links', () => {
  const base = { title: 'T', author: 'Ada Lovelace', source: 'a', target: 'main', state: 'OPEN', isDraft: false, updatedAt: '2026-11-02T10:00:00Z' };
  const github: PRItem = { ...base, number: 7, repo: 'acme/api', url: 'https://github.com/acme/api/pull/7' };
  const azure: PRItem = { ...base, number: 9, provider: 'azure', repo: 'acme/Platform/rules', url: 'https://dev.azure.com/acme/Platform/_git/rules/pullrequest/9' };
  const gitlab: PRItem = { ...base, number: 5, provider: 'gitlab', repo: 'acme/platform/api', url: 'https://gitlab.example.com/acme/platform/api/-/merge_requests/5' };

  it('links a GitHub PR to its repository, owner and author on GitHub', () => {
    const [org] = groupByOrg([github]);
    expect(org.orgUrl).toBe('https://github.com/acme');
    expect(org.repos[0].webUrl).toBe('https://github.com/acme/api');
    expect(org.repos[0].host).toBe('GitHub');
    expect(authorProfileURL(github)).toBe('https://github.com/Ada Lovelace');
    expect(authorProfileURL(github, 'grace'), 'a commenter on the PR').toBe('https://github.com/grace');
    expect(repoWebURL(github)).toBe('https://github.com/acme/api');
  });

  it('links an Azure DevOps PR to its repository, and to no GitHub owner or profile', () => {
    const [org] = groupByOrg([azure]);
    expect(org.org).toBe('acme');
    expect(org.orgUrl).toBeUndefined();
    expect(org.repos[0].repoShort).toBe('Platform/rules');
    expect(org.repos[0].webUrl).toBe('https://dev.azure.com/acme/Platform/_git/rules');
    expect(org.repos[0].host).toBe('Azure DevOps');
    expect(authorProfileURL(azure)).toBeUndefined();
    expect(authorProfileURL(azure, 'grace')).toBeUndefined();
    expect(repoWebURL(azure)).toBe('https://dev.azure.com/acme/Platform/_git/rules');
  });

  it('links a GitLab merge request to its repository, and to no GitHub owner or profile', () => {
    const [org] = groupByOrg([gitlab]);
    expect(org.orgUrl).toBeUndefined();
    expect(org.repos[0].webUrl).toBe('https://gitlab.example.com/acme/platform/api');
    expect(org.repos[0].host).toBe('GitLab');
    expect(authorProfileURL(gitlab)).toBeUndefined();
    expect(repoWebURL(gitlab)).toBe('https://gitlab.example.com/acme/platform/api');
  });

  it("links a commit to its repository's page on the PR's host", () => {
    expect(commitWebURL(github, 'abc123')).toBe('https://github.com/acme/api/commit/abc123');
    expect(commitWebURL(azure, 'abc123')).toBe('https://dev.azure.com/acme/Platform/_git/rules/commit/abc123');
    expect(commitWebURL(gitlab, 'abc123')).toBe('https://gitlab.example.com/acme/platform/api/-/commit/abc123');
  });
});

describe('prsAwaitingFirstFetch', () => {
  it('is loading until the server has finished a fetch', () => {
    expect(prsAwaitingFirstFetch({ fetchedAt: '' })).toBe(true);
    expect(prsAwaitingFirstFetch({ fetchedAt: '0001-01-01T00:00:00Z' })).toBe(true);
    expect(prsAwaitingFirstFetch({ fetchedAt: '2026-11-23T09:00:00Z' })).toBe(false);
  });

  it('is not loading once a fetch failed, so the error shows instead', () => {
    expect(prsAwaitingFirstFetch({ fetchedAt: '0001-01-01T00:00:00Z', error: 'GitHub token not configured' })).toBe(false);
  });
});
