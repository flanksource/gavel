import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { TestRunner, emptyTestFilters, type Test as ClickyTest } from '@flanksource/clicky-ui/data';
import { gavelRegistry, type NodeActionsConfig } from '.';
import type { ProcessTraceNode, Test, TracesStatus } from '../types';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const idleActions: NodeActionsConfig = {
  nodeRouteState: () => undefined,
  failingOnlyRouteState: () => undefined,
  testEditSupported: false,
  onTestEdit: async () => {},
  testEditBusy: false,
};

interface HarnessOptions {
  processes?: ProcessTraceNode[];
  onOpenProcess?: (key: string) => void;
  actions?: Partial<NodeActionsConfig>;
  traces?: TracesStatus;
}

// renderNode shows one node selected in the real test runner, with gavel's
// adapters in front of clicky's verification adapters, as the host wires them.
function renderNode(node: Test, options: HarnessOptions = {}) {
  const registry = gavelRegistry({
    traces: options.traces,
    processes: options.processes ?? [],
    onOpenProcess: options.onOpenProcess ?? (() => {}),
    actions: { ...idleActions, ...options.actions },
  });
  return render(
    <TestRunner
      tests={[node as ClickyTest]}
      selected={node as ClickyTest}
      filters={emptyTestFilters()}
      expandAll={null}
      done
      hideHeader
      adapters={registry}
      onSelect={() => {}}
      onFiltersChange={() => {}}
      onExpandAllChange={() => {}}
    />,
  );
}

const vitestNode: Test = { name: 'works', framework: 'vitest', file: 'sum.test.ts', line: 3, failed: true };

describe('gavel detail body', () => {
  it('shows cached status, inherited labels, retry count, and Ginkgo report entries', () => {
    renderNode({
      name: 'reports metadata',
      framework: 'ginkgo',
      passed: true,
      cached: true,
      context: {
        labels: ['inherited', 'leaf'],
        num_attempts: 2,
        report_entries: [{ name: 'resource', value: { kind: 'Deployment' } }],
      },
    });

    expect(screen.getByText('cached')).toBeTruthy();
    expect(screen.getByText('inherited')).toBeTruthy();
    expect(screen.getByText('leaf')).toBeTruthy();
    expect(screen.getByText('Ginkgo attempts')).toBeTruthy();
    expect(screen.getByText('resource')).toBeTruthy();
    expect(screen.getByText(/Deployment/)).toBeTruthy();
  });

  it('shows the Go subtest parent and import path, the package and the command', () => {
    renderNode({
      name: 'TestRegistry/DetectsGoTest',
      framework: 'go test',
      failed: true,
      package_path: './testrunner',
      command: 'go test ./testrunner -run TestRegistry',
      message: 'runner not found for framework: gotest',
      context: { parent_test: 'TestRegistry', import_path: 'github.com/flanksource/gavel/testrunner' },
    });

    expect(screen.getByText('Parent Test')).toBeTruthy();
    expect(screen.getByText('TestRegistry')).toBeTruthy();
    expect(screen.getByText('github.com/flanksource/gavel/testrunner')).toBeTruthy();
    expect(screen.getByText('./testrunner')).toBeTruthy();
    expect(screen.getByText(/go test \.\/testrunner -run TestRegistry/)).toBeTruthy();
    expect(screen.getAllByText('runner not found for framework: gotest').length).toBeGreaterThan(0);
  });

  it('hides the Ginkgo failure location the structured failure already shows', () => {
    renderNode({
      name: 'handles malformed input',
      framework: 'ginkgo',
      failed: true,
      suite: ['Parser', 'Parse'],
      failure_detail: { kind: 'gomega', summary: 'Expected nil', location: 'parser_test.go:78' },
      context: { suite_description: 'Parsers Suite', suite_path: './testrunner/parsers', failure_location: 'parser_test.go:78' },
    });

    expect(screen.getByText('Parser > Parse')).toBeTruthy();
    expect(screen.getByText('Parsers Suite')).toBeTruthy();
    expect(screen.queryByText('Failure Location')).toBeNull();
  });
});

describe('source actions', () => {
  it('offers no source edits when the server cannot edit sources', () => {
    renderNode(vitestNode);
    expect(screen.queryByText('Source')).toBeNull();
  });

  it('does not edit when the confirmation is cancelled', async () => {
    const onTestEdit = vi.fn(async () => {});
    renderNode(vitestNode, { actions: { testEditSupported: true, onTestEdit } });

    fireEvent.click(screen.getByText('Source'));
    fireEvent.click(await screen.findByText('Skip Test'));
    expect(await screen.findByText('Skip test works?')).toBeTruthy();
    fireEvent.click(screen.getByText('Cancel'));

    expect(onTestEdit).not.toHaveBeenCalled();
  });

  it('edits with the confirmed action and scope', async () => {
    const onTestEdit = vi.fn(async () => {});
    renderNode(vitestNode, { actions: { testEditSupported: true, onTestEdit } });

    fireEvent.click(screen.getByText('Source'));
    fireEvent.click(await screen.findByText('Delete File'));
    expect(await screen.findByText('Delete file sum.test.ts?')).toBeTruthy();
    fireEvent.click(within(screen.getByRole('dialog')).getByText('Delete'));

    expect(onTestEdit).toHaveBeenCalledWith(vitestNode, 'delete', 'file');
  });

  it('offers no source edits for a framework the server cannot edit', () => {
    renderNode({ name: 'fixture step', framework: 'fixture', file: 'a.fixture.md', failed: true }, { actions: { testEditSupported: true } });
    expect(screen.queryByText('Source')).toBeNull();
  });
});

