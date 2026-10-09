import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { GitMetrics } from '../types';
import { formatMs, GitTrackingPanel } from './GitTrackingPanel';

const tracking: GitMetrics = {
  tracking: true,
  fsmonitor: true,
  untrackedCache: false,
  trackedRepos: 25,
  hotWorktrees: 6,
  idleWorktrees: 111,
  scansInFlight: 4,
  rangesComputed: 3,
  refsScans: [{ labels: { result: 'unchanged' }, count: 34, totalMs: 2245, avgMs: 66, p50Ms: 52, p95Ms: 180 }],
  statusScans: [{ labels: { result: 'unchanged', fsmonitor: 'true', untracked_cache: 'false' }, count: 362, totalMs: 19_884, avgMs: 54.9, p50Ms: 41.2, p95Ms: 128 }],
  commands: null,
  queueWait: [],
  reads: null,
};

describe('GitTrackingPanel', () => {
  it('shows the cache settings, tracker size and one row per recorded series', () => {
    render(<GitTrackingPanel metrics={tracking} />);

    expect(screen.getByText('fsmonitor on')).toBeTruthy();
    expect(screen.getByText('untracked cache off')).toBeTruthy();
    expect(screen.getByText('117')).toBeTruthy();
    expect(screen.getByText('6 hot · 111 idle')).toBeTruthy();
    const rows = screen.getAllByRole('row').slice(1).map(row => Array.from(row.querySelectorAll('td'), cell => cell.textContent));
    expect(rows).toEqual([
      ['Status scansfsmonitor=true result=unchanged untracked_cache=false', '362', '55 ms', '41 ms', '128 ms', '19.9 s'],
      ['Ref scansresult=unchanged', '34', '66 ms', '52 ms', '180 ms', '2.2 s'],
    ]);
  });

  it('says so when the process tracks no git state', () => {
    render(<GitTrackingPanel metrics={{ ...tracking, tracking: false }} />);

    expect(screen.getByText(/does not track git state/)).toBeTruthy();
    expect(screen.queryByRole('table')).toBeNull();
  });

  it.each([
    [0.4, '0.4 ms'],
    [9.96, '10.0 ms'],
    [54.9, '55 ms'],
    [1500, '1.5 s'],
  ])('formats %d ms as %s', (ms, expected) => {
    expect(formatMs(ms)).toBe(expected);
  });
});
