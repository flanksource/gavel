import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { FixtureBenchView } from './FixtureBenchView';
import type { FixturePerformance, Test } from '../types';

const benchmark: FixturePerformance = { mode: 'benchmark', status: 'passed', duration_ms: 1000 };

function fixture(name: string, key: string, durationMs: number): Test {
  return { name, framework: 'fixture', duration: durationMs * 1e6, passed: true, fixture: { key, kind: 'test', state: 'passed' } };
}

describe('fixture Bench view', () => {
  it('shows SQL timings and failed benchmark limits from test nodes', () => {
    const test: Test = { ...fixture('SQL fixture', 'sql', 50), failed: true, passed: false,
      fixture: { key: 'sql', kind: 'test', state: 'failed', violations: ['SQL total 30.00ms exceeds 25.00ms'] },
      sql_profile: { path: 'sqlprofile.jsonl', query_count: 2, slow_query_count: 1, total_duration_ms: 30, max_query_ms: 20 } };
    render(<FixtureBenchView performance={{ ...benchmark, status: 'failed' }} tests={[test]} sampleIndex={undefined} onSampleChange={vi.fn()} />);
    expect(screen.getByText('SQL fixture')).toBeTruthy();
    expect(screen.getByText('2 queries · 1 slow')).toBeTruthy();
    expect(screen.getByText('SQL total 30.00ms exceeds 25.00ms')).toBeTruthy();
  });

  it('shows SQL text and bound parameters for a fixture statement', () => {
    const test: Test = { ...fixture('SQL fixture', 'sql', 10), sql_profile: {
      path: 'sqlprofile.jsonl', query_count: 1, slow_query_count: 0, total_duration_ms: 10, max_query_ms: 10,
      statements: [{ duration_ms: 10, rows: 1, slow: false, error: false, sql: 'SELECT * FROM users WHERE id = ?', params: ['42'] }],
    } };
    render(<FixtureBenchView performance={benchmark} tests={[test]} sampleIndex={undefined} onSampleChange={vi.fn()} />);
    fireEvent.click(screen.getByText('1 query'));
    expect(screen.getByText('10.0 ms · 1 row')).toBeTruthy();
    expect(screen.getByText('SELECT * FROM users WHERE id = ?')).toBeTruthy();
    expect(screen.getByText('42')).toBeTruthy();
  });

  it('shows setup as a virtual test while the fixture is running', () => {
    const setup: Test = { name: 'Setup', duration: 125_000_000, passed: true, fixture: { key: 'setup', kind: 'setup', state: 'passed' } };
    render(<FixtureBenchView performance={{ ...benchmark, status: 'running' }} tests={[{ name: 'Fixture lifecycle', children: [setup] }]} sampleIndex={undefined} onSampleChange={vi.fn()} />);
    expect(screen.getByText('Setup')).toBeTruthy();
    expect(screen.getByText('125.0 ms')).toBeTruthy();
    expect(screen.getByText('Running')).toBeTruthy();
  });

  it('shows the failure reason alongside partial timings', () => {
    const setup: Test = { name: 'Setup', duration: 125_000_000, passed: true, fixture: { key: 'setup', kind: 'setup', state: 'passed' } };
    render(<FixtureBenchView performance={{ ...benchmark, status: 'errored' }} tests={[setup]} error="sample process tree failed" sampleIndex={undefined} onSampleChange={vi.fn()} />);
    expect(screen.getByRole('alert').textContent).toContain('sample process tree failed');
    expect(screen.getByText('125.0 ms')).toBeTruthy();
  });

  it('shows comparison, sampled peaks, and the selected process tree', () => {
    const performance: FixturePerformance = { ...benchmark, artifact_path: '.gavel/benchmarks/fixtures/run.json',
      comparison: { baseline: 'baseline.json', deltas: [{ key: 'setup', name: 'Setup', kind: 'setup', baseline_ms: 100, current_ms: 125, delta_ms: 25, delta_pct: 25 }], added: ['new'], missing: ['old'] },
      profile: { interval_ms: 500, peak_cpu_percent: 12, peak_memory_percent: 0.5, peak_rss_bytes: 1048576, samples: [
        { sampled_at: '2026-01-01T00:00:01Z', cpu_percent: 12, memory_percent: 0.5, rss_bytes: 1048576, processes: [
          { pid: 100, command: 'gavel', cpuPercent: 4, rssBytes: 524288 },
          { pid: 101, ppid: 100, command: 'sh -c sleep', cpuPercent: 8, rssBytes: 524288 },
        ] },
      ] },
    };
    const setup: Test = { name: 'Setup', duration: 125_000_000, passed: true, fixture: { key: 'setup', kind: 'setup', state: 'passed' } };
    const onSampleChange = vi.fn();
    render(<FixtureBenchView performance={performance} tests={[setup]} sampleIndex={0} onSampleChange={onSampleChange} />);
    expect(screen.getByText('+25.0%')).toBeTruthy();
    expect(screen.getByText('Peak CPU')).toBeTruthy();
    expect(screen.getByText('sh -c sleep')).toBeTruthy();
    expect(screen.getByText('Added: new')).toBeTruthy();
    expect(screen.getByText('Missing: old')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /sample 1/i }));
    expect(onSampleChange).toHaveBeenCalledWith(0);
  });

  it('shows observed disk bytes and a captured Go profile download', () => {
    const test: Test = { ...fixture('Disk fixture', 'disk', 1000),
      fixture_profile: { scope: 'process_tree', sample_count: 2, peak_cpu_percent: 1, peak_memory_percent: 1, peak_rss_bytes: 1024, disk_io: { disk_read_bytes: 1024, disk_write_bytes: 2048, sample_count: 2 } },
      go_profiles: [{ name: 'cpu', id: 'run-a/cpu.pprof', status: 'captured', bytes: 400 }],
    };
    const performance: FixturePerformance = { ...benchmark, comparison: { baseline: 'baseline.json',
      deltas: [{ key: 'disk', name: 'Disk fixture', kind: 'test', baseline_ms: 1000, current_ms: 1000, delta_ms: 0, disk_read_bytes_delta: -1024, disk_write_bytes_delta: 1024 }] } };
    render(<FixtureBenchView performance={performance} tests={[test]} sampleIndex={undefined} onSampleChange={vi.fn()} />);
    expect(screen.getByText('1.0 KiB')).toBeTruthy();
    expect(screen.getByText('2.0 KiB')).toBeTruthy();
    expect(screen.getByText('Disk R -1.0 KiB · W +1.0 KiB')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Download pprof' }).getAttribute('href')).toBe('/api/tests/fixture-profile?id=run-a%2Fcpu.pprof');
  });
});
