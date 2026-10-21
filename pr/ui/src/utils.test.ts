import { describe, expect, it } from 'vitest';
import { aggregateGavelShards, failedTestsLabel } from './utils';
import type { GavelResultsSummary } from './types';

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