describe('fixture step tabs', () => {
  const traceNode = (key: string, pid: number, command: string): ProcessTraceNode => ({
    key, pid, ppid: 1, name: command.split(' ')[0]!, command, status: 'exited',
    startedAt: '2026-01-01T00:00:00Z', exitedAt: '2026-01-01T00:00:03Z', lastSeenAt: '2026-01-01T00:00:03Z',
    cpuPercent: 0, rssBytes: 0, peakCpuPercent: 12.5, peakRssBytes: 2 * 1024 * 1024,
  });
  const step: Test = { name: 'go test', framework: 'fixture', passed: true, fixture: { key: 'k', kind: 'exec', state: 'passed', pids: [4243, 9999] } };

  it('lists the trace nodes for the step pids and opens one on the Processes tab, skipping pids without a node', async () => {
    const onOpenProcess = vi.fn();
    renderNode(step, { processes: [traceNode('4243@2', 4243, 'go test ./pkg'), traceNode('1@1', 1, 'init')], onOpenProcess });

    fireEvent.click(screen.getByRole('tab', { name: /Processes/ }));
    const row = await screen.findByText('go test ./pkg', { selector: '[data-process-command]' });
    expect(screen.queryByText('init')).toBeNull();
    fireEvent.click(row);

    expect(onOpenProcess).toHaveBeenCalledWith('4243@2');
  });

  it('has no Processes tab when the trace has no node for the step pids', () => {
    renderNode(step, { processes: [traceNode('1@1', 1, 'init')] });
    expect(screen.queryByRole('tab', { name: /Processes/ })).toBeNull();
  });

  it('shows resource peaks on the Resources tab, with no Go profile downloads there', () => {
    renderNode({
      name: 'profiled fixture', framework: 'fixture', passed: true,
      fixture_profile: { scope: 'process_tree', pid: 11232, sample_count: 4, peak_cpu_percent: 104.4, peak_memory_percent: 0.3, peak_rss_bytes: 119914496,
        disk_io: { disk_read_bytes: 1024, disk_write_bytes: 2048, sample_count: 4 } },
    });

    fireEvent.click(screen.getByRole('tab', { name: /Resources/ }));

    expect(screen.getByText('104.4%')).toBeTruthy();
    expect(screen.getByText('114 MB')).toBeTruthy();
    expect(screen.getByText('Observed disk read')).toBeTruthy();
    expect(screen.getByText(/4 samples/)).toBeTruthy();
    expect(screen.getByText(/PID 11232/)).toBeTruthy();
    expect(screen.queryByText('Go profiles')).toBeNull();
  });

  it('shows the goprof window of a fixture step on its Traces tab', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify([
      { name: 'goprof', title: 'Go profiles', exportFormats: [{ name: 'pprof', ext: 'pb.gz', mime: 'application/gzip' }] },
    ]))));
    renderNode({
      name: 'profiled fixture', framework: 'fixture', passed: true,
      context: { traces: [{ run_id: 'run-1', kind: 'goprof', stream: 'g1', count: 3, summary: { count: 3, profiles: 'cpu' } }] },
    }, { traces: { runs: { 'run-1': { mount: '/query/run-1' } } } });

    fireEvent.click(screen.getByRole('tab', { name: /Traces/ }));

    expect(screen.getByText('goprof')).toBeTruthy();
    expect(screen.getByText('3 records')).toBeTruthy();
    expect(screen.getByText('Open in query').getAttribute('href')).toBe('/query/run-1/traces/goprof?stream=g1');
    await waitFor(() => expect(screen.getByText('PPROF').getAttribute('href')).toBe('/query/run-1/api/v1/traces/goprof/export?stream=g1&format=pprof'));
  });

  it('shows SQL text and bound parameters for a fixture test', async () => {
    renderNode({
      name: 'SQL fixture', framework: 'fixture', passed: true,
      sql_profile: { path: 'sqlprofile.jsonl', query_count: 1, slow_query_count: 0, total_duration_ms: 10, max_query_ms: 10,
        statements: [{ sql: 'SELECT * FROM users WHERE id = ?', params: ['42'], duration_ms: 10, rows: 1, slow: false, error: false }] },
    });

    fireEvent.click(screen.getByRole('tab', { name: /Resources/ }));
    fireEvent.click(screen.getByText('1 query'));

    expect(await screen.findByText('SELECT * FROM users WHERE id = ?')).toBeTruthy();
    expect(screen.getByText('42')).toBeTruthy();
    await waitFor(() => expect(screen.getByText('10.0 ms · 1 row')).toBeTruthy());
  });
});
