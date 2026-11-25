import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
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

describe('PRRow branch line', () => {
  it.each([
    ['main', 'main', false],
    ['develop', 'develop', false],
    ['main', 'develop', true],
    ['release/1.2', 'main', true],
    ['main', undefined, true],
  ])('shows the %s base when the repo default is %s: %s', (target, repoDefaultBranch, shown) => {
    const row = { ...pr, target, ...(repoDefaultBranch ? { repoDefaultBranch } : {}) };
    const { container } = render(<PRRow pr={row} selected={false} onClick={() => {}} />);

    expect(container.textContent?.includes(`→${target}`)).toBe(shown);
    expect(screen.getByText(pr.source)).toBeTruthy();
  });
});

describe('PRRow author avatar', () => {
  it.each([
    ['the viewer authored it', 'octocat', false],
    ['the viewer login differs only in case', 'OctoCat', false],
    ['someone else authored it', 'hubot', true],
    ['the viewer is unknown', undefined, true],
  ])('shows the author avatar when %s: %s', (_name, viewer, shown) => {
    render(<PRRow pr={pr} viewer={viewer} selected={false} onClick={() => {}} />);

    expect(screen.queryByTitle(`@${pr.author}`) !== null).toBe(shown);
  });
});

describe('PRRow age', () => {
  afterEach(() => { vi.useRealTimers(); });

  it('renders the update age as one compact token with the timestamp as its title', () => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-01-01T00:35:00Z'));

    render(<PRRow pr={pr} selected={false} onClick={() => {}} />);

    const age = screen.getByText('35m');
    expect(age.getAttribute('title')).toBe(pr.updatedAt);
    expect(age.className.split(' ')).toContain('whitespace-nowrap');
  });
});

describe('PRRow mobile status', () => {
  const failing: PRItem = { ...pr, mergeable: 'CONFLICTING', checkStatus: { passed: 2, failed: 1, running: 0, pending: 0 } };

  it('keeps only the status icon on mobile: no coloured edge, check tally or conflict text', () => {
    const { container } = render(<PRRow pr={failing} mobile selected={false} onClick={() => {}} />);

    expect(screen.getByRole('img', { name: '1 checks failed' })).toBeTruthy();
    expect(container.querySelector('.border-red-400')).toBeNull();
    expect(screen.queryByText('✓2 ✗1')).toBeNull();
    expect(screen.queryByText('CONFLICTS')).toBeNull();
  });

  it('keeps the coloured edge, check tally and conflict text on desktop', () => {
    const { container } = render(<PRRow pr={failing} selected={false} onClick={() => {}} />);

    expect(container.querySelector('.border-red-400')).not.toBeNull();
    expect(screen.getByText('✓2 ✗1')).toBeTruthy();
    expect(screen.getByText('CONFLICTS')).toBeTruthy();
  });
});

describe('PRRow title', () => {
  it('renders the PR name at normal weight', () => {
    render(<PRRow pr={pr} selected={false} onClick={() => {}} />);

    expect(screen.getByText(pr.title).className.split(' ')).toEqual(expect.arrayContaining(['font-normal']));
  });
});
