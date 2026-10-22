import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { TestRunnerContext } from '@flanksource/clicky-ui/data';
import type { ProcessTraceNode, ProcessTraceState, TracesStatus } from '@flanksource/gavel/testrunner';
import { runRegistry, useRunProcesses } from './runTraces';

const traceCalls = vi.hoisted(() => ({ options: [] as unknown[], state: { current: null as unknown } }));

vi.mock('@flanksource/gavel/testrunner', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/gavel/testrunner')>()),
  useProcessTrace: (options: unknown) => {
    traceCalls.options.push(options);
    return traceCalls.state.current;
  },
}));

const runner = {} as TestRunnerContext;
const base = '/api/project-runs/run-9';
const reported: TracesStatus = {
  process_run: 'run-1',
  process: {
    tree: { kind: 'process_tree', stream: 'tree' },
    samples: { kind: 'process_samples', stream: 'samples' },
    stacks: { kind: 'stacks', stream: 'stacks' },
  },
  runs: { 'run-1': { mount: '/api/traces/k/query/run-1' } },
} as TracesStatus;
const window = { run_id: 'run-1', kind: 'terminal', stream: 's1', count: 2 };

const rootNode = { key: '100@1', pid: 100, name: 'gavel', isRoot: true, status: 'sleeping' } as ProcessTraceNode;
const childNode = { key: '101@2', parentKey: '100@1', pid: 101, name: 'sleep', status: 'sleeping' } as ProcessTraceNode;

function traceState(nodes: ProcessTraceNode[]): ProcessTraceState {
  return {
    nodes, samples: [], stacks: [], goprof: [], loading: false, canCollect: false,
    collecting: false, collectStack: async () => {}, refreshing: false, refreshFacets: async () => {},
  };
}

beforeEach(() => {
  traceCalls.options.length = 0;
  traceCalls.state.current = traceState([]);
});

describe('useRunProcesses', () => {
  it('reads the process trace from the traces as the server reported them, scoped by base, and exposes them rebased once for the adapters', () => {
    const { result } = renderHook(() => useRunProcesses({ traces: reported, base, done: true, active: false }));

    expect(traceCalls.options.at(-1)).toMatchObject({ traces: reported, base, done: true });
    expect(result.current.traces?.runs['run-1']?.mount).toBe(`${base}/api/traces/k/query/run-1`);
  });

  it('selects the root process once the Processes view is active and its tree has arrived', () => {
    traceCalls.state.current = traceState([childNode, rootNode]);
    const { result, rerender } = renderHook(({ active }) => useRunProcesses({ traces: reported, base, done: true, active }), {
      initialProps: { active: false },
    });
    expect(result.current.selectedKey).toBeNull();

    rerender({ active: true });
    expect(result.current.selectedKey).toBe(rootNode.key);
    expect(traceCalls.options.at(-1)).toMatchObject({ selectedKey: rootNode.key });
  });

  it('keeps the process an operator picked instead of reselecting the root', () => {
    traceCalls.state.current = traceState([rootNode, childNode]);
    const { result } = renderHook(() => useRunProcesses({ traces: reported, base, done: true, active: true }));

    act(() => result.current.setSelectedKey(childNode.key));

    expect(result.current.selectedKey).toBe(childNode.key);
  });

  it('is unavailable for a run that kept no process capture', () => {
    const { result } = renderHook(() => useRunProcesses({ traces: undefined, base, done: true, active: true }));
    expect(result.current.available).toBe(false);
  });
});

describe('runRegistry', () => {
  const fixtureNode = { name: 'step', framework: 'fixture', fixture: { pids: [101] }, context: { traces: [window] } };

  it('gives a fixture step its Traces tab against the rebased stores and a Processes tab for the processes it started', () => {
    const registry = runRegistry({
      traces: { runs: { 'run-1': { mount: `${base}/api/traces/k/query/run-1` } } } as TracesStatus,
      processes: [rootNode, childNode],
      onOpenProcess: vi.fn(),
    });

    expect(registry.resolveTabs(fixtureNode as never, runner).map(tab => tab.id)).toEqual(['traces', 'processes']);
  });

  it('offers no Processes tab for a step whose pids are not in the capture, and no Traces tab without armed windows', () => {
    const registry = runRegistry({ traces: undefined, processes: [rootNode], onOpenProcess: vi.fn() });
    const bare = { name: 'step', framework: 'fixture', fixture: { pids: [999] }, context: {} };

    expect(registry.resolveTabs(bare as never, runner)).toEqual([]);
  });

  it('renders go test nodes through gavel detail rather than the verification body', () => {
    const registry = runRegistry({ traces: undefined, processes: [], onOpenProcess: vi.fn() });

    expect(registry.resolveDetail({ name: 't', framework: 'go test' } as never, runner)?.id).toBe('gavel-detail');
    expect(registry.resolveDetail(fixtureNode as never, runner)?.id).not.toBe('gavel-detail');
  });
});
