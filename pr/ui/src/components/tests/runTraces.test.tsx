import type React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TestRunnerContext } from '@flanksource/clicky-ui/data';
import { tracesAdapter } from './runTraces';

vi.mock('@flanksource/gavel/testrunner', () => ({
  TracesSection: ({ windows, traces }: { windows: Array<{ run_id: string }>; traces?: { runs: Record<string, { mount?: string }> } }) => (
    <output data-testid="traces-section">{windows.map(window => traces?.runs[window.run_id]?.mount).join(',')}</output>
  ),
}));

const runner = {} as TestRunnerContext;
const window = { run_id: 'run-1', kind: 'terminal', stream: 's1', count: 2 };
const traces = { runs: { 'run-1': { mount: '/api/tests/traces/k/query/run-1' } } };

describe('the Traces tab of a run viewed in the dashboard', () => {
  it('claims only the fixture nodes that armed traces', () => {
    const adapter = tracesAdapter(traces);
    expect(adapter.match({ name: 'f', framework: 'fixture', context: { traces: [window] } } as never, runner)).toBe(true);
    expect(adapter.match({ name: 'f', framework: 'fixture', context: {} } as never, runner)).toBe(false);
    expect(adapter.match({ name: 't', framework: 'go test', context: { traces: [window] } } as never, runner)).toBe(false);
  });

  it('renders the node windows against the run stores the dashboard serves', () => {
    const node = { name: 'f', framework: 'fixture', context: { traces: [window] } };
    const tabs = tracesAdapter(traces).detailTabs!({ node: node as never, runner });
    expect(tabs.map(tab => tab.label)).toEqual(['Traces']);
    render(<>{tabs[0].render() as React.ReactNode}</>);
    expect(screen.getByTestId('traces-section').textContent).toBe('/api/tests/traces/k/query/run-1');
  });
});
