import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { DetailPanel } from './DetailPanel';
import type { Test } from '../types';

const editableTest: Test = {
  name: 'works',
  framework: 'vitest',
  file: 'sum.test.ts',
  line: 3,
  failed: true,
};

describe('DetailPanel test edit actions', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('hides edit actions when source edits are unsupported', () => {
    const onTestEdit = vi.fn();
    render(<DetailPanel test={editableTest} testEditSupported={false} onTestEdit={onTestEdit} />);

    expect(screen.queryByText('Skip Test')).toBeNull();
    expect(screen.queryByText('Delete File')).toBeNull();
  });

  it('does not call onTestEdit when confirmation is cancelled', () => {
    const onTestEdit = vi.fn();
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    render(<DetailPanel test={editableTest} testEditSupported onTestEdit={onTestEdit} />);

    fireEvent.click(screen.getByText('Skip Test'));

    expect(window.confirm).toHaveBeenCalledWith('Skip test works?');
    expect(onTestEdit).not.toHaveBeenCalled();
  });

  it('calls onTestEdit with the confirmed action and scope', () => {
    const onTestEdit = vi.fn();
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    render(<DetailPanel test={editableTest} testEditSupported onTestEdit={onTestEdit} />);

    fireEvent.click(screen.getByText('Delete File'));

    expect(window.confirm).toHaveBeenCalledWith('Delete file sum.test.ts?');
    expect(onTestEdit).toHaveBeenCalledWith(editableTest, 'delete', 'file');
  });
});

describe('DetailPanel fixture CEL trace', () => {
  it('renders the native trace instead of the plain expression', () => {
    const trace = 'cel: actual == expected\n     │         │\n     │         └─ int(2)\n     └─ int(1)';
    render(<DetailPanel test={{
      name: 'CEL assertion',
      framework: 'fixture',
      failed: true,
      context: {
        cel_expression: 'actual == expected',
        cel_trace: trace,
      },
    }} />);

    expect(screen.getByText((_, element) => element?.textContent === trace)).toBeTruthy();
    expect(screen.queryByText('actual == expected')).toBeNull();
  });
});

describe('DetailPanel fixture resources', () => {
	it('shows SQL text and bound parameters for a fixture test', () => {
		render(<DetailPanel test={{ name: 'SQL fixture', framework: 'fixture', passed: true,
			sql_profile: { path: 'sqlprofile.jsonl', query_count: 1, slow_query_count: 0, total_duration_ms: 10, max_query_ms: 10,
				statements: [{ sql: 'SELECT * FROM users WHERE id = ?', params: ['42'], duration_ms: 10, rows: 1, slow: false, error: false }] },
		}} />);
		fireEvent.click(screen.getByText('1 query'));
		expect(screen.getByText('SELECT * FROM users WHERE id = ?')).toBeTruthy();
		expect(screen.getByText('42')).toBeTruthy();
	});
	it('distinguishes an absent Go profile from a captured download', () => {
		render(<DetailPanel test={{ name: 'profiled fixture', framework: 'fixture', passed: true,
			fixture_profile: { scope: 'process_tree', sample_count: 2, peak_cpu_percent: 2, peak_memory_percent: 1, peak_rss_bytes: 1024,
				disk_io: { disk_read_bytes: 1024, disk_write_bytes: 2048, sample_count: 2 } },
			go_profiles: [{ name: 'cpu', id: 'run-a/cpu.pprof', status: 'captured', bytes: 4096 }, { name: 'heap', status: 'not_emitted' }],
		}} />);
		expect(screen.getByText('Observed disk read')).toBeTruthy();
		expect(screen.getByText('Not emitted')).toBeTruthy();
		expect(screen.getByRole('link', { name: /Download/ }).getAttribute('href')).toBe('/api/tests/fixture-profile?id=run-a%2Fcpu.pprof');
	});

  it('shows the selected fixture process tree peaks and sample coverage', () => {
    render(<DetailPanel test={{
      name: 'single star function', framework: 'fixture', passed: true,
      fixture_profile: { scope: 'process_tree', pid: 11232, sample_count: 4, peak_cpu_percent: 104.4, peak_memory_percent: 0.3, peak_rss_bytes: 119914496 },
    }} />);

    expect(screen.getByText('Resources')).toBeTruthy();
    expect(screen.getByText('104.4%')).toBeTruthy();
    expect(screen.getByText('114 MB')).toBeTruthy();
    expect(screen.getByText('0.3%')).toBeTruthy();
    expect(screen.getByText(/4 samples/)).toBeTruthy();
    expect(screen.getByText(/PID 11232/)).toBeTruthy();
  });
});
