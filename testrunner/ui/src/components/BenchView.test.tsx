import { afterEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { BenchView } from './BenchView';
import type { BenchComparison } from '../types';

afterEach(cleanup);

const comparison: BenchComparison = {
  base_label: 'main', head_label: 'feature', threshold: 5, geomean_delta: 7.5, has_regression: true,
  deltas: [
    { name: 'BenchmarkParse', base_mean: 1_000_000, head_mean: 1_200_000, delta_pct: 20, significant: true, p_value: 0.012, samples: 10 },
    { name: 'BenchmarkRender', base_mean: 2e9, head_mean: 1.5e9, delta_pct: -25, significant: true },
    { name: 'BenchmarkNoise', base_mean: 500, head_mean: 505, delta_pct: 1, significant: false },
    { name: 'BenchmarkNew', base_mean: 0, head_mean: 0, delta_pct: 0, only_in: 'head' },
  ],
};

describe('BenchView', () => {
  it('points at the command that loads a comparison when none is loaded', () => {
    render(<BenchView bench={undefined} />);
    expect(screen.getByText('gavel bench compare --ui')).toBeTruthy();
  });

  it('summarises regressions, improvements and the geomean against the threshold', () => {
    render(<BenchView bench={comparison} />);

    expect(screen.getByText('main')).toBeTruthy();
    expect(screen.getByText('feature')).toBeTruthy();
    expect(screen.getByText('threshold ±5.0%')).toBeTruthy();
    expect(screen.getByText('1 regression')).toBeTruthy();
    expect(screen.getByText('1 improvement')).toBeTruthy();
    expect(screen.getByText('1 unchanged')).toBeTruthy();
    expect(screen.getByText('geomean: +7.50%')).toBeTruthy();
  });

  it('shows each benchmark\'s means, signed delta and significance, and marks one present in only one run', () => {
    render(<BenchView bench={comparison} />);

    expect(screen.getByText('BenchmarkParse')).toBeTruthy();
    expect(screen.getByText('1.00ms')).toBeTruthy();
    expect(screen.getByText('1.20ms')).toBeTruthy();
    expect(screen.getByText('+20.00%')).toBeTruthy();
    expect(screen.getByText('-25.00%')).toBeTruthy();
    expect(screen.getByText(/0\.012/)).toBeTruthy();
    expect(screen.getByText('only in head')).toBeTruthy();
  });
});
