import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { PRRow } from './PRRow';
import type { GavelResultsSummary, PRItem } from '../types';

const pr: PRItem = {
  number: 7,
  title: 'Test PR',
  author: 'octocat',
  repo: 'acme/widget',
  source: 'feature',
  target: 'main',
  state: 'OPEN',
  isDraft: false,
  url: 'https://github.com/acme/widget/pull/7',
  updatedAt: '2026-01-01T00:00:00Z',
};

function makeGavel(overrides: Partial<GavelResultsSummary> = {}): GavelResultsSummary {
  return {
    artifactId: 1,
    artifactUrl: '',
    testsPassed: 9,
    testsFailed: 3,
    testsSkipped: 0,
    testsTotal: 12,
    lintViolations: 0,
    lintLinters: 0,
    hasBench: false,
    ...overrides,
  };
}

describe('PRRow gavel failed badge', () => {
  it.each([
    ['timed-out tests annotate the failed count', { testsTimedOut: 1 }, '3 tests failed (1 timed out)'],
    ['no timeouts leaves the failed count plain', {}, '3 tests failed'],
  ])('%s', (_name, overrides, title) => {
    render(<PRRow pr={pr} selected={false} gavelResults={makeGavel(overrides)} onClick={() => {}} />);

    expect(screen.getAllByTitle(title).length).toBeGreaterThan(0);
  });
});
